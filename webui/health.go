/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
	"unsafe"

	"github.com/amnezia-vpn/amneziawg-windows-client/manager"
	"github.com/amnezia-vpn/amneziawg-windows/v3/conf"
	"golang.org/x/sys/windows"
)

// Connection check: while a tunnel is connected, ping an address that is
// routed through it every 15 seconds. A handshake time alone says little (it
// is normal for it to be two minutes old), but an echo reply proves packets
// really go through the tunnel and back.

const (
	healthInterval      = 15 * time.Second
	healthRetryInterval = 5 * time.Second  // after a failed round
	healthFirstCheck    = 3 * time.Second  // let routes settle after connecting
	healthPingTimeout   = 2 * time.Second  // per echo request
	healthFailsForDown  = 2                // failed rounds in a row
	reconnectAfter      = 60 * time.Second // of lost connection, when enabled
	reconnectMaxTries   = 3                // in a row without the link coming back
)

// healthInfo is what the page shows for a connected tunnel.
type healthInfo struct {
	Status  string `json:"status"`           // "checking", "ok", "down", "reconnecting", "unknown" (nothing to ping)
	RTT     int    `json:"rtt,omitempty"`    // ms, when ok
	Target  string `json:"target,omitempty"` // address that answered / was tried
	Since   int64  `json:"since,omitempty"`  // unix ms: when the link was lost
	Checked int64  `json:"checked"`          // unix ms of the last check
}

type healthMonitor struct {
	stop chan struct{}
}

// syncHealth starts a monitor for every connected tunnel and stops the rest.
// Safe to call from any goroutine, as often as needed.
func (a *app) syncHealth() {
	on := a.settings.get().HealthCheck
	names, err := a.be.TunnelNames()
	if err != nil {
		return
	}
	running := map[string]bool{}
	for _, n := range names {
		if st, err := a.be.State(n); err == nil && st == manager.TunnelStarted {
			running[n] = true
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.monitors == nil {
		a.monitors = map[string]*healthMonitor{}
		a.health = map[string]healthInfo{}
		a.reconnects = map[string]int{}
	}
	for n, m := range a.monitors {
		if !on || !running[n] {
			close(m.stop)
			delete(a.monitors, n)
			delete(a.health, n)
			a.emit("health", map[string]any{"name": n, "health": nil})
		}
	}
	if !on {
		return
	}
	for n := range running {
		if _, ok := a.monitors[n]; ok {
			continue
		}
		m := &healthMonitor{stop: make(chan struct{})}
		a.monitors[n] = m
		go a.runHealth(n, m.stop)
	}
}

func (a *app) healthSnapshot() map[string]healthInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make(map[string]healthInfo, len(a.health))
	for k, v := range a.health {
		out[k] = v
	}
	return out
}

func (a *app) setHealth(name string, h healthInfo, stop <-chan struct{}) bool {
	a.mu.Lock()
	select {
	case <-stop:
		a.mu.Unlock()
		return false // monitor was stopped meanwhile; don't resurrect the entry
	default:
	}
	a.health[name] = h
	a.mu.Unlock()
	a.emit("health", map[string]any{"name": name, "health": h})
	return true
}

func (a *app) runHealth(name string, stop <-chan struct{}) {
	h := healthInfo{Status: "checking", Checked: time.Now().UnixMilli()}
	if !a.setHealth(name, h, stop) {
		return
	}
	fails := 0
	var downSince time.Time
	lastGood := ""
	wait := healthFirstCheck
	for {
		select {
		case <-stop:
			return
		case <-time.After(wait):
		}

		var targets []net.IP
		if a.demo {
			targets = nil
		} else if cfg, err := a.be.StoredConfig(name); err == nil {
			targets = healthTargets(&cfg)
		}
		now := time.Now()
		h = healthInfo{Checked: now.UnixMilli()}

		if a.demo {
			h.Status, h.RTT, h.Target = "ok", 38, "10.8.1.1"
		} else if len(targets) == 0 {
			h.Status = "unknown"
		} else if ip, rtt, ok := probe(targets, lastGood); ok {
			lastGood = ip.String()
			h.Status, h.RTT, h.Target = "ok", int(rtt/time.Millisecond), lastGood
		} else {
			h.Status, h.Target = "fail", targets[0].String()
		}

		switch h.Status {
		case "ok":
			if !downSince.IsZero() {
				log.Printf("[%s] Connection check: link restored after %s", name, now.Sub(downSince).Round(time.Second))
				a.notifyHealth(name, "restored")
			}
			fails, downSince = 0, time.Time{}
			a.mu.Lock()
			delete(a.reconnects, name)
			a.mu.Unlock()
			wait = healthInterval
		case "unknown":
			wait = healthInterval
		default:
			fails++
			wait = healthRetryInterval
			if fails < healthFailsForDown {
				h.Status = "checking"
				break
			}
			if downSince.IsZero() {
				downSince = now
				log.Printf("[%s] Connection check: no reply from %v", name, targets)
				a.notifyHealth(name, "lost")
			}
			h.Status, h.Since = "down", downSince.UnixMilli()
			if a.settings.get().AutoReconnect && now.Sub(downSince) >= reconnectAfter && a.startReconnect(name) {
				h.Status = "reconnecting"
				a.setHealth(name, h, stop)
				return // the tunnel restarts; a new monitor takes over
			}
			wait = healthInterval
		}
		if !a.setHealth(name, h, stop) {
			return
		}
	}
}

// startReconnect restarts the tunnel in the background, unless it was already
// restarted too many times in a row without the link coming back.
func (a *app) startReconnect(name string) bool {
	a.mu.Lock()
	n := a.reconnects[name]
	if n >= reconnectMaxTries {
		a.mu.Unlock()
		return false
	}
	a.reconnects[name] = n + 1
	a.mu.Unlock()
	log.Printf("[%s] Connection check: reconnecting (attempt %d of %d)", name, n+1, reconnectMaxTries)
	a.notifyHealth(name, "reconnecting")
	go func() {
		if err := a.be.Stop(name); err != nil {
			log.Printf("[%s] Reconnect: unable to stop: %v", name, err)
			return
		}
		a.be.WaitForStop(name)
		if err := a.be.Start(name); err != nil {
			log.Printf("[%s] Reconnect: unable to start: %v", name, err)
		}
	}()
	return true
}

func (a *app) notifyHealth(name, what string) {
	s := a.settings.get()
	if !s.Notifications {
		return
	}
	dispatch(func() {
		if a.main.visible() {
			return // the page shows its own toasts
		}
		tr := trText(s.Language)
		a.tray.notify(appName, fmt.Sprintf(tr("health."+what), name), what == "lost")
	})
}

// healthTargets returns the IPv4 addresses to ping that the tunnel routes:
// its DNS servers first (usually the server itself), then public resolvers
// when the tunnel carries all traffic.
func healthTargets(cfg *conf.Config) []net.IP {
	routed := func(ip net.IP) bool {
		for _, p := range cfg.Peers {
			for _, c := range p.AllowedIPs {
				n := c.IPNet()
				if n.Contains(ip) {
					return true
				}
			}
		}
		return false
	}
	var out []net.IP
	seen := map[string]bool{}
	add := func(ip net.IP) {
		ip4 := ip.To4()
		if ip4 == nil || seen[ip4.String()] || !routed(ip4) {
			return
		}
		seen[ip4.String()] = true
		out = append(out, ip4)
	}
	for _, d := range cfg.Interface.DNS {
		add(d)
	}
	add(net.IPv4(1, 1, 1, 1))
	add(net.IPv4(8, 8, 8, 8))
	return out
}

// probe pings the targets, the one that answered last time first, and
// returns the first that replies.
func probe(targets []net.IP, preferred string) (net.IP, time.Duration, bool) {
	ordered := make([]net.IP, 0, len(targets))
	for _, t := range targets {
		if t.String() == preferred {
			ordered = append(ordered, t)
		}
	}
	for _, t := range targets {
		if t.String() != preferred {
			ordered = append(ordered, t)
		}
	}
	for _, t := range ordered {
		if rtt, err := ping4(t, healthPingTimeout); err == nil {
			return t, rtt, true
		}
	}
	return nil, 0, false
}

var (
	modIphlpapi         = windows.NewLazySystemDLL("iphlpapi.dll")
	procIcmpCreateFile  = modIphlpapi.NewProc("IcmpCreateFile")
	procIcmpCloseHandle = modIphlpapi.NewProc("IcmpCloseHandle")
	procIcmpSendEcho    = modIphlpapi.NewProc("IcmpSendEcho")
)

// icmpEchoReply is ICMP_ECHO_REPLY (the native one for this architecture).
type icmpEchoReply struct {
	Address       uint32
	Status        uint32
	RoundTripTime uint32
	DataSize      uint16
	Reserved      uint16
	Data          uintptr
	Options       struct {
		TTL, TOS, Flags, OptionsSize uint8
		OptionsData                  uintptr
	}
}

// ping4 sends one ICMP echo request. It needs no administrator rights.
func ping4(ip net.IP, timeout time.Duration) (time.Duration, error) {
	ip4 := ip.To4()
	if ip4 == nil {
		return 0, fmt.Errorf("not an IPv4 address: %v", ip)
	}
	h, _, err := procIcmpCreateFile.Call()
	if windows.Handle(h) == windows.InvalidHandle {
		return 0, err
	}
	defer procIcmpCloseHandle.Call(h)

	payload := []byte("NetPortal")
	reply := make([]byte, int(unsafe.Sizeof(icmpEchoReply{}))+len(payload)+8+64)
	start := time.Now()
	n, _, err := procIcmpSendEcho.Call(
		h,
		uintptr(binary.LittleEndian.Uint32(ip4)), // IPAddr: network byte order in memory
		uintptr(unsafe.Pointer(&payload[0])),
		uintptr(len(payload)),
		0,
		uintptr(unsafe.Pointer(&reply[0])),
		uintptr(len(reply)),
		uintptr(timeout/time.Millisecond),
	)
	elapsed := time.Since(start)
	if n == 0 {
		return 0, err
	}
	r := (*icmpEchoReply)(unsafe.Pointer(&reply[0]))
	if r.Status != 0 { // IP_SUCCESS
		return 0, fmt.Errorf("ICMP status %d", r.Status)
	}
	rtt := time.Duration(r.RoundTripTime) * time.Millisecond
	if rtt == 0 && elapsed < timeout {
		rtt = elapsed
	}
	return rtt, nil
}

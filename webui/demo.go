/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"errors"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/amnezia-vpn/amneziawg-windows-client/manager"
	"github.com/amnezia-vpn/amneziawg-windows-client/updater"
	"github.com/amnezia-vpn/amneziawg-windows/v3/conf"
)

// demoBackend simulates the manager service in memory. It is only used by
// the /uidemo command line switch to preview the interface; it never touches
// the network or the system's tunnel services.
type demoBackend struct {
	mu      sync.Mutex
	configs map[string]*conf.Config
	states  map[string]manager.TunnelState
	started map[string]time.Time
	sink    backendEvents
}

const demoTemplate = `[Interface]
PrivateKey = %KEY%
Address = 10.8.0.%N%/32, fd42:42:42::%N%/128
DNS = 1.1.1.1, 1.0.0.1
MTU = 1280
Jc = 4
Jmin = 40
Jmax = 70
S1 = 52
S2 = 18
H1 = 1189405542
H2 = 3725167803
H3 = 2416970128
H4 = 905781264

[Peer]
PublicKey = Hk2LpZ8nQw5rT1yUe7oI3aSdF9gJcVbX4mN0tR6uWqE=
PresharedKey = p7XcR2vN9bQ4tL1mZ8kW3yH6sD0fG5jA2eU7iO4rTnE=
AllowedIPs = %ALLOWED%
Endpoint = 203.0.113.%N%:51820
PersistentKeepalive = 25
`

func newDemoBackend() *demoBackend {
	d := &demoBackend{
		configs: map[string]*conf.Config{},
		states:  map[string]manager.TunnelState{},
		started: map[string]time.Time{},
	}
	for i, name := range []string{"ams-01", "fra-home", "office-msk", "hel-backup", "nyc-stream", "tokyo-02"} {
		key, _ := conf.NewPrivateKey()
		allowed := "0.0.0.0/0, ::/0"
		if name == "office-msk" {
			allowed = "10.20.0.0/16, 192.168.50.0/24"
		}
		n := []string{"24", "31", "12", "45", "77", "90"}[i]
		text := strings.NewReplacer("%KEY%", key.String(), "%N%", n, "%ALLOWED%", allowed).Replace(demoTemplate)
		cfg, err := conf.FromWgQuick(text, name)
		if err != nil {
			continue
		}
		d.configs[name] = cfg
		d.states[name] = manager.TunnelStopped
	}
	d.states["ams-01"] = manager.TunnelStarted
	d.started["ams-01"] = time.Now().Add(-42 * time.Minute)
	return d
}

func (d *demoBackend) TunnelNames() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	names := make([]string, 0, len(d.configs))
	for n := range d.configs {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return conf.TunnelNameIsLess(names[i], names[j]) })
	return names, nil
}

func (d *demoBackend) State(name string) (manager.TunnelState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.configs[name]; !ok {
		return manager.TunnelUnknown, errors.New("no such tunnel")
	}
	return d.states[name], nil
}

func (d *demoBackend) StoredConfig(name string) (conf.Config, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	c, ok := d.configs[name]
	if !ok {
		return conf.Config{}, errors.New("no such tunnel")
	}
	return *c, nil
}

func (d *demoBackend) RuntimeConfig(name string) (conf.Config, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	c, ok := d.configs[name]
	if !ok || d.states[name] != manager.TunnelStarted {
		return conf.Config{}, errors.New("tunnel is not running")
	}
	rc := *c
	rc.Peers = append([]conf.Peer(nil), c.Peers...)
	secs := time.Since(d.started[name]).Seconds()
	for i := range rc.Peers {
		// Monotonic counters whose derivative wobbles around ~48/6 Mbit/s.
		w := float64(i + 1)
		rc.Peers[i].RxBytes = conf.Bytes((secs*6e6 + 9e6*math.Sin(secs/4.3) + 4e6*math.Sin(secs/1.6+1) + 2e7) / w)
		rc.Peers[i].TxBytes = conf.Bytes((secs*0.75e6 + 1.2e6*math.Sin(secs/2.7) + 3e6) / w)
		rc.Peers[i].LastHandshakeTime = conf.HandshakeTime(time.Duration(time.Now().Add(-38 * time.Second).UnixNano()))
	}
	return rc, nil
}

func (d *demoBackend) setState(name string, state manager.TunnelState) {
	d.mu.Lock()
	d.states[name] = state
	if state == manager.TunnelStarted {
		d.started[name] = time.Now()
	}
	sink := d.sink
	global := manager.TunnelStopped
	for _, s := range d.states {
		if s == manager.TunnelStarted || s == manager.TunnelStarting {
			global = manager.TunnelStarted
		}
	}
	d.mu.Unlock()
	if sink.TunnelChanged != nil {
		sink.TunnelChanged(name, state, global, nil)
	}
}

func (d *demoBackend) Start(name string) error {
	// Only one tunnel at a time, like the real manager.
	names, _ := d.TunnelNames()
	for _, other := range names {
		if s, _ := d.State(other); other != name && (s == manager.TunnelStarted || s == manager.TunnelStarting) {
			d.setState(other, manager.TunnelStopped)
		}
	}
	d.setState(name, manager.TunnelStarting)
	go func() {
		time.Sleep(1500 * time.Millisecond)
		if name == "hel-backup" {
			d.setState(name, manager.TunnelStopped)
			d.mu.Lock()
			sink := d.sink
			d.mu.Unlock()
			if sink.TunnelChanged != nil {
				sink.TunnelChanged(name, manager.TunnelStopped, manager.TunnelStopped, errors.New("handshake did not complete after 20 attempts"))
			}
			return
		}
		d.setState(name, manager.TunnelStarted)
	}()
	return nil
}

func (d *demoBackend) Stop(name string) error {
	d.setState(name, manager.TunnelStopping)
	go func() {
		time.Sleep(400 * time.Millisecond)
		d.setState(name, manager.TunnelStopped)
	}()
	return nil
}

func (d *demoBackend) Delete(name string) error {
	d.mu.Lock()
	delete(d.configs, name)
	delete(d.states, name)
	sink := d.sink
	d.mu.Unlock()
	if sink.TunnelsChanged != nil {
		sink.TunnelsChanged()
	}
	return nil
}

func (d *demoBackend) WaitForStop(name string) error { return nil }

func (d *demoBackend) Create(cfg *conf.Config) error {
	d.mu.Lock()
	for n := range d.configs {
		if strings.EqualFold(n, cfg.Name) {
			d.mu.Unlock()
			return errors.New("tunnel already exists")
		}
	}
	c := *cfg
	d.configs[cfg.Name] = &c
	d.states[cfg.Name] = manager.TunnelStopped
	sink := d.sink
	d.mu.Unlock()
	if sink.TunnelsChanged != nil {
		sink.TunnelsChanged()
	}
	return nil
}

func (d *demoBackend) GlobalState() (manager.TunnelState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, s := range d.states {
		if s == manager.TunnelStarted || s == manager.TunnelStarting {
			return manager.TunnelStarted, nil
		}
	}
	return manager.TunnelStopped, nil
}

func (d *demoBackend) UpdateState() (manager.UpdateState, error) {
	return manager.UpdateStateFoundUpdate, nil
}

func (d *demoBackend) Update() error {
	go func() {
		const total = 12_500_000
		for done := uint64(0); done <= total; done += 900_000 {
			time.Sleep(250 * time.Millisecond)
			d.mu.Lock()
			sink := d.sink
			d.mu.Unlock()
			if sink.UpdateProgress != nil {
				sink.UpdateProgress(updater.DownloadProgress{Activity: "Downloading update", BytesDownloaded: done, BytesTotal: total})
			}
		}
	}()
	return nil
}

func (d *demoBackend) Quit(stopTunnels bool) error { return nil }

func (d *demoBackend) Subscribe(sink backendEvents) {
	d.mu.Lock()
	d.sink = sink
	d.mu.Unlock()
}

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"

	"github.com/amnezia-vpn/amneziawg-windows-client/manager"
	"github.com/amnezia-vpn/amneziawg-windows-client/ringlogger"
	"github.com/amnezia-vpn/amneziawg-windows-client/version"
	"github.com/amnezia-vpn/amneziawg-windows-client/webui/syntax"
	"github.com/amnezia-vpn/amneziawg-windows/v3/conf"
)

// The page calls Go with chrome.webview.postMessage(JSON) of
// {id, m: method, a: args}; Go answers with np._reply(id, ok, data).

type rpcRequest struct {
	ID     int64           `json:"id"`
	Method string          `json:"m"`
	Args   json.RawMessage `json:"a"`
}

// rpcError carries a machine-readable code the page translates, plus the
// original message as detail.
type rpcError struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

func (e *rpcError) Error() string { return e.Code + ": " + e.Detail }

func codeErr(code string, err error) error {
	d := ""
	if err != nil {
		d = err.Error()
	}
	return &rpcError{Code: code, Detail: d}
}

type handler struct {
	ui    bool // must run on the UI thread (dialogs, window operations)
	admin bool // only for administrators
	fn    func(a *app, w *webWindow, args json.RawMessage) (any, error)
}

var handlers map[string]handler

func init() {
	handlers = map[string]handler{
		"init":          {fn: (*app).rpcInit},
		"list":          {fn: (*app).rpcList},
		"details":       {fn: (*app).rpcDetails},
		"stats":         {fn: (*app).rpcStats},
		"start":         {fn: (*app).rpcStart},
		"stop":          {fn: (*app).rpcStop},
		"conflicts":     {fn: (*app).rpcConflicts},
		"remove":        {admin: true, fn: (*app).rpcRemove},
		"getConfig":     {admin: true, fn: (*app).rpcGetConfig},
		"newConfig":     {admin: true, fn: (*app).rpcNewConfig},
		"analyze":       {fn: (*app).rpcAnalyze},
		"killSwitch":    {fn: (*app).rpcKillSwitch},
		"save":          {admin: true, fn: (*app).rpcSave},
		"importDialog":  {ui: true, admin: true, fn: (*app).rpcImportDialog},
		"importData":    {admin: true, fn: (*app).rpcImportData},
		"importText":    {admin: true, fn: (*app).rpcImportText},
		"importQR":      {admin: true, fn: (*app).rpcImportQR},
		"clipboardText": {fn: (*app).rpcClipboardText},
		"copyText":      {fn: (*app).rpcCopyText},
		"exportAll":     {ui: true, admin: true, fn: (*app).rpcExportAll},
		"exportTunnel":  {ui: true, admin: true, fn: (*app).rpcExportTunnel},
		"log":           {fn: (*app).rpcLog},
		"saveLog":       {ui: true, fn: (*app).rpcSaveLog},
		"setSettings":   {fn: (*app).rpcSetSettings},
		"setAutostart":  {admin: true, fn: (*app).rpcSetAutostart},
		"update":        {admin: true, fn: (*app).rpcUpdate},
		"externalIP":    {fn: (*app).rpcExternalIP},
		"openURL":       {fn: (*app).rpcOpenURL},
		"setChrome":     {ui: true, fn: (*app).rpcSetChrome},
		"showMain":      {ui: true, fn: (*app).rpcShowMain},
		"hidePopup":     {ui: true, fn: (*app).rpcHidePopup},
		"quit":          {ui: true, admin: true, fn: (*app).rpcQuit},
	}
}

func (a *app) handleMessage(w *webWindow, s string) {
	var req rpcRequest
	if err := json.Unmarshal([]byte(s), &req); err != nil {
		return
	}
	h, ok := handlers[req.Method]
	reply := func(data any, err error) {
		var payload []byte
		okFlag := "true"
		if err != nil {
			okFlag = "false"
			var re *rpcError
			if !errors.As(err, &re) {
				re = &rpcError{Code: "error", Detail: err.Error()}
			}
			payload, _ = json.Marshal(re)
		} else {
			payload, _ = json.Marshal(data)
		}
		js := fmt.Sprintf("window.np&&np._reply(%d,%s,%s)", req.ID, okFlag, payload)
		dispatch(func() { w.eval(js) })
	}
	if !ok {
		reply(nil, codeErr("unknownMethod", errors.New(req.Method)))
		return
	}
	if h.admin && !a.isAdmin {
		reply(nil, codeErr("notAdmin", nil))
		return
	}
	call := func() {
		defer func() {
			if r := recover(); r != nil {
				reply(nil, codeErr("error", fmt.Errorf("%v", r)))
			}
		}()
		reply(h.fn(a, w, req.Args))
	}
	if h.ui {
		// Run outside of the WebView2 event handler: modal dialogs spin a
		// nested message loop, which WebView2 does not allow from callbacks.
		dispatch(call)
	} else {
		go call()
	}
}

func decodeArgs(raw json.RawMessage, v any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, v)
}

type nameArg struct {
	Name string `json:"name"`
}

// ---- init / list -----------------------------------------------------------

type tunnelInfo struct {
	Name  string `json:"name"`
	State string `json:"state"`
	Full  bool   `json:"full"`
	Error string `json:"error,omitempty"`
}

func (a *app) rpcInit(w *webWindow, _ json.RawMessage) (any, error) {
	auto, autoErr := a.settings.autostart()
	a.mu.Lock()
	upd := updateStateString(a.updateState)
	a.mu.Unlock()
	tunnels, _ := a.listTunnels()
	return map[string]any{
		"admin":          a.isAdmin,
		"demo":           a.demo,
		"settings":       a.settings.get(),
		"autostart":      auto,
		"autostartKnown": autoErr == nil,
		"update":         upd,
		"tunnels":        tunnels,
		"about": map[string]string{
			"app":    version.Number,
			"wintun": wintunVersion(),
			"go":     strings.TrimPrefix(runtime.Version(), "go"),
			"os":     version.OsName(),
			"arch":   version.Arch(),
		},
	}, nil
}

func wintunVersion() (v string) {
	defer func() {
		if recover() != nil {
			v = "—"
		}
	}()
	return version.WintunVersion()
}

func (a *app) listTunnels() ([]tunnelInfo, error) {
	names, err := a.be.TunnelNames()
	if err != nil {
		return nil, err
	}
	sort.Slice(names, func(i, j int) bool { return conf.TunnelNameIsLess(names[i], names[j]) })
	out := make([]tunnelInfo, 0, len(names))
	for _, n := range names {
		st, _ := a.be.State(n)
		info := tunnelInfo{Name: n, State: stateString(st)}
		if cfg, err := a.be.StoredConfig(n); err == nil {
			info.Full = routesAll(&cfg)
		}
		a.mu.Lock()
		info.Error = a.lastErrors[n]
		a.mu.Unlock()
		out = append(out, info)
	}
	return out, nil
}

func (a *app) rpcList(w *webWindow, _ json.RawMessage) (any, error) {
	return a.listTunnels()
}

// routesAll reports whether the configuration sends all traffic through the
// tunnel (0.0.0.0/0, ::/0, or the two /1 halves).
func routesAll(c *conf.Config) bool {
	for _, p := range c.Peers {
		var lowV4, highV4, lowV6, highV6 bool
		for _, ip := range p.AllowedIPs {
			if ip.Cidr == 0 {
				return true
			}
			if ip.Cidr == 1 {
				if v4 := ip.IP.To4(); v4 != nil {
					if v4[0] == 0 {
						lowV4 = true
					} else if v4[0] == 0x80 {
						highV4 = true
					}
				} else if len(ip.IP) == net.IPv6len {
					if ip.IP[0] == 0 {
						lowV6 = true
					} else if ip.IP[0] == 0x80 {
						highV6 = true
					}
				}
			}
		}
		if (lowV4 && highV4) || (lowV6 && highV6) {
			return true
		}
	}
	return false
}

func killSwitchOn(c *conf.Config) bool {
	if len(c.Peers) != 1 || c.Interface.TableOff {
		return false
	}
	for _, ip := range c.Peers[0].AllowedIPs {
		if ip.Cidr == 0 {
			return true
		}
	}
	return false
}

// ---- details / stats -------------------------------------------------------

type kv struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ifaceView struct {
	PublicKey  string   `json:"publicKey"`
	ListenPort string   `json:"listenPort"`
	MTU        string   `json:"mtu"`
	Addresses  []string `json:"addresses"`
	DNS        []string `json:"dns"`
	TableOff   bool     `json:"tableOff"`
	PreUp      string   `json:"preUp"`
	PostUp     string   `json:"postUp"`
	PreDown    string   `json:"preDown"`
	PostDown   string   `json:"postDown"`
}

type peerView struct {
	PublicKey    string   `json:"publicKey"`
	PresharedKey string   `json:"presharedKey,omitempty"`
	HasPSK       bool     `json:"hasPsk"`
	Endpoint     string   `json:"endpoint"`
	AllowedIPs   []string `json:"allowedIps"`
	Keepalive    string   `json:"keepalive"`
	Rx           uint64   `json:"rx"`
	Tx           uint64   `json:"tx"`
	Handshake    int64    `json:"handshake"` // unix seconds, 0 for never
}

type tunnelDetails struct {
	Name        string     `json:"name"`
	State       string     `json:"state"`
	Full        bool       `json:"full"`
	KillSwitch  bool       `json:"killSwitch"`
	Interface   ifaceView  `json:"interface"`
	Obfuscation []kv       `json:"obfuscation"`
	Peers       []peerView `json:"peers"`
	Error       string     `json:"error,omitempty"`
}

func handshakeUnix(t conf.HandshakeTime) int64 {
	if t.IsEmpty() {
		return 0
	}
	return int64(time.Duration(t) / time.Second)
}

func ipsToStrings(ips []conf.IPCidr) []string {
	out := make([]string, 0, len(ips))
	for i := range ips {
		out = append(out, ips[i].String())
	}
	return out
}

func obfuscationOf(i *conf.Interface) []kv {
	var out []kv
	addU := func(k string, v uint16) {
		if v != 0 {
			out = append(out, kv{k, strconv.Itoa(int(v))})
		}
	}
	addS := func(k, v string) {
		if v != "" {
			out = append(out, kv{k, v})
		}
	}
	addU("Jc", i.JunkPacketCount)
	addU("Jmin", i.JunkPacketMinSize)
	addU("Jmax", i.JunkPacketMaxSize)
	addU("S1", i.InitPacketJunkSize)
	addU("S2", i.ResponsePacketJunkSize)
	addU("S3", i.CookieReplyPacketJunkSize)
	addU("S4", i.TransportPacketJunkSize)
	addS("H1", i.InitPacketMagicHeader)
	addS("H2", i.ResponsePacketMagicHeader)
	addS("H3", i.UnderloadPacketMagicHeader)
	addS("H4", i.TransportPacketMagicHeader)
	ikeys := make([]string, 0, len(i.IPackets))
	for k := range i.IPackets {
		ikeys = append(ikeys, k)
	}
	sort.Strings(ikeys)
	for _, k := range ikeys {
		addS(strings.ToUpper(k[:1])+k[1:], i.IPackets[k])
	}
	if !i.HeaderProtectionKey.IsZero() {
		out = append(out, kv{"HeaderProtectionKey", "set"})
	}
	addS("ContentPaddingAddition", i.ContentPaddingAddition)
	addS("RekeyAfterTime", i.RekeyAfterTime)
	addS("RekeyTimeout", i.RekeyTimeout)
	addS("RejectAfterTime", i.RejectAfterTime)
	addS("KeepaliveTimeout", i.KeepaliveTimeout)
	addS("MaxHandshakeAttempts", i.MaxHandshakeAttempts)
	addS("RandomTrailers", i.RandomTrailers)
	addS("DisableCookies", i.DisableCookies)
	return out
}

func (a *app) rpcDetails(w *webWindow, raw json.RawMessage) (any, error) {
	var args nameArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	cfg, err := a.be.StoredConfig(args.Name)
	if err != nil {
		return nil, codeErr("noTunnel", err)
	}
	st, _ := a.be.State(args.Name)
	d := tunnelDetails{
		Name:        args.Name,
		State:       stateString(st),
		Full:        routesAll(&cfg),
		KillSwitch:  killSwitchOn(&cfg),
		Obfuscation: obfuscationOf(&cfg.Interface),
	}
	i := &cfg.Interface
	d.Interface = ifaceView{
		PublicKey: i.PrivateKey.Public().String(),
		Addresses: ipsToStrings(i.Addresses),
		TableOff:  i.TableOff,
		PreUp:     i.PreUp, PostUp: i.PostUp, PreDown: i.PreDown, PostDown: i.PostDown,
	}
	if i.ListenPort != 0 {
		d.Interface.ListenPort = strconv.Itoa(int(i.ListenPort))
	}
	if i.MTU != 0 {
		d.Interface.MTU = strconv.Itoa(int(i.MTU))
	}
	for _, ip := range i.DNS {
		d.Interface.DNS = append(d.Interface.DNS, ip.String())
	}
	d.Interface.DNS = append(d.Interface.DNS, i.DNSSearch...)

	var runtimePeers map[conf.Key]conf.Peer
	if st == manager.TunnelStarted {
		if rc, err := a.be.RuntimeConfig(args.Name); err == nil {
			runtimePeers = map[conf.Key]conf.Peer{}
			for _, p := range rc.Peers {
				runtimePeers[p.PublicKey] = p
			}
		}
	}
	for _, p := range cfg.Peers {
		pv := peerView{
			PublicKey:  p.PublicKey.String(),
			HasPSK:     !p.PresharedKey.IsZero(),
			AllowedIPs: ipsToStrings(p.AllowedIPs),
			Keepalive:  p.PersistentKeepalive,
		}
		if pv.HasPSK && a.isAdmin {
			pv.PresharedKey = p.PresharedKey.String()
		}
		if !p.Endpoint.IsEmpty() {
			pv.Endpoint = p.Endpoint.String()
		}
		if rp, ok := runtimePeers[p.PublicKey]; ok {
			pv.Rx, pv.Tx = uint64(rp.RxBytes), uint64(rp.TxBytes)
			pv.Handshake = handshakeUnix(rp.LastHandshakeTime)
		}
		d.Peers = append(d.Peers, pv)
	}
	a.mu.Lock()
	d.Error = a.lastErrors[args.Name]
	a.mu.Unlock()
	return d, nil
}

type peerStats struct {
	PublicKey string `json:"publicKey"`
	Rx        uint64 `json:"rx"`
	Tx        uint64 `json:"tx"`
	Handshake int64  `json:"handshake"`
}

func (a *app) rpcStats(w *webWindow, raw json.RawMessage) (any, error) {
	var args nameArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	rc, err := a.be.RuntimeConfig(args.Name)
	if err != nil {
		return nil, codeErr("notRunning", err)
	}
	var rx, tx uint64
	var hs int64
	peers := make([]peerStats, 0, len(rc.Peers))
	for _, p := range rc.Peers {
		ps := peerStats{PublicKey: p.PublicKey.String(), Rx: uint64(p.RxBytes), Tx: uint64(p.TxBytes), Handshake: handshakeUnix(p.LastHandshakeTime)}
		rx += ps.Rx
		tx += ps.Tx
		if ps.Handshake > hs {
			hs = ps.Handshake
		}
		peers = append(peers, ps)
	}
	return map[string]any{"rx": rx, "tx": tx, "handshake": hs, "peers": peers, "now": time.Now().UnixMilli()}, nil
}

// ---- start / stop / remove -------------------------------------------------

func (a *app) rpcStart(w *webWindow, raw json.RawMessage) (any, error) {
	var args nameArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	a.mu.Lock()
	delete(a.lastErrors, args.Name)
	a.mu.Unlock()
	// One tunnel at a time unless the user allowed several: the manager only
	// stops tunnels whose routes intersect, so stop the rest here.
	if !a.settings.get().MultipleTunnels {
		for _, other := range a.conflictingTunnels(args.Name) {
			if err := a.be.Stop(other); err != nil {
				log.Printf("Unable to stop %s before starting %s: %v", other, args.Name, err)
			}
		}
	}
	if err := a.be.Start(args.Name); err != nil {
		return nil, codeErr("startFailed", err)
	}
	return nil, nil
}

// conflictingTunnels lists the active tunnels that starting name will stop.
// In one-tunnel mode that is every other active tunnel. Otherwise it is what
// the manager stops: running tunnels whose addresses or routes intersect with
// the new one (see ManagerService.Start).
func (a *app) conflictingTunnels(name string) []string {
	cfg, err := a.be.StoredConfig(name)
	if err != nil {
		return nil
	}
	onlyOne := !a.settings.get().MultipleTunnels
	names, err := a.be.TunnelNames()
	if err != nil {
		return nil
	}
	var out []string
	for _, n := range names {
		if n == name {
			continue
		}
		st, err := a.be.State(n)
		if err != nil || (st != manager.TunnelStarted && st != manager.TunnelStarting) {
			continue
		}
		if !onlyOne {
			other, err := a.be.StoredConfig(n)
			if err != nil || !cfg.IntersectsWith(&other) {
				continue
			}
		}
		out = append(out, n)
	}
	return out
}

func (a *app) rpcConflicts(w *webWindow, raw json.RawMessage) (any, error) {
	var args nameArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	return map[string]any{"names": a.conflictingTunnels(args.Name)}, nil
}

func (a *app) rpcStop(w *webWindow, raw json.RawMessage) (any, error) {
	var args nameArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	if err := a.be.Stop(args.Name); err != nil {
		return nil, codeErr("stopFailed", err)
	}
	return nil, nil
}

func (a *app) rpcRemove(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Names []string `json:"names"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	var failed []string
	for _, n := range args.Names {
		if err := a.be.Delete(n); err != nil {
			failed = append(failed, n+": "+err.Error())
		}
	}
	if len(failed) > 0 {
		return nil, codeErr("removeFailed", errors.New(strings.Join(failed, "\n")))
	}
	return nil, nil
}

// ---- editor ----------------------------------------------------------------

func (a *app) rpcGetConfig(w *webWindow, raw json.RawMessage) (any, error) {
	var args nameArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	cfg, err := a.be.StoredConfig(args.Name)
	if err != nil {
		return nil, codeErr("noTunnel", err)
	}
	return map[string]any{"text": cfg.ToWgQuick()}, nil
}

func (a *app) rpcNewConfig(w *webWindow, _ json.RawMessage) (any, error) {
	pk, err := conf.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	cfg := conf.Config{Interface: conf.Interface{PrivateKey: *pk}}
	return map[string]any{"text": cfg.ToWgQuick()}, nil
}

type textArg struct {
	Text string `json:"text"`
}

func (a *app) rpcAnalyze(w *webWindow, raw json.RawMessage) (any, error) {
	var args textArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	return syntax.Analyze(args.Text), nil
}

func (a *app) rpcKillSwitch(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Text string `json:"text"`
		On   bool   `json:"on"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	text, err := setKillSwitch(args.Text, args.On)
	if err != nil {
		return nil, codeErr("killSwitchNA", err)
	}
	return map[string]any{"text": text}, nil
}

func (a *app) tunnelExists(name string) (bool, error) {
	names, err := a.be.TunnelNames()
	if err != nil {
		return false, err
	}
	lower := strings.ToLower(name)
	for _, n := range names {
		if strings.ToLower(n) == lower {
			return true, nil
		}
	}
	return false, nil
}

func (a *app) rpcSave(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		OldName string `json:"oldName"`
		Name    string `json:"name"`
		Text    string `json:"text"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	if args.Name == "" {
		return nil, codeErr("nameRequired", nil)
	}
	if !conf.TunnelNameIsValid(args.Name) {
		return nil, codeErr("nameInvalid", nil)
	}
	if !strings.EqualFold(args.Name, args.OldName) {
		exists, err := a.tunnelExists(args.Name)
		if err != nil {
			return nil, codeErr("listFailed", err)
		}
		if exists {
			return nil, codeErr("nameExists", nil)
		}
	}
	cfg, err := conf.FromWgQuick(args.Text, args.Name)
	if err != nil {
		return nil, codeErr("parse", err)
	}
	if args.OldName == "" {
		if err := a.be.Create(cfg); err != nil {
			return nil, codeErr("createFailed", err)
		}
		return map[string]any{"name": cfg.Name}, nil
	}
	// Same as the original client: delete, wait, recreate and restart if it
	// was running.
	prior, stateErr := a.be.State(args.OldName)
	a.be.Delete(args.OldName)
	a.be.WaitForStop(args.OldName)
	if err := a.be.Create(cfg); err != nil {
		return nil, codeErr("createFailed", err)
	}
	if stateErr == nil && (prior == manager.TunnelStarting || prior == manager.TunnelStarted) {
		a.be.Start(cfg.Name)
	}
	return map[string]any{"name": cfg.Name}, nil
}

// ---- import ----------------------------------------------------------------

type unparsedConfig struct {
	Name   string
	Config string
}

type importResult struct {
	Imported int      `json:"imported"`
	Total    int      `json:"total"`
	Names    []string `json:"names"`
	Errors   []string `json:"errors"`
}

func configsFromZip(data []byte) ([]unparsedConfig, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	var out []unparsedConfig
	for _, f := range r.File {
		if strings.ToLower(filepath.Ext(f.Name)) != ".conf" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return out, err
		}
		b, err := io.ReadAll(io.LimitReader(rc, 1<<20))
		rc.Close()
		if err != nil {
			return out, err
		}
		out = append(out, unparsedConfig{Name: strings.TrimSuffix(filepath.Base(f.Name), filepath.Ext(f.Name)), Config: string(b)})
	}
	return out, nil
}

func configsFromFile(name string, data []byte) ([]unparsedConfig, error) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".conf":
		return []unparsedConfig{{Name: strings.TrimSuffix(filepath.Base(name), filepath.Ext(name)), Config: string(data)}}, nil
	case ".zip":
		return configsFromZip(data)
	}
	return nil, fmt.Errorf("%s: unsupported file type", filepath.Base(name))
}

func (a *app) importConfigs(configs []unparsedConfig, firstErr error) (importResult, error) {
	res := importResult{Total: len(configs)}
	if firstErr != nil {
		res.Errors = append(res.Errors, firstErr.Error())
	}
	if len(configs) == 0 {
		if firstErr != nil {
			return res, codeErr("importFailed", firstErr)
		}
		return res, codeErr("noConfigs", nil)
	}
	sort.Slice(configs, func(i, j int) bool { return conf.TunnelNameIsLess(configs[i].Name, configs[j].Name) })
	names, err := a.be.TunnelNames()
	if err != nil {
		return res, codeErr("listFailed", err)
	}
	existing := map[string]bool{}
	for _, n := range names {
		existing[strings.ToLower(n)] = true
	}
	for _, c := range configs {
		if existing[strings.ToLower(c.Name)] {
			res.Errors = append(res.Errors, fmt.Sprintf("‘%s’: exists", c.Name))
			continue
		}
		cfg, err := conf.FromWgQuickWithUnknownEncoding(c.Config, c.Name)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("‘%s’: %v", c.Name, err))
			continue
		}
		if err := a.be.Create(cfg); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("‘%s’: %v", c.Name, err))
			continue
		}
		existing[strings.ToLower(c.Name)] = true
		res.Imported++
		res.Names = append(res.Names, cfg.Name)
	}
	return res, nil
}

var confFilter = fileFilter{{"Configuration files (*.zip, *.conf)", "*.zip;*.conf"}, {"All files (*.*)", "*.*"}}

func (a *app) rpcImportDialog(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Title string `json:"title"`
	}
	decodeArgs(raw, &args)
	paths := openFilesDialog(w.hwnd, args.Title, confFilter, true)
	if paths == nil {
		return map[string]any{"cancelled": true}, nil
	}
	go func() {
		var configs []unparsedConfig
		var lastErr error
		for _, p := range paths {
			data, err := os.ReadFile(p)
			if err != nil {
				lastErr = err
				continue
			}
			c, err := configsFromFile(p, data)
			if err != nil {
				lastErr = err
			}
			configs = append(configs, c...)
		}
		res, err := a.importConfigs(configs, lastErr)
		a.emitImport(res, err)
	}()
	return map[string]any{"started": true}, nil
}

func (a *app) emitImport(res importResult, err error) {
	payload := map[string]any{"result": res}
	if err != nil {
		var re *rpcError
		if errors.As(err, &re) {
			payload["error"] = re
		}
	}
	a.emit("imported", payload)
}

func (a *app) rpcImportData(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Files []struct {
			Name string `json:"name"`
			Data string `json:"data"` // base64
		} `json:"files"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	var configs []unparsedConfig
	var lastErr error
	for _, f := range args.Files {
		data, err := base64.StdEncoding.DecodeString(f.Data)
		if err != nil {
			lastErr = err
			continue
		}
		c, err := configsFromFile(f.Name, data)
		if err != nil {
			lastErr = err
		}
		configs = append(configs, c...)
	}
	return a.importConfigs(configs, lastErr)
}

func suggestName(text string) string {
	// Use the first "# Name = x" style comment or the endpoint host.
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "endpoint") {
			if i := strings.Index(line, "="); i >= 0 {
				host := strings.TrimSpace(line[i+1:])
				if h, _, err := net.SplitHostPort(host); err == nil {
					host = h
				}
				host = strings.Map(func(r rune) rune {
					if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
						return r
					}
					return '-'
				}, host)
				host = strings.Trim(host, "-")
				if len(host) > 32 {
					host = host[:32]
				}
				if conf.TunnelNameIsValid(host) {
					return host
				}
			}
		}
	}
	return "tunnel"
}

func (a *app) uniqueName(base string) string {
	name := base
	for i := 2; i < 1000; i++ {
		exists, err := a.tunnelExists(name)
		if err != nil || !exists {
			return name
		}
		name = fmt.Sprintf("%s-%d", base, i)
		if len(name) > 32 {
			name = name[len(name)-32:]
		}
	}
	return name
}

// importText imports a configuration given as text (clipboard or QR code).
// It does not save it: the page opens the editor so the user can review it
// and choose the name.
func (a *app) prepareText(text string) (any, error) {
	text = strings.TrimSpace(strings.TrimPrefix(text, string(rune(0xFEFF))))
	if strings.HasPrefix(strings.ToLower(text), "vpn://") {
		return nil, codeErr("vpnLinkUnsupported", nil)
	}
	if !strings.Contains(text, "[Interface]") {
		return nil, codeErr("notAConfig", nil)
	}
	if _, err := conf.FromWgQuickWithUnknownEncoding(text, "temporary"); err != nil {
		return nil, codeErr("parse", err)
	}
	return map[string]any{"text": text + "\n", "name": a.uniqueName(suggestName(text))}, nil
}

func (a *app) rpcImportText(w *webWindow, raw json.RawMessage) (any, error) {
	var args textArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	return a.prepareText(args.Text)
}

func (a *app) rpcImportQR(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Source string `json:"source"` // "screen" or "file"
		Title  string `json:"title"`
	}
	decodeArgs(raw, &args)
	// Runs on a worker goroutine: only the file dialog goes to the UI thread,
	// the screen grab and decoding stay off it.
	var text string
	var err error
	if args.Source == "file" {
		var paths []string
		onUI(func() {
			paths = openFilesDialog(w.hwnd, args.Title, fileFilter{{"Images (*.png, *.jpg, *.gif)", "*.png;*.jpg;*.jpeg;*.gif"}, {"All files (*.*)", "*.*"}}, false)
		})
		if len(paths) == 0 {
			return map[string]any{"cancelled": true}, nil
		}
		text, err = decodeQRFromFile(paths[0])
	} else {
		// Get our window out of the way so it does not cover the code.
		onUI(func() { showWindow(w.hwnd, swMinimize) })
		time.Sleep(350 * time.Millisecond)
		text, err = decodeQRFromScreen()
		onUI(func() { showWindow(w.hwnd, swRestore); procSetForegroundWindow.Call(uintptr(w.hwnd)) })
	}
	if err != nil {
		if errors.Is(err, errNoQRCode) {
			return nil, codeErr("noQR", nil)
		}
		return nil, codeErr("error", err)
	}
	return a.prepareText(text)
}

// onUI runs f on the UI thread and waits for it. Never call it from the UI
// thread itself.
func onUI(f func()) {
	done := make(chan struct{})
	dispatch(func() {
		defer close(done)
		f()
	})
	<-done
}

func (a *app) rpcClipboardText(w *webWindow, _ json.RawMessage) (any, error) {
	text, err := clipboardText()
	if err != nil {
		return nil, codeErr("clipboard", err)
	}
	return map[string]any{"text": text}, nil
}

func (a *app) rpcCopyText(w *webWindow, raw json.RawMessage) (any, error) {
	var args textArg
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	if err := setClipboardText(args.Text); err != nil {
		return nil, codeErr("clipboard", err)
	}
	return nil, nil
}

// ---- export ----------------------------------------------------------------

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (a *app) rpcExportAll(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Title string   `json:"title"`
		Names []string `json:"names"` // empty: all
	}
	decodeArgs(raw, &args)
	path := saveFileDialog(w.hwnd, args.Title, "netportal-tunnels.zip", ".zip", fileFilter{{"ZIP (*.zip)", "*.zip"}})
	if path == "" {
		return map[string]any{"cancelled": true}, nil
	}
	names := args.Names
	if len(names) == 0 {
		var err error
		if names, err = a.be.TunnelNames(); err != nil {
			return nil, codeErr("listFailed", err)
		}
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, n := range names {
		cfg, err := a.be.StoredConfig(n)
		if err != nil {
			return nil, codeErr("exportFailed", err)
		}
		f, err := zw.Create(n + ".conf")
		if err != nil {
			return nil, codeErr("exportFailed", err)
		}
		if _, err := f.Write([]byte(cfg.ToWgQuick())); err != nil {
			return nil, codeErr("exportFailed", err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, codeErr("exportFailed", err)
	}
	if err := writeFileAtomic(path, buf.Bytes()); err != nil {
		return nil, codeErr("exportFailed", err)
	}
	return map[string]any{"path": path, "count": len(names)}, nil
}

func (a *app) rpcExportTunnel(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Name  string `json:"name"`
		Title string `json:"title"`
	}
	decodeArgs(raw, &args)
	cfg, err := a.be.StoredConfig(args.Name)
	if err != nil {
		return nil, codeErr("noTunnel", err)
	}
	path := saveFileDialog(w.hwnd, args.Title, args.Name+".conf", ".conf", fileFilter{{"Configuration (*.conf)", "*.conf"}})
	if path == "" {
		return map[string]any{"cancelled": true}, nil
	}
	if err := writeFileAtomic(path, []byte(cfg.ToWgQuick())); err != nil {
		return nil, codeErr("exportFailed", err)
	}
	return map[string]any{"path": path}, nil
}

// ---- log -------------------------------------------------------------------

type logLine struct {
	Stamp int64  `json:"t"` // unix ms
	Line  string `json:"m"`
}

var demoLogStart = time.Now()

func (a *app) rpcLog(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Cursor *uint32 `json:"cursor"`
	}
	decodeArgs(raw, &args)
	if a.demo || ringlogger.Global == nil {
		if args.Cursor != nil {
			return map[string]any{"lines": []logLine{}, "cursor": *args.Cursor}, nil
		}
		return map[string]any{"lines": demoLog(), "cursor": 1}, nil
	}
	cursor := ringlogger.CursorAll
	if args.Cursor != nil {
		cursor = *args.Cursor
	}
	items, next := ringlogger.Global.FollowFromCursor(cursor)
	lines := make([]logLine, 0, len(items))
	for _, it := range items {
		lines = append(lines, logLine{Stamp: it.Stamp.UnixMilli(), Line: it.Line})
	}
	return map[string]any{"lines": lines, "cursor": next}, nil
}

func demoLog() []logLine {
	msgs := []string{
		"[MGR] Starting NET PORTAL manager",
		"[GUI] Starting UI process",
		"[TUN] [ams-01] Starting AmneziaWG/3.1.1 (Windows 11 Pro; amd64)",
		"[TUN] [ams-01] Watching network interfaces",
		"[TUN] [ams-01] Resolving DNS names",
		"[TUN] [ams-01] Creating network adapter",
		"[TUN] [ams-01] Using existing driver 0.14",
		"[TUN] [ams-01] Creating adapter",
		"[TUN] [ams-01] Enabling firewall rules",
		"[TUN] [ams-01] Interface created",
		"[TUN] [ams-01] Dropping privileges",
		"[TUN] [ams-01] Setting interface configuration",
		"[TUN] [ams-01] Peer 1 created",
		"[TUN] [ams-01] Sending handshake initiation to peer 1 (203.0.113.24:51820)",
		"[TUN] [ams-01] Receiving handshake response from peer 1 (203.0.113.24:51820)",
		"[TUN] [ams-01] Keypair 1 created for peer 1",
		"[TUN] [ams-01] Startup complete",
		"[TUN] [hel-backup] Handshake for peer 1 (198.51.100.7:51820) did not complete after 5 seconds, retrying (try 2)",
		"[TUN] [hel-backup] Unable to complete handshake: peer not responding",
		"[TUN] [ams-01] Sending keepalive packet to peer 1 (203.0.113.24:51820)",
	}
	out := make([]logLine, len(msgs))
	for i, m := range msgs {
		out[i] = logLine{Stamp: demoLogStart.Add(time.Duration(i*1700) * time.Millisecond).UnixMilli(), Line: m}
	}
	return out
}

func (a *app) rpcSaveLog(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Title string `json:"title"`
	}
	decodeArgs(raw, &args)
	path := saveFileDialog(w.hwnd, args.Title, "netportal-log-"+time.Now().Format("2006-01-02T150405")+".txt", ".txt", fileFilter{{"Text files (*.txt)", "*.txt"}, {"All files (*.*)", "*.*"}})
	if path == "" {
		return map[string]any{"cancelled": true}, nil
	}
	var buf bytes.Buffer
	if a.demo || ringlogger.Global == nil {
		for _, l := range demoLog() {
			fmt.Fprintf(&buf, "%s: %s\r\n", time.UnixMilli(l.Stamp).Format("2006-01-02 15:04:05.000000"), l.Line)
		}
	} else if _, err := ringlogger.Global.WriteTo(&buf); err != nil {
		return nil, codeErr("exportFailed", err)
	}
	if err := writeFileAtomic(path, buf.Bytes()); err != nil {
		return nil, codeErr("exportFailed", err)
	}
	return map[string]any{"path": path}, nil
}

// ---- settings / update / misc ----------------------------------------------

func (a *app) rpcSetSettings(w *webWindow, raw json.RawMessage) (any, error) {
	s := a.settings.get()
	if err := decodeArgs(raw, &s); err != nil {
		return nil, err
	}
	if err := a.settings.set(s); err != nil {
		return nil, codeErr("settingsFailed", err)
	}
	a.emit("settings", a.settings.get())
	go a.refreshTray()
	return a.settings.get(), nil
}

func (a *app) rpcSetAutostart(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		On bool `json:"on"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	if err := a.settings.setAutostart(args.On); err != nil {
		return nil, codeErr("autostartFailed", err)
	}
	return nil, nil
}

func (a *app) rpcUpdate(w *webWindow, _ json.RawMessage) (any, error) {
	s := a.settings.get()
	s.ShowOnNextStart = true
	a.settings.set(s)
	if err := a.be.Update(); err != nil {
		return nil, codeErr("updateFailed", err)
	}
	return nil, nil
}

func (a *app) rpcExternalIP(w *webWindow, _ json.RawMessage) (any, error) {
	if !a.settings.get().ExternalIP {
		return nil, codeErr("disabled", nil)
	}
	if a.demo {
		return ipInfo{IP: "203.0.113.24", City: "Amsterdam", Country: "Netherlands", Code: "NL"}, nil
	}
	info, err := externalIP(context.Background())
	if err != nil {
		return nil, codeErr("externalIPFailed", err)
	}
	return info, nil
}

func (a *app) rpcOpenURL(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		URL string `json:"url"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(args.URL, "https://") {
		return nil, codeErr("badURL", nil)
	}
	if err := windows.ShellExecute(0, windows.StringToUTF16Ptr("open"), windows.StringToUTF16Ptr(args.URL), nil, nil, windows.SW_SHOWNORMAL); err != nil {
		return nil, codeErr("error", err)
	}
	return nil, nil
}

func rgb(r, g, b byte) uint32 { return uint32(r) | uint32(g)<<8 | uint32(b)<<16 }

func (a *app) rpcSetChrome(w *webWindow, raw json.RawMessage) (any, error) {
	var args struct {
		Dark bool `json:"dark"`
	}
	decodeArgs(raw, &args)
	// Match the design's window background (c.win) and secondary text.
	if args.Dark {
		w.setTitleBarColors(true, rgb(0x1A, 0x1C, 0x20), rgb(0xB6, 0xBA, 0xC2))
	} else {
		w.setTitleBarColors(false, rgb(0xE9, 0xEA, 0xEB), rgb(0x3E, 0x42, 0x45))
	}
	return nil, nil
}

func (a *app) rpcShowMain(w *webWindow, raw json.RawMessage) (any, error) {
	a.showMain()
	var args struct {
		Screen string `json:"screen"`
		Tunnel string `json:"tunnel"`
	}
	decodeArgs(raw, &args)
	if args.Screen != "" || args.Tunnel != "" {
		a.emit("navigate", args)
	}
	return nil, nil
}

func (a *app) rpcHidePopup(w *webWindow, _ json.RawMessage) (any, error) {
	if a.popup != nil {
		a.popup.hide()
	}
	return nil, nil
}

func (a *app) rpcQuit(w *webWindow, _ json.RawMessage) (any, error) {
	dispatch(func() { a.exit(true) })
	return nil, nil
}

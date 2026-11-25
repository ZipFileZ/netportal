/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

// Package webui is the NET PORTAL user interface: a Win32 window hosting a
// WebView2 control that renders an HTML/CSS/JS frontend (see assets/). It
// replaces the walk-based ui package and talks to the manager service over
// the same IPC API.
package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/jchv/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"

	"github.com/amnezia-vpn/amneziawg-windows-client/manager"
	"github.com/amnezia-vpn/amneziawg-windows-client/updater"
)

// IsAdmin is set by main before RunUI; non-admins get a read-only UI, like
// in the original client.
var IsAdmin = false

//go:embed assets
var assets embed.FS

const appName = "NET PORTAL"

type app struct {
	be       backend
	demo     bool
	isAdmin  bool
	settings *settingsStore

	main  *webWindow
	popup *webWindow
	tray  *trayIcon

	mu          sync.Mutex
	states      map[string]manager.TunnelState
	lastErrors  map[string]string
	updateState manager.UpdateState
	quitManager bool
	started     time.Time

	monitors   map[string]*healthMonitor // connection checks, by tunnel
	health     map[string]healthInfo
	reconnects map[string]int // automatic reconnects in a row
}

// RunUI runs the interface for the manager service. It returns when the user
// exits or the manager stops.
func RunUI() {
	run(ipcBackend{}, false, IsAdmin)
}

// RunDemo runs the interface against an in-memory backend, for previewing the
// design without installing or touching any service.
// demoHints are optional "key=value" switches for /uidemo (screen, theme,
// overlay, lang, size) used to preview a particular state.
var demoHints = map[string]string{}

func RunDemo(args ...string) {
	for _, a := range args {
		if k, v, ok := strings.Cut(a, "="); ok {
			demoHints[k] = v
		}
	}
	run(newDemoBackend(), true, true)
}

func run(be backend, demo, isAdmin bool) {
	runtime.LockOSThread()
	windows.SetProcessPriorityBoost(windows.CurrentProcess(), false)
	defer func() {
		if err := recover(); err != nil {
			messageBox(0, "Panic", fmt.Sprint(err, "\n\n", string(debug.Stack())), windows.MB_ICONERROR)
			panic(err)
		}
	}()

	a := &app{
		be:         be,
		demo:       demo,
		isAdmin:    isAdmin,
		settings:   newSettingsStore(!demo),
		states:     map[string]manager.TunnelState{},
		lastErrors: map[string]string{},
		started:    time.Now(),
	}

	var err error
	for attempt := 0; ; attempt++ {
		a.main, err = newWebWindow(windowOptions{
			class:      manageWindowClass,
			title:      a.title(),
			style:      wsFixedWindow,
			w:          demoSize(0, 1040),
			h:          demoSize(1, 680),
			background: edge.COREWEBVIEW2_COLOR{A: 255, R: 0x1A, G: 0x1C, B: 0x20},
			devTools:   demo,
		})
		if err == nil {
			break
		}
		// Right after logon the manager starts us before WebView2 is ready
		// to create controllers, so keep retrying for about half a minute.
		if attempt >= 20 {
			messageBox(0, appName, err.Error(), windows.MB_ICONERROR)
			return
		}
		time.Sleep(time.Duration(500+attempt*100) * time.Millisecond)
	}
	procChangeWindowMsgFilterEx.Call(uintptr(a.main.hwnd), uintptr(raiseMsg), 1 /* MSGFLT_ALLOW */, 0)
	a.main.fit, a.main.wantW, a.main.wantH = true, demoSize(0, 1040), demoSize(1, 680)
	a.main.fitToMonitor(true)

	a.tray = newTrayIcon(a.main.hwnd)
	a.tray.onClick = a.togglePopup
	a.tray.onDoubleClick = a.showMain
	trayOK := a.tray.add()
	a.tray.set(trayIconOff, appName)

	a.main.onMessage = func(s string) { a.handleMessage(a.main, s) }
	a.main.onClose = func() {
		if trayOK {
			a.main.hide()
		} else {
			a.exit(false)
		}
	}
	a.main.onWndProc = func(m uint32, wp, lp uintptr) (uintptr, bool) {
		if a.tray.handle(m, wp, lp) {
			return 0, true
		}
		if m == raiseMsg {
			a.showMain()
			return 0, true
		}
		return 0, false
	}
	a.main.chromium.NavigateToString(a.page("main"))

	a.be.Subscribe(backendEvents{
		TunnelChanged:   a.onTunnelChanged,
		TunnelsChanged:  func() { a.emit("tunnels", nil) },
		ManagerStopping: func() { dispatch(func() { a.exit(false) }) },
		UpdateFound:     a.onUpdateFound,
		UpdateProgress:  a.onUpdateProgress,
	})
	go func() {
		if st, err := a.be.UpdateState(); err == nil {
			a.onUpdateFound(st)
		}
	}()
	go a.refreshTray()
	go a.autoConnect()
	go a.syncHealth()

	if s := a.settings.get(); s.ShowOnNextStart {
		s.ShowOnNextStart = false
		a.settings.set(s)
		a.showMain()
	} else if !trayOK || !s.StartMinimized || (demo && demoHints["hidden"] != "1") {
		a.showMain()
	}
	if demo && demoHints["tray"] == "1" {
		go func() {
			time.Sleep(1500 * time.Millisecond)
			dispatch(a.togglePopup)
		}()
	}

	runMessageLoop()

	a.tray.remove()
	if a.quitManager && !demo {
		if err := a.be.Quit(true); err != nil {
			messageBox(0, appName, fmt.Sprintf("Unable to exit service due to: %v. You may want to stop the manager from the service manager.", err), windows.MB_ICONERROR)
		}
	}
}

func (a *app) title() string {
	if a.demo {
		return appName + " (demo)"
	}
	return appName
}

func (a *app) page(view string) string {
	read := func(name string) string {
		b, err := assets.ReadFile("assets/" + name)
		if err != nil {
			panic(err)
		}
		return string(b)
	}
	boot, _ := json.Marshal(map[string]any{"view": view, "demo": a.demo, "hints": demoHints})
	html := read("index.html")
	// The page is loaded with NavigateToString, so styles and scripts are
	// inlined. Use functions so "$" in the sources is not interpreted.
	html = strings.Replace(html, "/*NP_CSS*/", read("app.css"), 1)
	html = strings.Replace(html, "/*NP_BOOT*/", "window.NP_BOOT="+string(boot)+";", 1)
	html = strings.Replace(html, "/*NP_JS*/", read("i18n.js")+"\n"+read("app.js"), 1)
	return html
}

func (a *app) showMain() {
	if a.popup != nil {
		a.popup.hide()
	}
	a.main.show()
}

func (a *app) togglePopup() {
	if a.popup == nil {
		p, err := newWebWindow(windowOptions{
			class:      trayPopupClass,
			title:      appName,
			style:      wsPopup,
			exStyle:    wsExToolWindow | wsExTopmost,
			w:          360,
			h:          560,
			background: edge.COREWEBVIEW2_COLOR{A: 255, R: 0x24, G: 0x27, B: 0x2C},
			devTools:   a.demo,
		})
		if err != nil {
			a.showMain()
			return
		}
		setDwmInt(p.hwnd, dwmwaWindowCornerPref, dwmwcpRound)
		p.onMessage = func(s string) { a.handleMessage(p, s) }
		p.onClose = func() { p.hide() }
		p.onDeactivate = func() {
			// Let a click on the tray icon that caused the deactivation
			// toggle it shut instead of immediately reopening it.
			hiddenAt = time.Now()
			p.hide()
		}
		p.chromium.NavigateToString(a.page("tray"))
		a.popup = p
	}
	if a.popup.visible() {
		a.popup.hide()
		return
	}
	if time.Since(hiddenAt) < 300*time.Millisecond {
		return
	}
	a.positionPopup()
	a.popup.show()
	a.popup.eval("window.np&&np._event({name:'shown'})")
}

var hiddenAt time.Time

func (a *app) positionPopup() {
	dpi := dpiForWindow(a.popup.hwnd)
	w, h := scale(360, dpi), scale(560, dpi)
	margin := scale(12, dpi)
	c := cursorPos()
	work := workAreaAt(c)
	x := work.Right - w - margin
	y := work.Bottom - h - margin
	// Taskbar at the top or left: anchor to the cursor side instead.
	if c.Y < work.Top {
		y = work.Top + margin
	}
	if c.X < work.Left {
		x = work.Left + margin
	}
	setWindowPos(a.popup.hwnd, hwndTopmost, x, y, w, h, swpNoActivate)
}

func (a *app) exit(quitManager bool) {
	a.quitManager = quitManager
	if a.popup != nil {
		procDestroyWindow.Call(uintptr(a.popup.hwnd))
		a.popup = nil
	}
	procDestroyWindow.Call(uintptr(a.main.hwnd))
}

// emit sends an event to every page. Safe to call from any goroutine.
func (a *app) emit(name string, payload any) {
	b, err := json.Marshal(map[string]any{"name": name, "data": payload})
	if err != nil {
		return
	}
	js := "window.np&&np._event(" + string(b) + ")"
	dispatch(func() {
		a.main.eval(js)
		if a.popup != nil {
			a.popup.eval(js)
		}
	})
}

func stateString(s manager.TunnelState) string {
	switch s {
	case manager.TunnelStarted:
		return "started"
	case manager.TunnelStopped:
		return "stopped"
	case manager.TunnelStarting:
		return "starting"
	case manager.TunnelStopping:
		return "stopping"
	}
	return "unknown"
}

func (a *app) onTunnelChanged(name string, state, global manager.TunnelState, err error) {
	a.mu.Lock()
	prev := a.states[name]
	a.states[name] = state
	errText := ""
	if err != nil {
		errText = err.Error()
		a.lastErrors[name] = errText
	} else if state == manager.TunnelStarting || state == manager.TunnelStarted {
		delete(a.lastErrors, name)
	}
	a.mu.Unlock()

	a.emit("tunnel", map[string]any{"name": name, "state": stateString(state), "error": errText})
	go a.syncHealth()
	go a.refreshTray()

	s := a.settings.get()
	if !s.Notifications {
		return
	}
	dispatch(func() {
		if a.main.visible() {
			return // the page shows its own toasts
		}
		tr := trText(s.Language)
		switch {
		case err != nil:
			a.tray.notify(tr("tunnelError"), fmt.Sprintf(tr("tunnelErrorBody"), name, errText), true)
		case state == manager.TunnelStarted && prev != manager.TunnelStarted:
			a.tray.notify(appName, fmt.Sprintf(tr("activated"), name), false)
		case state == manager.TunnelStopped && prev == manager.TunnelStarted:
			a.tray.notify(appName, fmt.Sprintf(tr("deactivated"), name), false)
		}
	})
}

func (a *app) refreshTray() {
	names, err := a.be.TunnelNames()
	if err != nil {
		return
	}
	icon, active, busy := uintptr(trayIconOff), "", false
	for _, n := range names {
		st, err := a.be.State(n)
		if err != nil {
			continue
		}
		a.mu.Lock()
		a.states[n] = st
		a.mu.Unlock()
		switch st {
		case manager.TunnelStarted:
			active = n
		case manager.TunnelStarting, manager.TunnelStopping:
			busy = true
			if active == "" {
				active = n
			}
		}
	}
	tr := trText(a.settings.get().Language)
	tip := appName + ": " + tr("disconnected")
	switch {
	case busy:
		icon, tip = trayIconConnecting, appName+": "+tr("connecting")+" "+active
	case active != "":
		icon, tip = trayIconConnected, appName+": "+tr("connected")+" "+active
	}
	dispatch(func() { a.tray.set(icon, tip) })
}

func (a *app) autoConnect() {
	s := a.settings.get()
	if !s.AutoConnect || s.AutoConnectTunnel == "" || a.demo {
		return
	}
	if global, err := a.be.GlobalState(); err != nil || global != manager.TunnelStopped {
		return
	}
	names, err := a.be.TunnelNames()
	if err != nil {
		return
	}
	for _, n := range names {
		if n == s.AutoConnectTunnel {
			a.be.Start(n)
			return
		}
	}
}

func (a *app) onUpdateFound(st manager.UpdateState) {
	a.mu.Lock()
	a.updateState = st
	a.mu.Unlock()
	a.emit("update", map[string]any{"state": updateStateString(st)})
}

func updateStateString(st manager.UpdateState) string {
	switch st {
	case manager.UpdateStateFoundUpdate:
		return "available"
	case manager.UpdateStateUpdatesDisabledUnofficialBuild:
		return "unofficial"
	}
	return "none"
}

func (a *app) onUpdateProgress(dp updater.DownloadProgress) {
	p := map[string]any{
		"activity": dp.Activity,
		"done":     dp.BytesDownloaded,
		"total":    dp.BytesTotal,
		"complete": dp.Complete,
	}
	if dp.Error != nil {
		p["error"] = dp.Error.Error()
	}
	a.emit("updateProgress", p)
}

// trText returns the few strings the Go side shows itself (tray tooltip and
// notifications); everything else is translated in the page.
func trText(lang string) func(string) string {
	ru := map[string]string{
		"disconnected":        "отключено",
		"connecting":          "подключение…",
		"connected":           "подключено —",
		"activated":           "Туннель %s подключён",
		"deactivated":         "Туннель %s отключён",
		"tunnelError":         "Ошибка туннеля",
		"tunnelErrorBody":     "%s: %s",
		"health.lost":         "%s: нет связи с сервером",
		"health.restored":     "%s: связь восстановлена",
		"health.reconnecting": "%s: переподключаем…",
	}
	en := map[string]string{
		"disconnected":        "disconnected",
		"connecting":          "connecting…",
		"connected":           "connected —",
		"activated":           "Tunnel %s is connected",
		"deactivated":         "Tunnel %s is disconnected",
		"tunnelError":         "Tunnel error",
		"tunnelErrorBody":     "%s: %s",
		"health.lost":         "%s: no connection to the server",
		"health.restored":     "%s: connection restored",
		"health.reconnecting": "%s: reconnecting…",
	}
	m := ru
	if lang == "en" {
		m = en
	}
	return func(k string) string { return m[k] }
}

func demoSize(i int, def int32) int32 {
	parts := strings.Split(demoHints["size"], "x")
	if len(parts) != 2 {
		return def
	}
	var v int32
	if _, err := fmt.Sscan(parts[i], &v); err != nil || v < 400 {
		return def
	}
	return v
}

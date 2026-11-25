/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	"github.com/jchv/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

const (
	manageWindowClass = "NetPortalManageWindow"
	trayPopupClass    = "NetPortalTrayPopup"
	logoIconID        = 7
)

// webWindow is a top-level Win32 window hosting a WebView2 control.
type webWindow struct {
	hwnd     windows.HWND
	chromium *edge.Chromium

	minW, minH int32 // in DIPs, 0 for none

	// fit keeps the window inside its monitor's work area. The main window
	// cannot be resized by the user, so on small or highly scaled screens
	// it is shrunk to fit instead (the page has a compact layout for that).
	fit          bool
	wantW, wantH int32 // design size in DIPs

	onMessage    func(string)
	onClose      func() // nil means destroy
	onDeactivate func()
	onWndProc    func(m uint32, w, l uintptr) (uintptr, bool)
}

var (
	windowsByHwnd   = map[windows.HWND]*webWindow{}
	wndProcCallback = windows.NewCallback(wndProc)
	registerOnce    sync.Map

	dispatchMu    sync.Mutex
	dispatchQueue []func()
	dispatchHwnd  windows.HWND
)

// dispatch runs f on the UI thread. WebView2 must only be called from there.
func dispatch(f func()) {
	dispatchMu.Lock()
	dispatchQueue = append(dispatchQueue, f)
	h := dispatchHwnd
	dispatchMu.Unlock()
	if h != 0 {
		postMessage(h, wmAppDispatch, 0, 0)
	}
}

func drainDispatch() {
	dispatchMu.Lock()
	q := dispatchQueue
	dispatchQueue = nil
	dispatchMu.Unlock()
	for _, f := range q {
		f()
	}
}

func registerClass(name string, icon windows.Handle) error {
	if _, done := registerOnce.Load(name); done {
		return nil
	}
	inst := moduleHandle()
	cursor, _, _ := procLoadCursorW.Call(0, 32512 /* IDC_ARROW */)
	// Dark background (the design's window color) so the window never
	// flashes white before the page paints.
	brush, _, _ := procCreateSolidBrush.Call(uintptr(rgb(0x1A, 0x1C, 0x20)))
	wc := wndClassEx{
		Size:       uint32(unsafe.Sizeof(wndClassEx{})),
		WndProc:    wndProcCallback,
		Instance:   inst,
		Icon:       icon,
		IconSm:     icon,
		Cursor:     windows.Handle(cursor),
		Background: windows.Handle(brush),
		ClassName:  windows.StringToUTF16Ptr(name),
	}
	if r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return err
	}
	registerOnce.Store(name, true)
	return nil
}

func webViewDataPath() string {
	base, err := windows.KnownFolderPath(windows.FOLDERID_LocalAppData, windows.KF_FLAG_CREATE)
	if err != nil {
		base = os.TempDir()
	}
	p := filepath.Join(base, "NetPortal", "WebView2")
	os.MkdirAll(p, 0o700)
	return p
}

type windowOptions struct {
	class      string
	title      string
	style      uint32
	exStyle    uint32
	x, y       int32 // in pixels; cwUseDefault if not set
	w, h       int32 // in DIPs
	minW, minH int32
	background edge.COREWEBVIEW2_COLOR
	devTools   bool
}

func newWebWindow(o windowOptions) (*webWindow, error) {
	icon := loadIcon(logoIconID, 32)
	if err := registerClass(o.class, icon); err != nil {
		return nil, err
	}
	inst := moduleHandle()
	dpi := dpiForWindow(0)
	x, y := uintptr(cwUseDefault), uintptr(cwUseDefault)
	if o.x != 0 || o.y != 0 {
		x, y = uintptr(o.x), uintptr(o.y)
	}
	h, _, err := procCreateWindowExW.Call(
		uintptr(o.exStyle),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(o.class))),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(o.title))),
		uintptr(o.style|wsClipChildren),
		x, y,
		uintptr(scale(o.w, dpi)), uintptr(scale(o.h, dpi)),
		0, 0, uintptr(inst), 0)
	if h == 0 {
		return nil, err
	}
	w := &webWindow{hwnd: windows.HWND(h), minW: o.minW, minH: o.minH}
	windowsByHwnd[w.hwnd] = w
	if dispatchHwnd == 0 {
		dispatchMu.Lock()
		dispatchHwnd = w.hwnd
		dispatchMu.Unlock()
	}

	w.chromium = edge.NewChromium()
	w.chromium.DataPath = webViewDataPath()
	w.chromium.SetPermission(edge.CoreWebView2PermissionKindClipboardRead, edge.CoreWebView2PermissionStateAllow)
	w.chromium.MessageCallback = func(s string) {
		if w.onMessage != nil {
			w.onMessage(s)
		}
	}
	if !w.chromium.Embed(uintptr(w.hwnd)) {
		code := w.chromium.InitError()
		delete(windowsByHwnd, w.hwnd)
		if dispatchHwnd == w.hwnd {
			dispatchMu.Lock()
			dispatchHwnd = 0
			dispatchMu.Unlock()
		}
		procDestroyWindow.Call(uintptr(w.hwnd))
		if code != 0 {
			return nil, fmt.Errorf("unable to start WebView2 (error 0x%08X). Please make sure the Microsoft Edge WebView2 Runtime is installed", code)
		}
		return nil, errors.New("unable to start WebView2. Please make sure the Microsoft Edge WebView2 Runtime is installed")
	}
	// go-webview2 reports any leftover GetLastError value as an error even
	// when the COM call succeeded, so only the returned object is checked.
	if settings, _ := w.chromium.GetSettings(); settings != nil {
		settings.PutAreDevToolsEnabled(o.devTools)
		settings.PutAreDefaultContextMenusEnabled(o.devTools)
		settings.PutIsStatusBarEnabled(false)
		settings.PutIsZoomControlEnabled(false)
		settings.PutIsPinchZoomEnabled(false)
		settings.PutIsSwipeNavigationEnabled(false)
		settings.PutAreBrowserAcceleratorKeysEnabled(o.devTools)
		settings.PutAreDefaultScriptDialogsEnabled(false)
	}
	if c2 := w.chromium.GetController().GetICoreWebView2Controller2(); c2 != nil {
		c2.PutDefaultBackgroundColor(o.background)
	}
	w.chromium.Resize()
	// Match the WebView to the window. Something may already have shown the
	// window while WebView2 was starting: the launcher raises it as soon as
	// it is created (WaitForRaiseUIThenQuit).
	w.syncWebViewVisibility(isWindowVisible(w.hwnd))
	return w, nil
}

// syncWebViewVisibility keeps the WebView's visibility equal to the
// window's, however the window was shown or hidden (our code, the raise from
// another process, the taskbar). A hidden WebView over a visible window is a
// blank white rectangle; a visible one in a hidden window keeps the page's
// timers running for nothing.
func (w *webWindow) syncWebViewVisibility(visible bool) {
	if w == nil || w.chromium == nil || w.chromium.GetController() == nil {
		return
	}
	if visible {
		w.chromium.Show()
	} else {
		w.chromium.Hide()
	}
}

func (w *webWindow) eval(js string) {
	if w != nil && w.chromium != nil {
		w.chromium.Eval(js)
	}
}

func (w *webWindow) show() {
	if isIconic(w.hwnd) {
		showWindow(w.hwnd, swRestore)
	} else {
		showWindow(w.hwnd, swShow)
	}
	w.chromium.Show()
	procSetForegroundWindow.Call(uintptr(w.hwnd))
	w.chromium.Focus()
}

// hide also marks the WebView as invisible so the page sees
// document.visibilityState == "hidden" and stops its timers' work.
func (w *webWindow) hide() {
	showWindow(w.hwnd, swHide)
	w.chromium.Hide()
}

func (w *webWindow) visible() bool { return isWindowVisible(w.hwnd) && !isIconic(w.hwnd) }

// setTitleBarColors makes the native caption match the page so the title bar
// blends with the sidebar, like in the design.
func (w *webWindow) setTitleBarColors(dark bool, caption, text uint32) {
	v := uint32(0)
	if dark {
		v = 1
	}
	setDwmInt(w.hwnd, dwmwaUseImmersiveDarkMode, v)
	setDwmInt(w.hwnd, dwmwaCaptionColor, caption)
	setDwmInt(w.hwnd, dwmwaTextColor, text)
}

func wndProc(hwnd windows.HWND, m uint32, wp, lp uintptr) uintptr {
	w := windowsByHwnd[hwnd]
	if w != nil && w.onWndProc != nil {
		if r, handled := w.onWndProc(m, wp, lp); handled {
			return r
		}
	}
	switch m {
	case wmAppDispatch:
		drainDispatch()
		return 0
	case wmSize:
		if w != nil && w.chromium != nil {
			w.chromium.Resize()
		}
	case wmMove, wmMoving:
		if w != nil && w.chromium != nil {
			w.chromium.NotifyParentWindowPositionChanged()
		}
	case wmSetFocus:
		if w != nil && w.chromium != nil {
			w.chromium.Focus()
		}
	case wmShowWindow:
		w.syncWebViewVisibility(wp != 0)
	case wmWindowPosChanged:
		// SetWindowPos with SWP_SHOWWINDOW/SWP_HIDEWINDOW does not send
		// WM_SHOWWINDOW, so watch the position change too.
		pos := (*windowPos)(unsafe.Pointer(lp))
		if pos.Flags&swpShowWindow != 0 {
			w.syncWebViewVisibility(true)
		} else if pos.Flags&swpHideWindow != 0 {
			w.syncWebViewVisibility(false)
		}
	case wmActivate:
		if w != nil && wp&0xffff == waInactive && w.onDeactivate != nil {
			w.onDeactivate()
		}
	case wmGetMinMaxInfo:
		if w != nil && (w.minW != 0 || w.minH != 0) {
			dpi := dpiForWindow(hwnd)
			mmi := (*minMaxInfo)(unsafe.Pointer(lp))
			mmi.MinTrackSize = point{scale(w.minW, dpi), scale(w.minH, dpi)}
			return 0
		}
	case wmDpiChanged:
		r := (*rect)(unsafe.Pointer(lp))
		setWindowPos(hwnd, 0, r.Left, r.Top, r.Right-r.Left, r.Bottom-r.Top, swpNoZOrder|swpNoActivate)
		if w != nil && w.fit {
			w.fitToMonitor(false)
		}
		return 0
	case wmClose:
		if w != nil && w.onClose != nil {
			w.onClose()
			return 0
		}
	case wmDestroy:
		delete(windowsByHwnd, hwnd)
		if hwnd == dispatchHwnd {
			procPostQuitMessage.Call(0)
		}
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(m), wp, lp)
	return r
}

func runMessageLoop() {
	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

// fitToMonitor sizes the window to its design size at the current DPI,
// shrunk to the work area of the monitor it is on, and keeps it inside that
// area. With center set it is also centered there.
func (w *webWindow) fitToMonitor(center bool) {
	mon, _, _ := procMonitorFromWindow.Call(uintptr(w.hwnd), monitorDefaultNear)
	mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	if r, _, _ := procGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi))); r == 0 {
		return
	}
	work := mi.Work
	dpi := dpiForWindow(w.hwnd)
	margin := scale(8, dpi)
	maxW := work.Right - work.Left - 2*margin
	maxH := work.Bottom - work.Top - 2*margin
	width, height := scale(w.wantW, dpi), scale(w.wantH, dpi)
	if width > maxW {
		width = maxW
	}
	if height > maxH {
		height = maxH
	}
	var cur rect
	procGetWindowRect.Call(uintptr(w.hwnd), uintptr(unsafe.Pointer(&cur)))
	x, y := cur.Left, cur.Top
	if center {
		x = work.Left + (work.Right-work.Left-width)/2
		y = work.Top + (work.Bottom-work.Top-height)/2
	}
	if x+width > work.Right-margin {
		x = work.Right - margin - width
	}
	if y+height > work.Bottom-margin {
		y = work.Bottom - margin - height
	}
	if x < work.Left+margin {
		x = work.Left + margin
	}
	if y < work.Top+margin {
		y = work.Top + margin
	}
	setWindowPos(w.hwnd, 0, x, y, width, height, swpNoZOrder|swpNoActivate)
}

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Minimal Win32 bindings used by the web UI. Kept local so the package does
// not depend on the walk toolkit.

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	comdlg32 = windows.NewLazySystemDLL("comdlg32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procShowWindow               = user32.NewProc("ShowWindow")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procSendMessageW             = user32.NewProc("SendMessageW")
	procPostQuitMessage          = user32.NewProc("PostQuitMessage")
	procLoadImageW               = user32.NewProc("LoadImageW")
	procDestroyIcon              = user32.NewProc("DestroyIcon")
	procGetDpiForWindow          = user32.NewProc("GetDpiForWindow")
	procGetDpiForSystem          = user32.NewProc("GetDpiForSystem")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procMonitorFromPoint         = user32.NewProc("MonitorFromPoint")
	procMonitorFromWindow        = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW          = user32.NewProc("GetMonitorInfoW")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procIsIconic                 = user32.NewProc("IsIconic")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procFindWindowW              = user32.NewProc("FindWindowW")
	procRegisterWindowMessageW   = user32.NewProc("RegisterWindowMessageW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput        = user32.NewProc("AttachThreadInput")
	procSetFocus                 = user32.NewProc("SetFocus")
	procSetActiveWindow          = user32.NewProc("SetActiveWindow")
	procOpenClipboard            = user32.NewProc("OpenClipboard")
	procCloseClipboard           = user32.NewProc("CloseClipboard")
	procEmptyClipboard           = user32.NewProc("EmptyClipboard")
	procGetClipboardData         = user32.NewProc("GetClipboardData")
	procSetClipboardData         = user32.NewProc("SetClipboardData")
	procGetSystemMetrics         = user32.NewProc("GetSystemMetrics")
	procGetDC                    = user32.NewProc("GetDC")
	procReleaseDC                = user32.NewProc("ReleaseDC")
	procSetWinEventHook          = user32.NewProc("SetWinEventHook")
	procUnhookWinEvent           = user32.NewProc("UnhookWinEvent")
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procSetWindowTextW           = user32.NewProc("SetWindowTextW")
	procLoadCursorW              = user32.NewProc("LoadCursorW")
	procChangeWindowMsgFilterEx  = user32.NewProc("ChangeWindowMessageFilterEx")
	procCreateSolidBrush         = gdi32.NewProc("CreateSolidBrush")

	procShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")

	procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	procGetSaveFileNameW = comdlg32.NewProc("GetSaveFileNameW")

	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")

	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalFree       = kernel32.NewProc("GlobalFree")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	wmDestroy          = 0x0002
	wmMove             = 0x0003
	wmSize             = 0x0005
	wmActivate         = 0x0006
	wmSetFocus         = 0x0007
	wmClose            = 0x0010
	wmSettingChange    = 0x001A
	wmGetMinMaxInfo    = 0x0024
	wmShowWindow       = 0x0018
	wmWindowPosChanged = 0x0047
	wmMoving           = 0x0216
	wmDpiChanged       = 0x02E0
	wmLButtonUp        = 0x0202
	wmLButtonDblClk    = 0x0203
	wmRButtonUp        = 0x0205
	wmContextMenu      = 0x007B
	wmApp              = 0x8000
	wmAppDispatch      = wmApp + 1
	wmAppTray          = wmApp + 2
	ninSelect          = 0x0400
	ninKeySelect       = 0x0401
	wsOverlappedWindow = 0x00CF0000
	// Caption, system menu and minimize box only: no sizing border and no
	// maximize box, so the window keeps its size.
	wsFixedWindow      = 0x00C00000 | 0x00080000 | 0x00020000
	wsPopup            = 0x80000000
	wsClipChildren     = 0x02000000
	wsExToolWindow     = 0x00000080
	wsExTopmost        = 0x00000008
	cwUseDefault       = ^uintptr(0x7FFFFFFF) // 0x80000000 as signed int
	swHide             = 0
	swShowNormal       = 1
	swShowMinimized    = 2
	swShow             = 5
	swMinimize         = 6
	swRestore          = 9
	swpNoSize          = 0x0001
	swpNoMove          = 0x0002
	swpNoZOrder        = 0x0004
	swpNoActivate      = 0x0010
	swpShowWindow      = 0x0040
	swpHideWindow      = 0x0080
	waInactive         = 0
	imageIcon          = 1
	lrDefaultColor     = 0
	lrShared           = 0x8000
	smCxSmIcon         = 49
	smCySmIcon         = 50
	smXVirtualScreen   = 76
	smYVirtualScreen   = 77
	smCxVirtualScreen  = 78
	smCyVirtualScreen  = 79
	monitorDefaultNear = 2
	cfUnicodeText      = 13
	gmemMoveable       = 0x0002
	srccopy            = 0x00CC0020
	captureblt         = 0x40000000
	dibRgbColors       = 0

	nimAdd        = 0
	nimModify     = 1
	nimDelete     = 2
	nimSetVersion = 4
	nifMessage    = 0x01
	nifIcon       = 0x02
	nifTip        = 0x04
	nifInfo       = 0x10
	nifShowTip    = 0x80
	niifNone      = 0x0
	niifInfo      = 0x1
	niifWarning   = 0x2
	niifError     = 0x3
	niifUser      = 0x4

	ofnAllowMultiSelect = 0x00000200
	ofnExplorer         = 0x00080000
	ofnFileMustExist    = 0x00001000
	ofnPathMustExist    = 0x00000800
	ofnOverwritePrompt  = 0x00000002
	ofnNoChangeDir      = 0x00000008

	dwmwaUseImmersiveDarkMode = 20
	dwmwaWindowCornerPref     = 33
	dwmwaBorderColor          = 34
	dwmwaCaptionColor         = 35
	dwmwaTextColor            = 36
	dwmwcpRound               = 2

	eventObjectCreate     = 0x8000
	wineventOutOfContext  = 0x0000
	wineventSkipOwnProces = 0x0002
)

type point struct{ X, Y int32 }

type rect struct{ Left, Top, Right, Bottom int32 }

type msg struct {
	Hwnd    windows.HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   windows.Handle
	Icon       windows.Handle
	Cursor     windows.Handle
	Background windows.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     windows.Handle
}

type windowPos struct {
	Hwnd        windows.HWND
	InsertAfter windows.HWND
	X, Y        int32
	Cx, Cy      int32
	Flags       uint32
}

type minMaxInfo struct {
	Reserved     point
	MaxSize      point
	MaxPosition  point
	MinTrackSize point
	MaxTrackSize point
}

type monitorInfo struct {
	Size    uint32
	Monitor rect
	Work    rect
	Flags   uint32
}

type notifyIconData struct {
	Size            uint32
	Wnd             windows.HWND
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            windows.Handle
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Version         uint32 // union with uTimeout
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        windows.GUID
	BalloonIcon     windows.Handle
}

type openFileName struct {
	StructSize    uint32
	Owner         windows.HWND
	Instance      windows.Handle
	Filter        *uint16
	CustomFilter  *uint16
	MaxCustFilter uint32
	FilterIndex   uint32
	File          *uint16
	MaxFile       uint32
	FileTitle     *uint16
	MaxFileTitle  uint32
	InitialDir    *uint16
	Title         *uint16
	Flags         uint32
	FileOffset    uint16
	FileExtension uint16
	DefExt        *uint16
	CustData      uintptr
	FnHook        uintptr
	TemplateName  *uint16
	PvReserved    uintptr
	DwReserved    uint32
	FlagsEx       uint32
}

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

func makeIntResource(id uintptr) *uint16 { return (*uint16)(unsafe.Pointer(id)) }

func showWindow(h windows.HWND, cmd int32) {
	procShowWindow.Call(uintptr(h), uintptr(cmd))
}

func setWindowPos(h windows.HWND, after uintptr, x, y, cx, cy int32, flags uint32) {
	procSetWindowPos.Call(uintptr(h), after, uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(flags))
}

func postMessage(h windows.HWND, m uint32, w, l uintptr) {
	procPostMessageW.Call(uintptr(h), uintptr(m), w, l)
}

func isWindowVisible(h windows.HWND) bool {
	r, _, _ := procIsWindowVisible.Call(uintptr(h))
	return r != 0
}

func isIconic(h windows.HWND) bool {
	r, _, _ := procIsIconic.Call(uintptr(h))
	return r != 0
}

func dpiForWindow(h windows.HWND) uint32 {
	if procGetDpiForWindow.Find() == nil && h != 0 {
		if r, _, _ := procGetDpiForWindow.Call(uintptr(h)); r != 0 {
			return uint32(r)
		}
	}
	if procGetDpiForSystem.Find() == nil {
		if r, _, _ := procGetDpiForSystem.Call(); r != 0 {
			return uint32(r)
		}
	}
	return 96
}

func scale(v int32, dpi uint32) int32 { return v * int32(dpi) / 96 }

func cursorPos() point {
	var p point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	return p
}

func workAreaAt(p point) rect {
	// POINT is passed by value: packed into one register on 64-bit, as two
	// stack slots on 32-bit.
	var mon uintptr
	if unsafe.Sizeof(uintptr(0)) == 4 {
		mon, _, _ = procMonitorFromPoint.Call(uintptr(p.X), uintptr(p.Y), monitorDefaultNear)
	} else {
		packed := uint64(uint32(p.X)) | uint64(uint32(p.Y))<<32
		mon, _, _ = procMonitorFromPoint.Call(uintptr(packed), monitorDefaultNear)
	}
	mi := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	procGetMonitorInfoW.Call(mon, uintptr(unsafe.Pointer(&mi)))
	return mi.Work
}

func setDwmInt(h windows.HWND, attr uint32, v uint32) {
	windows.DwmSetWindowAttribute(h, attr, unsafe.Pointer(&v), uint32(unsafe.Sizeof(v)))
}

func loadIcon(id uintptr, size int32) windows.Handle {
	inst := moduleHandle()
	r, _, _ := procLoadImageW.Call(uintptr(inst), id, imageIcon, uintptr(size), uintptr(size), lrDefaultColor)
	return windows.Handle(r)
}

func systemMetric(i int32) int32 {
	r, _, _ := procGetSystemMetrics.Call(uintptr(i))
	return int32(r)
}

func messageBox(owner windows.HWND, title, text string, flags uint32) {
	windows.MessageBox(owner, windows.StringToUTF16Ptr(text), windows.StringToUTF16Ptr(title), flags)
}

func clipboardText() (string, error) {
	if r, _, err := procOpenClipboard.Call(0); r == 0 {
		return "", err
	}
	defer procCloseClipboard.Call()
	h, _, err := procGetClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return "", err
	}
	p, _, err := procGlobalLock.Call(h)
	if p == 0 {
		return "", err
	}
	defer procGlobalUnlock.Call(h)
	return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(p))), nil
}

func setClipboardText(s string) error {
	u, err := windows.UTF16FromString(s)
	if err != nil {
		return err
	}
	if r, _, err := procOpenClipboard.Call(0); r == 0 {
		return err
	}
	defer procCloseClipboard.Call()
	procEmptyClipboard.Call()
	size := uintptr(len(u) * 2)
	h, _, err := procGlobalAlloc.Call(gmemMoveable, size)
	if h == 0 {
		return err
	}
	p, _, err := procGlobalLock.Call(h)
	if p == 0 {
		procGlobalFree.Call(h)
		return err
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(u)), u)
	procGlobalUnlock.Call(h)
	if r, _, err := procSetClipboardData.Call(cfUnicodeText, h); r == 0 {
		procGlobalFree.Call(h)
		return err
	}
	return nil
}

func moduleHandle() windows.Handle {
	r, _, _ := procGetModuleHandleW.Call(0)
	return windows.Handle(r)
}

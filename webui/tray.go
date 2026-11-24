/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Icon resource IDs, see resources.rc.
const (
	trayIconConnected  = 9
	trayIconOff        = 10
	trayIconConnecting = 11
)

var taskbarCreatedMsg = func() uint32 {
	r, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("TaskbarCreated"))))
	return uint32(r)
}()

// trayIcon is the notification-area icon. It is owned by the main window,
// whose window procedure forwards its callback messages to onEvent.
type trayIcon struct {
	nid     notifyIconData
	icons   map[uintptr]windows.Handle
	current uintptr
	added   bool

	onClick       func()
	onDoubleClick func()
}

func newTrayIcon(owner windows.HWND) *trayIcon {
	t := &trayIcon{icons: map[uintptr]windows.Handle{}}
	t.nid.Size = uint32(unsafe.Sizeof(t.nid))
	t.nid.Wnd = owner
	t.nid.ID = 1
	t.nid.CallbackMessage = wmAppTray
	t.current = trayIconOff
	return t
}

func (t *trayIcon) icon(id uintptr) windows.Handle {
	if h, ok := t.icons[id]; ok {
		return h
	}
	size := systemMetric(smCxSmIcon)
	if size <= 0 {
		size = 16
	}
	h := loadIcon(id, size)
	if h == 0 {
		h = loadIcon(logoIconID, size)
	}
	t.icons[id] = h
	return h
}

func setUTF16(dst []uint16, s string) {
	u, _ := windows.UTF16FromString(s)
	if len(u) > len(dst) {
		u = u[:len(dst)]
		u[len(u)-1] = 0
	}
	for i := range dst {
		dst[i] = 0
	}
	copy(dst, u)
}

func (t *trayIcon) add() bool {
	t.nid.Flags = nifMessage | nifIcon | nifTip | nifShowTip
	t.nid.Icon = t.icon(t.current)
	r, _, _ := procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&t.nid)))
	if r == 0 {
		return false
	}
	t.nid.Version = 4 // NOTIFYICON_VERSION_4
	procShellNotifyIconW.Call(nimSetVersion, uintptr(unsafe.Pointer(&t.nid)))
	t.added = true
	return true
}

func (t *trayIcon) set(iconID uintptr, tip string) {
	t.current = iconID
	setUTF16(t.nid.Tip[:], tip)
	t.nid.Icon = t.icon(iconID)
	t.nid.Flags = nifIcon | nifTip | nifShowTip
	if t.added {
		procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&t.nid)))
	}
}

func (t *trayIcon) notify(title, text string, isError bool) {
	if !t.added {
		return
	}
	n := t.nid
	n.Flags = nifInfo
	setUTF16(n.InfoTitle[:], title)
	setUTF16(n.Info[:], text)
	n.InfoFlags = niifUser
	n.BalloonIcon = t.icon(logoIconID)
	if isError {
		n.InfoFlags = niifError
	}
	procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&n)))
}

func (t *trayIcon) remove() {
	if t.added {
		procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&t.nid)))
		t.added = false
	}
	for _, h := range t.icons {
		procDestroyIcon.Call(uintptr(h))
	}
	t.icons = map[uintptr]windows.Handle{}
}

// handle processes a window message for the owner window and reports whether
// it was a tray message.
func (t *trayIcon) handle(m uint32, wp, lp uintptr) bool {
	if m == taskbarCreatedMsg && taskbarCreatedMsg != 0 {
		t.added = false
		t.add()
		return true
	}
	if m != wmAppTray {
		return false
	}
	// With NOTIFYICON_VERSION_4 the event is in LOWORD(lParam).
	switch uint32(lp & 0xffff) {
	// Version 4 also sends the raw mouse messages; only the semantic ones
	// are handled so a click is not processed twice.
	case ninSelect, ninKeySelect, wmContextMenu:
		if t.onClick != nil {
			t.onClick()
		}
	case wmLButtonDblClk:
		if t.onDoubleClick != nil {
			t.onDoubleClick()
		}
	}
	return true
}

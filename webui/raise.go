/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var raiseMsg = func() uint32 {
	r, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(manageWindowClass + "Raise"))))
	return uint32(r)
}()

const (
	hwndTopmost   = ^uintptr(0)     // -1
	hwndNoTopmost = ^uintptr(1) + 0 // -2
)

func raiseRemote(hwnd windows.HWND) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	procSendMessageW.Call(uintptr(hwnd), uintptr(raiseMsg), 0, 0)
	fg, _, _ := procGetForegroundWindow.Call()
	fgThread, _, _ := procGetWindowThreadProcessId.Call(fg, 0)
	cur := windows.GetCurrentThreadId()
	procAttachThreadInput.Call(fgThread, uintptr(cur), 1)
	setWindowPos(hwnd, hwndTopmost, 0, 0, 0, 0, swpNoMove|swpNoSize|swpShowWindow)
	setWindowPos(hwnd, hwndNoTopmost, 0, 0, 0, 0, swpNoMove|swpNoSize|swpShowWindow)
	procSetForegroundWindow.Call(uintptr(hwnd))
	procAttachThreadInput.Call(fgThread, uintptr(cur), 0)
	procSetFocus.Call(uintptr(hwnd))
	procSetActiveWindow.Call(uintptr(hwnd))
}

// RaiseUI brings an already running UI to the front and reports whether one
// was found.
func RaiseUI() bool {
	hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(manageWindowClass))), 0)
	if hwnd == 0 {
		return false
	}
	raiseRemote(windows.HWND(hwnd))
	return true
}

// WaitForRaiseUIThenQuit waits for the UI window to be created by the
// manager service, raises it and exits the process.
func WaitForRaiseUIThenQuit() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var handle uintptr
	cb := windows.NewCallback(func(hook uintptr, event uint32, hwnd windows.HWND, idObject, idChild int32, idEventThread, eventTime uint32) uintptr {
		class := make([]uint16, len(manageWindowClass)+2)
		n, _, _ := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&class[0])), uintptr(len(class)))
		if int(n) != len(manageWindowClass) || windows.UTF16ToString(class) != manageWindowClass {
			return 0
		}
		procUnhookWinEvent.Call(handle)
		raiseRemote(hwnd)
		os.Exit(0)
		return 0
	})
	handle, _, err := procSetWinEventHook.Call(eventObjectCreate, eventObjectCreate, 0, cb, 0, 0, wineventOutOfContext|wineventSkipOwnProces)
	if handle == 0 {
		messageBox(0, "NET PORTAL", fmt.Sprintf("Unable to wait for the NET PORTAL window to appear: %v", err), windows.MB_ICONERROR)
	}
	runMessageLoop()
}

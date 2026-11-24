/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// filter is a list of (description, pattern) pairs.
type fileFilter [][2]string

func (f fileFilter) utf16() *uint16 {
	var b []uint16
	for _, p := range f {
		d, _ := windows.UTF16FromString(p[0])
		s, _ := windows.UTF16FromString(p[1])
		b = append(b, d...)
		b = append(b, s...)
	}
	b = append(b, 0)
	return &b[0]
}

// openFilesDialog shows the standard open dialog and returns the selected
// paths, or nil if cancelled. Must be called on the UI thread.
func openFilesDialog(owner windows.HWND, title string, filter fileFilter, multi bool) []string {
	buf := make([]uint16, 64*1024)
	ofn := openFileName{
		StructSize: uint32(unsafe.Sizeof(openFileName{})),
		Owner:      owner,
		Filter:     filter.utf16(),
		File:       &buf[0],
		MaxFile:    uint32(len(buf)),
		Title:      windows.StringToUTF16Ptr(title),
		Flags:      ofnExplorer | ofnFileMustExist | ofnPathMustExist | ofnNoChangeDir,
	}
	if multi {
		ofn.Flags |= ofnAllowMultiSelect
	}
	if r, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn))); r == 0 {
		return nil
	}
	// Multi-select returns "dir\0file1\0file2\0\0"; single returns a full path.
	var parts []string
	start := 0
	for i := 0; i < len(buf); i++ {
		if buf[i] == 0 {
			if i == start {
				break
			}
			parts = append(parts, windows.UTF16ToString(buf[start:i]))
			start = i + 1
		}
	}
	if len(parts) <= 1 {
		return parts
	}
	dir := parts[0]
	paths := make([]string, 0, len(parts)-1)
	for _, p := range parts[1:] {
		paths = append(paths, filepath.Join(dir, p))
	}
	return paths
}

// saveFileDialog shows the standard save dialog and returns the chosen path,
// or "" if cancelled. Must be called on the UI thread.
func saveFileDialog(owner windows.HWND, title, defaultName, defExt string, filter fileFilter) string {
	buf := make([]uint16, windows.MAX_LONG_PATH)
	name, _ := windows.UTF16FromString(defaultName)
	copy(buf, name)
	ofn := openFileName{
		StructSize: uint32(unsafe.Sizeof(openFileName{})),
		Owner:      owner,
		Filter:     filter.utf16(),
		File:       &buf[0],
		MaxFile:    uint32(len(buf)),
		Title:      windows.StringToUTF16Ptr(title),
		DefExt:     windows.StringToUTF16Ptr(strings.TrimPrefix(defExt, ".")),
		Flags:      ofnExplorer | ofnPathMustExist | ofnOverwritePrompt | ofnNoChangeDir,
	}
	if r, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn))); r == 0 {
		return ""
	}
	return windows.UTF16ToString(buf)
}

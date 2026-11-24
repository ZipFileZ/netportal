/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package updater

import (
	"crypto/hmac"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/sys/windows"

	"github.com/amnezia-vpn/amneziawg-windows-client/elevate"
	"github.com/amnezia-vpn/amneziawg-windows-client/updater/winhttp"
	"github.com/amnezia-vpn/amneziawg-windows-client/version"
)

type DownloadProgress struct {
	Activity        string
	BytesDownloaded uint64
	BytesTotal      uint64
	Error           error
	Complete        bool
}

type progressHashWatcher struct {
	dp        *DownloadProgress
	c         chan DownloadProgress
	hashState hash.Hash
}

func (pm *progressHashWatcher) Write(p []byte) (int, error) {
	bytes := len(p)
	pm.dp.BytesDownloaded += uint64(bytes)
	pm.c <- *pm.dp
	pm.hashState.Write(p)
	return bytes, nil
}

type UpdateFound struct {
	name    string
	version string
	hash    [blake2b.Size256]byte
}

// Version is the version of the release that was found.
func (u *UpdateFound) Version() string { return u.version }

func CheckForUpdate() (updateFound *UpdateFound, err error) {
	updateFound, _, _, err = checkForUpdate(false)
	return
}

func checkForUpdate(keepSession bool) (*UpdateFound, *winhttp.Session, *winhttp.Connection, error) {
	if !Enabled() {
		return nil, nil, nil, errors.New("Updates are not configured in this build")
	}
	session, err := winhttp.NewSession(version.UserAgent())
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		if err != nil || !keepSession {
			session.Close()
		}
	}()
	connection, err := session.Connect(updateServerHost, updateServerPort, updateServerUseHttps)
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		if err != nil || !keepSession {
			connection.Close()
		}
	}()
	// GitHub answers with a redirect to its download host; WinHTTP follows
	// HTTPS-to-HTTPS redirects by itself.
	response, err := connection.Get(latestVersionPath, true)
	if err != nil {
		return nil, nil, nil, err
	}
	defer response.Close()
	list, err := io.ReadAll(io.LimitReader(response, 512*1024))
	if err != nil {
		return nil, nil, nil, err
	}
	files, err := readFileList(list)
	if err != nil {
		return nil, nil, nil, err
	}
	updateFound, err := findCandidate(files)
	if err != nil {
		return nil, nil, nil, err
	}
	if keepSession {
		return updateFound, session, connection, nil
	}
	return updateFound, nil, nil, nil
}

var updateInProgress = uint32(0)

// Suffixes of the files next to the executable during an update.
const (
	newSuffix = ".new"
	oldSuffix = ".old"
)

// CleanupPreviousUpdate removes the executable an earlier update replaced.
// It stays locked while tunnel services started before the update are still
// running from it, so failure is expected and ignored.
func CleanupPreviousUpdate() {
	if exe, err := os.Executable(); err == nil {
		os.Remove(exe + oldSuffix)
		os.Remove(exe + newSuffix)
	}
}

// DownloadVerifyAndExecute downloads the new executable, checks it against
// the signed hash list, swaps it in place of the running one and restarts the
// manager service. Tunnels keep running during the swap. It must run as
// SYSTEM (the manager), or it elevates to SYSTEM itself when userToken is 0.
func DownloadVerifyAndExecute(userToken uintptr) (progress chan DownloadProgress) {
	progress = make(chan DownloadProgress, 128)
	progress <- DownloadProgress{Activity: "Initializing"}

	if !atomic.CompareAndSwapUint32(&updateInProgress, 0, 1) {
		progress <- DownloadProgress{Error: errors.New("An update is already in progress")}
		return
	}

	doIt := func() {
		defer atomic.StoreUint32(&updateInProgress, 0)

		progress <- DownloadProgress{Activity: "Checking for update"}
		update, session, connection, err := checkForUpdate(true)
		if err != nil {
			progress <- DownloadProgress{Error: err}
			return
		}
		defer connection.Close()
		defer session.Close()
		if update == nil {
			progress <- DownloadProgress{Error: errors.New("No update was found")}
			return
		}

		exe, err := os.Executable()
		if err != nil {
			progress <- DownloadProgress{Error: err}
			return
		}
		newPath := exe + newSuffix
		os.Remove(newPath)
		file, err := os.OpenFile(newPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			progress <- DownloadProgress{Error: err}
			return
		}
		keep := false
		defer func() {
			file.Close()
			if !keep {
				os.Remove(newPath)
			}
		}()

		dp := DownloadProgress{Activity: "Downloading update"}
		progress <- dp
		response, err := connection.Get(fmt.Sprintf(assetPath, update.version, update.name), false)
		if err != nil {
			progress <- DownloadProgress{Error: err}
			return
		}
		defer response.Close()
		if length, err := response.Length(); err == nil && length > 0 {
			dp.BytesTotal = length
			progress <- dp
		}
		hasher, err := blake2b.New256(nil)
		if err != nil {
			progress <- DownloadProgress{Error: err}
			return
		}
		pm := &progressHashWatcher{&dp, progress, hasher}
		if _, err = io.Copy(file, io.TeeReader(io.LimitReader(response, 1024*1024*100 /* 100 MiB */), pm)); err != nil {
			progress <- DownloadProgress{Error: err}
			return
		}
		if !hmac.Equal(hasher.Sum(nil), update.hash[:]) {
			progress <- DownloadProgress{Error: errors.New("The downloaded update has the wrong hash")}
			return
		}
		if err = file.Close(); err != nil {
			progress <- DownloadProgress{Error: err}
			return
		}

		progress <- DownloadProgress{Activity: "Installing update"}
		// A running executable can be renamed but not overwritten.
		oldPath := exe + oldSuffix
		os.Remove(oldPath)
		if err = windows.Rename(exe, oldPath); err != nil {
			progress <- DownloadProgress{Error: fmt.Errorf("unable to move the current version aside: %w", err)}
			return
		}
		if err = windows.Rename(newPath, exe); err != nil {
			windows.Rename(oldPath, exe)
			progress <- DownloadProgress{Error: fmt.Errorf("unable to put the new version in place: %w", err)}
			return
		}
		keep = true

		progress <- DownloadProgress{Activity: "Restarting"}
		// The new executable restarts the manager service; this process is
		// the manager and is about to be stopped.
		cmd := exec.Command(exe, "/restartmanager")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP}
		if err = cmd.Start(); err != nil {
			progress <- DownloadProgress{Error: fmt.Errorf("the update is installed, but restarting failed: %w", err)}
			return
		}
		cmd.Process.Release()
		progress <- DownloadProgress{Complete: true}
	}
	if userToken == 0 {
		go func() {
			err := elevate.DoAsSystem(func() error {
				doIt()
				return nil
			})
			if err != nil {
				progress <- DownloadProgress{Error: err}
			}
		}()
	} else {
		// Called by the manager, which already runs as SYSTEM.
		go doIt()
	}

	return progress
}


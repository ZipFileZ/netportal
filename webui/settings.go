/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

// Settings are per-user UI preferences. They are stored as JSON in the
// roaming application data folder.
type Settings struct {
	Theme             string `json:"theme"`    // "system", "dark", "light"
	Language          string `json:"language"` // "ru", "en"
	StartMinimized    bool   `json:"startMinimized"`
	AutoConnect       bool   `json:"autoConnect"`
	AutoConnectTunnel string `json:"autoConnectTunnel"`
	Notifications     bool   `json:"notifications"`
	ExternalIP        bool   `json:"externalIP"`

	// Don't ask before connecting a tunnel that will disconnect another one.
	SkipSwitchWarning bool `json:"skipSwitchWarning"`

	// Allow several tunnels at once (only those whose routes do not
	// intersect, as the manager decides). Off: connecting a tunnel
	// disconnects every other one.
	MultipleTunnels bool `json:"multipleTunnels"`

	// Set before an update so the restarted UI opens its window instead of
	// staying in the tray.
	ShowOnNextStart bool `json:"showOnNextStart,omitempty"`
}

func defaultSettings() Settings {
	return Settings{
		Theme:    "system",
		Language: "ru",
		// The manager starts the UI at every logon; stay in the tray then.
		// Launching the app explicitly still raises the window.
		StartMinimized: true,
		Notifications:  true,
		ExternalIP:     true,
	}
}

type settingsStore struct {
	mu   sync.Mutex
	path string // empty: in memory only (demo mode)
	cur  Settings

	demoAutostart bool
}

func newSettingsStore(persist bool) *settingsStore {
	s := &settingsStore{cur: defaultSettings(), demoAutostart: true}
	if !persist {
		return s
	}
	base, err := windows.KnownFolderPath(windows.FOLDERID_RoamingAppData, windows.KF_FLAG_CREATE)
	if err != nil {
		return s
	}
	s.path = filepath.Join(base, "NetPortal", "settings.json")
	if b, err := os.ReadFile(s.path); err == nil {
		cur := defaultSettings()
		if json.Unmarshal(b, &cur) == nil {
			s.cur = cur
		}
	}
	return s
}

func (s *settingsStore) get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cur
}

func (s *settingsStore) set(n Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch n.Theme {
	case "system", "dark", "light":
	default:
		n.Theme = "system"
	}
	switch n.Language {
	case "ru", "en":
	default:
		n.Language = "ru"
	}
	s.cur = n
	if s.path == "" {
		return nil
	}
	b, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

const managerServiceName = "NetPortalManager"

// Autostart is controlled by the start type of the manager service: the
// manager launches the UI for every user who logs in.
func (s *settingsStore) autostart() (bool, error) {
	if s.path == "" {
		return s.demoAutostart, nil
	}
	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}
	defer m.Disconnect()
	service, err := m.OpenService(managerServiceName)
	if err != nil {
		return false, err
	}
	defer service.Close()
	cfg, err := service.Config()
	if err != nil {
		return false, err
	}
	return cfg.StartType == mgr.StartAutomatic, nil
}

func (s *settingsStore) setAutostart(on bool) error {
	if s.path == "" {
		s.demoAutostart = on
		return nil
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	service, err := m.OpenService(managerServiceName)
	if err != nil {
		return err
	}
	defer service.Close()
	cfg, err := service.Config()
	if err != nil {
		return err
	}
	if on {
		cfg.StartType = mgr.StartAutomatic
	} else {
		cfg.StartType = mgr.StartManual
	}
	return service.UpdateConfig(cfg)
}

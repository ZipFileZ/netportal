/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"github.com/amnezia-vpn/amneziawg-windows-client/manager"
	"github.com/amnezia-vpn/amneziawg-windows-client/updater"
	"github.com/amnezia-vpn/amneziawg-windows/v3/conf"
)

// backend is everything the UI needs from the manager service. The real
// implementation talks to the manager over IPC; the demo implementation
// (used by /uidemo) keeps everything in memory so the UI can be previewed
// without installing any service.
type backend interface {
	TunnelNames() ([]string, error)
	State(name string) (manager.TunnelState, error)
	StoredConfig(name string) (conf.Config, error)
	RuntimeConfig(name string) (conf.Config, error)
	Start(name string) error
	Stop(name string) error
	Delete(name string) error
	WaitForStop(name string) error
	Create(cfg *conf.Config) error
	GlobalState() (manager.TunnelState, error)
	UpdateState() (manager.UpdateState, error)
	Update() error
	Quit(stopTunnels bool) error
	Subscribe(sink backendEvents)
}

type backendEvents struct {
	TunnelChanged   func(name string, state, globalState manager.TunnelState, err error)
	TunnelsChanged  func()
	ManagerStopping func()
	UpdateFound     func(state manager.UpdateState)
	UpdateProgress  func(dp updater.DownloadProgress)
}

type ipcBackend struct{}

func (ipcBackend) TunnelNames() ([]string, error) {
	tunnels, err := manager.IPCClientTunnels()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(tunnels))
	for _, t := range tunnels {
		names = append(names, t.Name)
	}
	return names, nil
}

func (ipcBackend) State(name string) (manager.TunnelState, error) {
	t := manager.Tunnel{Name: name}
	return t.State()
}

func (ipcBackend) StoredConfig(name string) (conf.Config, error) {
	t := manager.Tunnel{Name: name}
	return t.StoredConfig()
}

func (ipcBackend) RuntimeConfig(name string) (conf.Config, error) {
	t := manager.Tunnel{Name: name}
	return t.RuntimeConfig()
}

func (ipcBackend) Start(name string) error {
	t := manager.Tunnel{Name: name}
	return t.Start()
}

func (ipcBackend) Stop(name string) error {
	t := manager.Tunnel{Name: name}
	return t.Stop()
}

func (ipcBackend) Delete(name string) error {
	t := manager.Tunnel{Name: name}
	return t.Delete()
}

func (ipcBackend) WaitForStop(name string) error {
	t := manager.Tunnel{Name: name}
	return t.WaitForStop()
}

func (ipcBackend) Create(cfg *conf.Config) error {
	_, err := manager.IPCClientNewTunnel(cfg)
	return err
}

func (ipcBackend) GlobalState() (manager.TunnelState, error) {
	return manager.IPCClientGlobalState()
}

func (ipcBackend) UpdateState() (manager.UpdateState, error) {
	return manager.IPCClientUpdateState()
}

func (ipcBackend) Update() error {
	return manager.IPCClientUpdate()
}

func (ipcBackend) Quit(stopTunnels bool) error {
	_, err := manager.IPCClientQuit(stopTunnels)
	return err
}

func (ipcBackend) Subscribe(sink backendEvents) {
	manager.IPCClientRegisterTunnelChange(func(tunnel *manager.Tunnel, state, globalState manager.TunnelState, err error) {
		if tunnel == nil {
			return
		}
		sink.TunnelChanged(tunnel.Name, state, globalState, err)
	})
	manager.IPCClientRegisterTunnelsChange(sink.TunnelsChanged)
	manager.IPCClientRegisterManagerStopping(sink.ManagerStopping)
	manager.IPCClientRegisterUpdateFound(sink.UpdateFound)
	manager.IPCClientRegisterUpdateProgress(sink.UpdateProgress)
}

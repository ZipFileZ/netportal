/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"errors"
	"net"

	"github.com/amnezia-vpn/amneziawg-windows/v3/conf"
)

func equalIPCidrs(a, b conf.IPCidr) bool {
	return a.IP.Equal(b.IP) && (a.Cidr == b.Cidr)
}

var errKillSwitchNotApplicable = errors.New("kill-switch is not applicable to this configuration")

// setKillSwitch rewrites the AllowedIPs of the single peer so that the tunnel
// service does (0.0.0.0/0, ::/0) or does not (two /1 halves) engage its
// firewall ruleset. This is a port of the logic from the original edit dialog.
func setKillSwitch(text string, block bool) (string, error) {
	var (
		v400    = conf.IPCidr{IP: net.IPv4zero, Cidr: 0}
		v600000 = conf.IPCidr{IP: net.IPv6zero, Cidr: 0}
		v401    = conf.IPCidr{IP: net.IPv4zero, Cidr: 1}
		v600001 = conf.IPCidr{IP: net.IPv6zero, Cidr: 1}
		v41281  = conf.IPCidr{IP: net.IPv4(0x80, 0, 0, 0), Cidr: 1}
		v680001 = conf.IPCidr{IP: net.IP{0x80, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, Cidr: 1}
	)

	cfg, err := conf.FromWgQuick(text, "temporary")
	if err != nil {
		return "", err
	}
	if len(cfg.Peers) != 1 {
		return "", errKillSwitchNotApplicable
	}

	newAllowedIPs := make([]conf.IPCidr, 0, len(cfg.Peers[0].AllowedIPs))
	if block {
		var foundV401, foundV41281, foundV600001, foundV680001 bool
		for _, allowedip := range cfg.Peers[0].AllowedIPs {
			if equalIPCidrs(allowedip, v600001) {
				foundV600001 = true
			} else if equalIPCidrs(allowedip, v680001) {
				foundV680001 = true
			} else if equalIPCidrs(allowedip, v401) {
				foundV401 = true
			} else if equalIPCidrs(allowedip, v41281) {
				foundV41281 = true
			} else {
				newAllowedIPs = append(newAllowedIPs, allowedip)
			}
		}
		if !((foundV401 && foundV41281) || (foundV600001 && foundV680001)) {
			return "", errKillSwitchNotApplicable
		}
		if foundV401 && foundV41281 {
			newAllowedIPs = append(newAllowedIPs, v400)
		} else if foundV401 {
			newAllowedIPs = append(newAllowedIPs, v401)
		} else if foundV41281 {
			newAllowedIPs = append(newAllowedIPs, v41281)
		}
		if foundV600001 && foundV680001 {
			newAllowedIPs = append(newAllowedIPs, v600000)
		} else if foundV600001 {
			newAllowedIPs = append(newAllowedIPs, v600001)
		} else if foundV680001 {
			newAllowedIPs = append(newAllowedIPs, v680001)
		}
	} else {
		var foundV400, foundV600000 bool
		for _, allowedip := range cfg.Peers[0].AllowedIPs {
			if equalIPCidrs(allowedip, v600000) {
				foundV600000 = true
			} else if equalIPCidrs(allowedip, v400) {
				foundV400 = true
			} else {
				newAllowedIPs = append(newAllowedIPs, allowedip)
			}
		}
		if !(foundV400 || foundV600000) {
			return "", errKillSwitchNotApplicable
		}
		if foundV400 {
			newAllowedIPs = append(newAllowedIPs, v401, v41281)
		}
		if foundV600000 {
			newAllowedIPs = append(newAllowedIPs, v600001, v680001)
		}
	}
	cfg.Peers[0].AllowedIPs = newAllowedIPs
	return cfg.ToWgQuick(), nil
}

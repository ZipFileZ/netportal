/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"fmt"
	"testing"

	"github.com/amnezia-vpn/amneziawg-windows/v3/conf"
)

func TestHealthTargets(t *testing.T) {
	const base = "[Interface]\nPrivateKey = yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=\nAddress = 10.8.1.2/32\nDNS = 10.8.1.1, 9.9.9.9, fd00::1\n[Peer]\nPublicKey = xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=\nEndpoint = 192.0.2.10:51820\n"
	cases := []struct{ allowed, want string }{
		{"0.0.0.0/0, ::/0", "[10.8.1.1 9.9.9.9 1.1.1.1 8.8.8.8]"},
		{"0.0.0.0/1, 128.0.0.0/1", "[10.8.1.1 9.9.9.9 1.1.1.1 8.8.8.8]"}, // kill switch off
		{"10.8.1.0/24", "[10.8.1.1]"},
		{"192.168.50.0/24", "[]"},
	}
	for _, c := range cases {
		cfg, err := conf.FromWgQuick(base+"AllowedIPs = "+c.allowed+"\n", "t")
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprint(healthTargets(cfg)); got != c.want {
			t.Errorf("AllowedIPs %s: got %s, want %s", c.allowed, got, c.want)
		}
	}
}

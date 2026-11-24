/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package webui

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// ipInfo is what the page shows as "external IP": the address the internet
// sees, and where that address is.
type ipInfo struct {
	IP      string `json:"ip"`
	City    string `json:"city,omitempty"`
	Country string `json:"country,omitempty"`
	Code    string `json:"code,omitempty"` // ISO country code
}

// externalIP asks 2ip (the API of 2ip.ru) for the address the internet sees
// and its location. Only called when the user has the feature enabled; the
// request carries no identifying information.
func externalIP(ctx context.Context) (ipInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.2ip.io", nil)
	if err != nil {
		return ipInfo{}, err
	}
	req.Header.Set("User-Agent", "NetPortal")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ipInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ipInfo{}, errors.New(resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		return ipInfo{}, err
	}
	var info ipInfo
	if err := json.Unmarshal(b, &info); err != nil {
		return ipInfo{}, errors.New("invalid response")
	}
	info.IP = strings.TrimSpace(info.IP)
	if net.ParseIP(info.IP) == nil {
		return ipInfo{}, errors.New("invalid response")
	}
	if len(info.Code) != 2 {
		info.Code = ""
	}
	return info, nil
}

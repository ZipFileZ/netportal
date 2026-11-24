/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package syntax

import (
	"strings"

	"github.com/amnezia-vpn/amneziawg-windows/v3/conf"
)

// Token is a piece of the configuration text with a coarse kind that the web
// editor maps to a color. The tokens cover the whole input, in order, so the
// editor can rebuild the text from them.
type Token struct {
	Kind string `json:"k"`
	Text string `json:"t"`
}

// Problem describes one line that has an error or a warning.
type Problem struct {
	Line    int    `json:"line"`
	Kind    string `json:"kind"` // "unknownKey", "badValue", "badSection", "warning"
	Text    string `json:"text"`
	Warning bool   `json:"warning"`
}

// Result is what the editor needs to render and validate a configuration.
type Result struct {
	Tokens     []Token   `json:"tokens"`
	Problems   []Problem `json:"problems"`
	HasErrors  bool      `json:"hasErrors"`
	KillSwitch string    `json:"killSwitch"` // "na", "on" or "off"
	PublicKey  string    `json:"publicKey"`
}

func kindOf(t highlight) string {
	switch t {
	case highlightSection:
		return "sec"
	case highlightField:
		return "key"
	case highlightPrivateKey, highlightPublicKey, highlightPresharedKey, highlightHeaderProtectionKey:
		return "secret"
	case highlightIP, highlightHost:
		return "addr"
	case highlightCidr, highlightPort, highlightMTU, highlightKeepalive, highlightTable,
		highlightJc, highlightJmin, highlightJmax, highlightS1, highlightS2, highlightS3, highlightS4,
		highlightRekeyAfterTime, highlightRekeyTimeout, highlightRejectAfterTime, highlightKeepaliveTimeout,
		highlightMaxHandshakeAttempts:
		return "num"
	case highlightComment:
		return "cmt"
	case highlightDelimiter:
		return "eq"
	case highlightCmd:
		return "cmd"
	case highlightWarning:
		return "warn"
	case highlightError:
		return "err"
	}
	return "val"
}

// Analyze highlights and validates a configuration using the same rules as
// the original syntax editor.
func Analyze(cfg string) Result {
	spans := highlightConfig(cfg)
	highlightASecConfig(cfg, spans)

	res := Result{KillSwitch: "na"}
	lineStarts := []int{0}
	for i := 0; i < len(cfg); i++ {
		if cfg[i] == '\n' {
			lineStarts = append(lineStarts, i+1)
		}
	}
	lineOf := func(off int) int {
		lo, hi := 0, len(lineStarts)-1
		for lo < hi {
			mid := (lo + hi + 1) / 2
			if lineStarts[mid] <= off {
				lo = mid
			} else {
				hi = mid - 1
			}
		}
		return lo + 1
	}

	pos := 0
	seenLines := map[int]bool{}
	for i := range spans {
		span := &spans[i]
		if span.s < pos || span.s+span.len > len(cfg) {
			continue
		}
		if span.s > pos {
			res.Tokens = append(res.Tokens, Token{Kind: "plain", Text: cfg[pos:span.s]})
		}
		text := cfg[span.s : span.s+span.len]
		kind := kindOf(span.t)
		res.Tokens = append(res.Tokens, Token{Kind: kind, Text: text})
		pos = span.s + span.len

		if kind == "err" || kind == "warn" {
			line := lineOf(span.s)
			if kind == "err" {
				res.HasErrors = true
			}
			if seenLines[line] {
				continue
			}
			seenLines[line] = true
			p := Problem{Line: line, Text: strings.TrimSpace(text), Warning: kind == "warn"}
			switch {
			case kind == "warn":
				p.Kind = "warning"
			case strings.HasPrefix(strings.TrimSpace(text), "["):
				p.Kind = "badSection"
			case i+1 < len(spans) && spans[i+1].t == highlightDelimiter && lineOf(spans[i+1].s) == line:
				p.Kind = "unknownKey"
			case strings.TrimSpace(cfg[lineStarts[line-1]:span.s]) == "":
				p.Kind = "unknownKey"
			default:
				p.Kind = "badValue"
			}
			res.Problems = append(res.Problems, p)
		}

		if span.t == highlightPrivateKey && res.PublicKey == "" {
			if key, err := conf.NewPrivateKeyFromString(text); err == nil && key != nil {
				res.PublicKey = key.Public().String()
			}
		}
	}
	if pos < len(cfg) {
		res.Tokens = append(res.Tokens, Token{Kind: "plain", Text: cfg[pos:]})
	}
	res.KillSwitch = evaluateUntunneledBlocking(cfg, spans)
	return res
}

func evaluateUntunneledBlocking(cfg string, spans []highlightSpan) string {
	var onAllowedIPs, onTable, tableOff, seenPeer,
		seen00v6, seen00v4, seen01v6, seen80001v6, seen01v4, seen1281v4 bool

	for i := range spans {
		span := &spans[i]
		switch span.t {
		case highlightError:
			return "na"
		case highlightSection:
			if !strings.EqualFold(cfg[span.s:span.s+span.len], "[Peer]") {
				break
			}
			if seenPeer {
				return "na"
			}
			seenPeer = true
		case highlightField:
			onAllowedIPs = strings.EqualFold(cfg[span.s:span.s+span.len], "AllowedIPs")
			onTable = strings.EqualFold(cfg[span.s:span.s+span.len], "Table")
		case highlightTable:
			if onTable {
				tableOff = cfg[span.s:span.s+span.len] == "off"
			}
		case highlightIP:
			if !onAllowedIPs || !seenPeer {
				break
			}
			if i+2 >= len(spans) || spans[i+1].t != highlightDelimiter || spans[i+2].t != highlightCidr {
				break
			}
			if spans[i+2].len != 1 {
				break
			}
			switch cfg[spans[i+2].s] {
			case '0':
				switch cfg[span.s : span.s+span.len] {
				case "0.0.0.0":
					seen00v4 = true
				case "::":
					seen00v6 = true
				}
			case '1':
				switch cfg[span.s : span.s+span.len] {
				case "0.0.0.0":
					seen01v4 = true
				case "128.0.0.0":
					seen1281v4 = true
				case "::":
					seen01v6 = true
				case "8000::":
					seen80001v6 = true
				}
			}
		}
	}
	if tableOff {
		return "na"
	}
	if seen00v4 || seen00v6 {
		return "on"
	}
	if (seen01v4 && seen1281v4) || (seen01v6 && seen80001v6) {
		return "off"
	}
	return "na"
}

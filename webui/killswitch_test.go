package webui

import (
	"strings"
	"testing"
)

const ksSample = "[Interface]\nPrivateKey = yA8kN3pQ7wE2rT5uI9oP1aS4dF6gH0jK2lZ5xC8vB3k=\nAddress = 10.8.0.12/32\n\n[Peer]\nPublicKey = Hk2LpZ8nQw5rT1yUe7oI3aSdF9gJcVbX4mN0tR6uWqE=\nAllowedIPs = 0.0.0.0/0, ::/0, 10.9.0.0/16\nEndpoint = 203.0.113.24:51820\n"

func TestKillSwitchRoundTrip(t *testing.T) {
	off, err := setKillSwitch(ksSample, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(off, "0.0.0.0/1") || !strings.Contains(off, "128.0.0.0/1") || strings.Contains(off, "0.0.0.0/0") || !strings.Contains(off, "10.9.0.0/16") {
		t.Fatalf("unexpected off config:\n%s", off)
	}
	on, err := setKillSwitch(off, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(on, "0.0.0.0/0") || !strings.Contains(on, "::/0") || strings.Contains(on, "/1,") {
		t.Fatalf("unexpected on config:\n%s", on)
	}
	if _, err := setKillSwitch(strings.Replace(ksSample, "0.0.0.0/0, ::/0, ", "", 1), true); err == nil {
		t.Fatal("expected not applicable for split tunnel")
	}
}

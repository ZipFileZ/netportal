package syntax

import "testing"

const sample = "[Interface]\nPrivateKey = yA8kN3pQ7wE2rT5uI9oP1aS4dF6gH0jK2lZ5xC8vB3k=\nAdress = 10.8.0.12/32\nMTU = 1280\nJc = 4\n\n[Peer]\nPublicKey = Hk2LpZ8nQw5rT1yUe7oI3aSdF9gJcVbX4mN0tR6uWqE=\nAllowedIPs = 0.0.0.0/0, ::/0\nEndpoint = 203.0.113.24:51820\n"

func TestAnalyze(t *testing.T) {
	r := Analyze(sample)
	text := ""
	for _, tk := range r.Tokens {
		text += tk.Text
	}
	if text != sample {
		t.Fatalf("tokens do not rebuild input:\n%q", text)
	}
	if !r.HasErrors || len(r.Problems) != 1 || r.Problems[0].Line != 3 || r.Problems[0].Kind != "unknownKey" {
		t.Fatalf("unexpected problems: %+v", r.Problems)
	}
	if r.PublicKey == "" {
		t.Fatal("no public key")
	}
	ok := Analyze(sample[:0] + "[Interface]\nPrivateKey = yA8kN3pQ7wE2rT5uI9oP1aS4dF6gH0jK2lZ5xC8vB3k=\nAddress = 10.8.0.12/32\n\n[Peer]\nPublicKey = Hk2LpZ8nQw5rT1yUe7oI3aSdF9gJcVbX4mN0tR6uWqE=\nAllowedIPs = 0.0.0.0/0, ::/0\nEndpoint = 203.0.113.24:51820\n")
	if ok.HasErrors || ok.KillSwitch != "on" {
		t.Fatalf("expected valid config with kill-switch on: %+v %s", ok.Problems, ok.KillSwitch)
	}
}

package webui

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/amnezia-vpn/amneziawg-windows-client/manager"
)

func newTestApp(multiple bool) *app {
	a := &app{
		be:         newDemoBackend(),
		demo:       true,
		isAdmin:    true,
		settings:   newSettingsStore(false),
		states:     map[string]manager.TunnelState{},
		lastErrors: map[string]string{},
	}
	s := a.settings.get()
	s.MultipleTunnels = multiple
	a.settings.set(s)
	return a
}

func waitState(t *testing.T, a *app, name string, want manager.TunnelState) {
	t.Helper()
	for i := 0; i < 50; i++ {
		if st, _ := a.be.State(name); st == want {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	st, _ := a.be.State(name)
	t.Fatalf("%s: state %v, want %v", name, st, want)
}

func TestOneTunnelAtATime(t *testing.T) {
	// ams-01 (full route) is up; office-msk only routes office networks, so
	// the manager alone would keep both running.
	a := newTestApp(false)
	if got := a.conflictingTunnels("office-msk"); len(got) != 1 || got[0] != "ams-01" {
		t.Fatalf("conflicts = %v, want [ams-01]", got)
	}
	args, _ := json.Marshal(nameArg{Name: "office-msk"})
	if _, err := a.rpcStart(nil, args); err != nil {
		t.Fatal(err)
	}
	waitState(t, a, "ams-01", manager.TunnelStopped)
	waitState(t, a, "office-msk", manager.TunnelStarted)
}

func TestSeveralTunnelsAllowed(t *testing.T) {
	a := newTestApp(true)
	if got := a.conflictingTunnels("office-msk"); len(got) != 0 {
		t.Fatalf("conflicts = %v, want none", got)
	}
	args, _ := json.Marshal(nameArg{Name: "office-msk"})
	if _, err := a.rpcStart(nil, args); err != nil {
		t.Fatal(err)
	}
	waitState(t, a, "office-msk", manager.TunnelStarted)
	waitState(t, a, "ams-01", manager.TunnelStarted)
	// A second full-route tunnel still replaces ams-01, as the manager does.
	if got := a.conflictingTunnels("tokyo-02"); len(got) != 1 || got[0] != "ams-01" {
		t.Fatalf("conflicts = %v, want [ams-01]", got)
	}
}

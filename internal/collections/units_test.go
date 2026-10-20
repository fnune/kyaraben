package collections

import (
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/sync"
)

func TestUnitsEnsureAndRemove(t *testing.T) {
	env := newTestEnv(t, map[string]any{})
	svc := sync.NewFakeServiceManager()
	units := NewUnits(env.fs, svc, "/systemd", "kyaraben-collections", "")

	if err := units.Ensure("/home/u/.local/bin/kyaraben", 47410); err != nil {
		t.Fatal(err)
	}
	socket, err := env.fs.ReadFile("/systemd/kyaraben-collections.socket")
	if err != nil || !strings.Contains(string(socket), "ListenStream=127.0.0.1:47410") {
		t.Errorf("socket unit: %v %s", err, socket)
	}
	service, _ := env.fs.ReadFile("/systemd/kyaraben-collections.service")
	if !strings.Contains(string(service), "ExecStart=/home/u/.local/bin/kyaraben collections serve --port=47410") {
		t.Errorf("service unit: %s", service)
	}
	if !svc.IsEnabled("kyaraben-collections.socket") {
		t.Error("socket should be enabled")
	}

	if err := units.Remove(); err != nil {
		t.Fatal(err)
	}
	if units.Installed() {
		t.Error("units should be removed")
	}
}

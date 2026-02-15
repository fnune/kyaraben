package doctor

import (
	"context"
	"testing"

	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
	"github.com/fnune/kyaraben/internal/testutil"
)

func mustNewUserStore(t *testing.T, fs vfs.FS, path string) *store.UserStore {
	t.Helper()
	s, err := store.NewUserStore(fs, path)
	if err != nil {
		t.Fatalf("NewUserStore(%q) failed: %v", path, err)
	}
	return s
}

func TestRun(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation/bios/psx": &vfst.Dir{Perm: 0755},
	})

	userStorePath := "/Emulation"
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation},
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, userStorePath)

	result, err := Run(context.Background(), cfg, reg, userStore)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.Systems) != 2 {
		t.Errorf("Systems: got %d, want 2", len(result.Systems))
	}

	if result.UnsatisfiedGroups == 0 {
		t.Error("UnsatisfiedGroups should be > 0 (PSX BIOS group unsatisfied)")
	}

	if !result.HasIssues() {
		t.Error("HasIssues should return true when required groups are unsatisfied")
	}
}

func TestRunNoRequiredProvisions(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation": &vfst.Dir{Perm: 0755},
	})

	userStorePath := "/Emulation"
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDGBA: {model.EmulatorIDMGBA},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, userStorePath)

	result, err := Run(context.Background(), cfg, reg, userStore)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.Systems) != 1 {
		t.Fatalf("Systems: got %d, want 1", len(result.Systems))
	}

	sys := result.Systems[0]
	if sys.SystemID != model.SystemIDGBA {
		t.Errorf("SystemID: got %s, want %s", sys.SystemID, model.SystemIDGBA)
	}
	if sys.EmulatorName != "mGBA" {
		t.Errorf("EmulatorName: got %s, want mGBA", sys.EmulatorName)
	}

	if result.UnsatisfiedGroups != 0 {
		t.Errorf("UnsatisfiedGroups: got %d, want 0 (GBA BIOS is optional)", result.UnsatisfiedGroups)
	}
	if result.HasIssues() {
		t.Error("HasIssues should return false when no required groups are unsatisfied")
	}
}

func TestRunWithBiosFile(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation/bios/psx":              &vfst.Dir{Perm: 0755},
		"/Emulation/bios/psx/scph5501.bin": "fake bios content",
	})

	userStorePath := "/Emulation"
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, userStorePath)

	result, err := Run(context.Background(), cfg, reg, userStore)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.Systems) != 1 {
		t.Fatalf("Systems: got %d, want 1", len(result.Systems))
	}

	sys := result.Systems[0]

	var foundProv *ProvisionResult
	for i := range sys.Provisions {
		if sys.Provisions[i].Filename == "scph5501.bin" {
			foundProv = &sys.Provisions[i]
			break
		}
	}

	if foundProv == nil {
		t.Fatal("scph5501.bin provision not found in results")
		return
	}

	if foundProv.Status != model.ProvisionInvalid {
		t.Errorf("Provision status: got %s, want %s (fake content has wrong hash)", foundProv.Status, model.ProvisionInvalid)
	}
}

func TestRunSystemResult(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation/bios/psx": &vfst.Dir{Perm: 0755},
	})

	userStorePath := "/Emulation"
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, userStorePath)

	result, err := Run(context.Background(), cfg, reg, userStore)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	sys := result.Systems[0]

	if sys.SystemID != model.SystemIDPSX {
		t.Errorf("SystemID: got %s, want %s", sys.SystemID, model.SystemIDPSX)
	}
	if sys.EmulatorID != model.EmulatorIDDuckStation {
		t.Errorf("EmulatorID: got %s, want %s", sys.EmulatorID, model.EmulatorIDDuckStation)
	}
	if sys.EmulatorName != "DuckStation" {
		t.Errorf("EmulatorName: got %s, want DuckStation", sys.EmulatorName)
	}

	if len(sys.Provisions) == 0 {
		t.Error("PSX should have provisions defined")
	}
}

func TestRunMultipleEmulators(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/Emulation/bios/psx": &vfst.Dir{Perm: 0755},
	})

	userStorePath := "/Emulation"
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID][]model.EmulatorID{
			model.SystemIDPSX: {model.EmulatorIDDuckStation, model.EmulatorIDRetroArchBeetleSaturn},
		},
	}

	reg := registry.NewDefault()
	userStore := mustNewUserStore(t, fs, userStorePath)

	result, err := Run(context.Background(), cfg, reg, userStore)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.Systems) != 2 {
		t.Errorf("Systems: got %d, want 2 (one per emulator)", len(result.Systems))
	}

	emulatorsSeen := make(map[model.EmulatorID]bool)
	for _, sys := range result.Systems {
		emulatorsSeen[sys.EmulatorID] = true
	}

	if !emulatorsSeen[model.EmulatorIDDuckStation] {
		t.Error("DuckStation result not found")
	}
	if !emulatorsSeen[model.EmulatorIDRetroArchBeetleSaturn] {
		t.Error("RetroArch Beetle Saturn result not found")
	}
}

func TestHasIssues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		unsatisfiedGroups int
		optionalMissed    int
		want              bool
	}{
		{"no issues", 0, 0, false},
		{"only optional missing", 0, 5, false},
		{"required unsatisfied", 1, 0, true},
		{"both missing", 2, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{
				UnsatisfiedGroups:    tt.unsatisfiedGroups,
				OptionalGroupsMissed: tt.optionalMissed,
			}
			if got := r.HasIssues(); got != tt.want {
				t.Errorf("HasIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}

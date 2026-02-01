package doctor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

func mustNewUserStore(t *testing.T, path string) *store.UserStore {
	t.Helper()
	s, err := store.NewUserStore(path)
	if err != nil {
		t.Fatalf("NewUserStore(%q) failed: %v", path, err)
	}
	return s
}

func TestRun(t *testing.T) {
	tmpDir := t.TempDir()
	userStorePath := filepath.Join(tmpDir, "Emulation")

	if err := os.MkdirAll(filepath.Join(userStorePath, "bios", "psx"), 0755); err != nil {
		t.Fatalf("Failed to create bios dir: %v", err)
	}
	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDPSX: {Emulator: string(model.EmulatorIDDuckStation)},
			model.SystemIDGBA: {Emulator: string(model.EmulatorIDMGBA)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, userStorePath)

	result, err := Run(context.Background(), cfg, registry, userStore)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.Systems) != 2 {
		t.Errorf("Systems: got %d, want 2", len(result.Systems))
	}

	// PSX should have missing required BIOS
	if result.RequiredMissing == 0 {
		t.Error("RequiredMissing should be > 0 (PSX BIOS missing)")
	}

	if !result.HasIssues() {
		t.Error("HasIssues should return true when required files are missing")
	}
}

func TestRunNoRequiredProvisions(t *testing.T) {
	tmpDir := t.TempDir()
	userStorePath := filepath.Join(tmpDir, "Emulation")

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDGBA: {Emulator: string(model.EmulatorIDMGBA)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, userStorePath)

	result, err := Run(context.Background(), cfg, registry, userStore)
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

	if result.RequiredMissing != 0 {
		t.Errorf("RequiredMissing: got %d, want 0 (GBA BIOS is optional)", result.RequiredMissing)
	}
	if result.HasIssues() {
		t.Error("HasIssues should return false when no required files are missing")
	}
}

func TestRunWithBiosFile(t *testing.T) {
	tmpDir := t.TempDir()
	userStorePath := filepath.Join(tmpDir, "Emulation")

	biosDir := filepath.Join(userStorePath, "bios", "psx")
	if err := os.MkdirAll(biosDir, 0755); err != nil {
		t.Fatalf("Failed to create bios dir: %v", err)
	}

	// Create a fake BIOS file (with wrong hash, but let's test the flow)
	biosFile := filepath.Join(biosDir, "scph5501.bin")
	if err := os.WriteFile(biosFile, []byte("fake bios content"), 0644); err != nil {
		t.Fatalf("Failed to create bios file: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDPSX: {Emulator: string(model.EmulatorIDDuckStation)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, userStorePath)

	result, err := Run(context.Background(), cfg, registry, userStore)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.Systems) != 1 {
		t.Fatalf("Systems: got %d, want 1", len(result.Systems))
	}

	sys := result.Systems[0]

	// Find the scph5501.bin provision result
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

	// File exists but hash is wrong, so should be Invalid
	if foundProv.Status != model.ProvisionInvalid {
		t.Errorf("Provision status: got %s, want %s", foundProv.Status, model.ProvisionInvalid)
	}
}

func TestRunSystemResult(t *testing.T) {
	tmpDir := t.TempDir()
	userStorePath := filepath.Join(tmpDir, "Emulation")

	if err := os.MkdirAll(filepath.Join(userStorePath, "bios", "psx"), 0755); err != nil {
		t.Fatalf("Failed to create bios dir: %v", err)
	}

	cfg := &model.KyarabenConfig{
		Global: model.GlobalConfig{
			UserStore: userStorePath,
		},
		Systems: map[model.SystemID]model.SystemConf{
			model.SystemIDPSX: {Emulator: string(model.EmulatorIDDuckStation)},
		},
	}

	registry := registry.NewDefault()
	userStore := mustNewUserStore(t, userStorePath)

	result, err := Run(context.Background(), cfg, registry, userStore)
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

	expectedBiosDir := filepath.Join(userStorePath, "bios", "psx")
	if sys.BiosDir != expectedBiosDir {
		t.Errorf("BiosDir: got %s, want %s", sys.BiosDir, expectedBiosDir)
	}

	// PSX should have provisions
	if len(sys.Provisions) == 0 {
		t.Error("PSX should have provisions defined")
	}
}

func TestHasIssues(t *testing.T) {
	tests := []struct {
		name            string
		requiredMissing int
		optionalMissing int
		want            bool
	}{
		{"no issues", 0, 0, false},
		{"only optional missing", 0, 5, false},
		{"required missing", 1, 0, true},
		{"both missing", 2, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{
				RequiredMissing: tt.requiredMissing,
				OptionalMissing: tt.optionalMissing,
			}
			if got := r.HasIssues(); got != tt.want {
				t.Errorf("HasIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}

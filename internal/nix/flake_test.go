package nix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/versions"
)

func TestMain(m *testing.M) {
	if err := versions.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestFlakeGeneratorGenerateAllEmulators(t *testing.T) {
	tmpDir := t.TempDir()
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	allEmulators := reg.AllEmulators()
	emulatorIDs := make([]model.EmulatorID, len(allEmulators))
	for i, emu := range allEmulators {
		emulatorIDs[i] = emu.ID
	}

	result, err := fg.Generate(tmpDir, emulatorIDs, nil)
	if err != nil {
		t.Fatalf("Generate failed for all emulators: %v", err)
	}

	if len(result.SkippedEmulators) > 0 {
		t.Errorf("Expected no skipped emulators, got: %v", result.SkippedEmulators)
	}

	flakePath := filepath.Join(string(result.Path), "flake.nix")
	if _, err := os.Stat(flakePath); os.IsNotExist(err) {
		t.Fatal("flake.nix was not created")
	}

	content, err := os.ReadFile(flakePath)
	if err != nil {
		t.Fatalf("failed to read flake.nix: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "nixpkgs.url") {
		t.Error("flake.nix should contain nixpkgs input")
	}
	if !strings.Contains(contentStr, "kyaraben-profile") {
		t.Error("flake.nix should contain profile derivation")
	}
}

func TestFlakeGeneratorGenerateSingleEmulator(t *testing.T) {
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	for _, emu := range reg.AllEmulators() {
		t.Run(string(emu.ID), func(t *testing.T) {
			tmpDir := t.TempDir()

			result, err := fg.Generate(tmpDir, []model.EmulatorID{emu.ID}, nil)
			if err != nil {
				t.Fatalf("Generate failed for %s: %v", emu.ID, err)
			}

			content, err := os.ReadFile(filepath.Join(string(result.Path), "flake.nix"))
			if err != nil {
				t.Fatalf("failed to read flake.nix: %v", err)
			}

			if len(content) == 0 {
				t.Error("flake.nix is empty")
			}
		})
	}
}

func TestFlakeGeneratorGenerateUnknownEmulator(t *testing.T) {
	tmpDir := t.TempDir()
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	result, err := fg.Generate(tmpDir, []model.EmulatorID{"unknown-emulator"}, nil)
	if err != nil {
		t.Fatalf("Generate should not fail for unknown emulator: %v", err)
	}

	if len(result.SkippedEmulators) != 1 {
		t.Fatalf("expected 1 skipped emulator, got %d", len(result.SkippedEmulators))
	}

	if result.SkippedEmulators[0] != "unknown-emulator" {
		t.Errorf("expected skipped emulator 'unknown-emulator', got %s", result.SkippedEmulators[0])
	}
}

func TestFlakeGeneratorGenerateMixedEmulators(t *testing.T) {
	tmpDir := t.TempDir()
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	result, err := fg.Generate(tmpDir, []model.EmulatorID{model.EmulatorIDMGBA, "unknown-emulator", model.EmulatorIDDolphin}, nil)
	if err != nil {
		t.Fatalf("Generate should not fail: %v", err)
	}

	if len(result.SkippedEmulators) != 1 {
		t.Fatalf("expected 1 skipped emulator, got %d", len(result.SkippedEmulators))
	}

	if result.SkippedEmulators[0] != "unknown-emulator" {
		t.Errorf("expected skipped emulator 'unknown-emulator', got %s", result.SkippedEmulators[0])
	}

	content, err := os.ReadFile(filepath.Join(string(result.Path), "flake.nix"))
	if err != nil {
		t.Fatalf("failed to read flake.nix: %v", err)
	}

	if !strings.Contains(string(content), "mgba") {
		t.Error("flake.nix should contain mgba package")
	}
	if !strings.Contains(string(content), "dolphin") {
		t.Error("flake.nix should contain dolphin package")
	}
}

func TestFlakeGeneratorFlakeRef(t *testing.T) {
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	for _, emu := range reg.AllEmulators() {
		t.Run(string(emu.ID), func(t *testing.T) {
			ref, err := fg.FlakeRef("/tmp/flake", emu.ID)
			if err != nil {
				t.Errorf("FlakeRef(%s) returned error: %v", emu.ID, err)
				return
			}
			if !strings.HasPrefix(ref, "/") || !strings.Contains(ref, "#") {
				t.Errorf("FlakeRef(%s) = %s, expected path#package format", emu.ID, ref)
			}
		})
	}
}

func TestFlakeGeneratorDefaultFlakeRef(t *testing.T) {
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	ref := fg.DefaultFlakeRef("/tmp/flake")
	absPath, _ := filepath.Abs("/tmp/flake")

	if ref != absPath {
		t.Errorf("DefaultFlakeRef() = %s, expected %s", ref, absPath)
	}
}

func TestFlakeGeneratorCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "flake", "dir")
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	result, err := fg.Generate(nestedDir, []model.EmulatorID{model.EmulatorIDMGBA}, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if _, err := os.Stat(string(result.Path)); os.IsNotExist(err) {
		t.Error("generation directory should have been created")
	}

	if !strings.Contains(string(result.Path), "generations") {
		t.Error("generation path should contain 'generations' directory")
	}
}

func TestNewFlakeGenerator(t *testing.T) {
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	if fg == nil || fg.emulators == nil {
		t.Fatal("NewFlakeGenerator should return initialized generator")
	}
}

func TestPackageInfoFromRef(t *testing.T) {
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg, reg)

	tests := []struct {
		name     string
		ref      model.PackageRef
		wantName string
		wantErr  bool
	}{
		{
			name:     "appimage",
			ref:      model.AppImageRef("eden"),
			wantName: "eden",
		},
		{
			name:     "appimage mgba",
			ref:      model.AppImageRef("mgba"),
			wantName: "mgba",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := fg.packageInfoFromRef(tt.ref)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", info.Name, tt.wantName)
			}
		})
	}
}

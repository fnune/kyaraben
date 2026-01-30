package nix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
)

func TestFlakeGeneratorGenerateAllEmulators(t *testing.T) {
	tmpDir := t.TempDir()
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg)

	allEmulators := reg.AllEmulators()
	emulatorIDs := make([]model.EmulatorID, len(allEmulators))
	for i, emu := range allEmulators {
		emulatorIDs[i] = emu.ID
	}

	err := fg.Generate(tmpDir, emulatorIDs)
	if err != nil {
		t.Fatalf("Generate failed for all emulators: %v", err)
	}

	flakePath := filepath.Join(tmpDir, "flake.nix")
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
	if !strings.Contains(contentStr, "kyaraben-emulators") {
		t.Error("flake.nix should contain combined environment")
	}
}

func TestFlakeGeneratorGenerateSingleEmulator(t *testing.T) {
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg)

	for _, emu := range reg.AllEmulators() {
		t.Run(string(emu.ID), func(t *testing.T) {
			tmpDir := t.TempDir()

			err := fg.Generate(tmpDir, []model.EmulatorID{emu.ID})
			if err != nil {
				t.Fatalf("Generate failed for %s: %v", emu.ID, err)
			}

			content, err := os.ReadFile(filepath.Join(tmpDir, "flake.nix"))
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
	fg := NewFlakeGenerator(reg)

	err := fg.Generate(tmpDir, []model.EmulatorID{"unknown-emulator"})
	if err == nil {
		t.Fatal("expected error for unknown emulator")
	}

	if !strings.Contains(err.Error(), "unknown emulator") {
		t.Errorf("error should mention unknown emulator, got: %v", err)
	}
}

func TestFlakeGeneratorFlakeRef(t *testing.T) {
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg)

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
	fg := NewFlakeGenerator(reg)

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
	fg := NewFlakeGenerator(reg)

	// Use e2e-test emulator which should always exist
	err := fg.Generate(nestedDir, []model.EmulatorID{model.EmulatorE2ETest})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("nested directory should have been created")
	}
}

func TestNewFlakeGenerator(t *testing.T) {
	reg := registry.NewDefault()
	fg := NewFlakeGenerator(reg)

	if fg == nil || fg.emulators == nil {
		t.Fatal("NewFlakeGenerator should return initialized generator")
	}
}

func TestPackageInfoFromRef(t *testing.T) {
	tests := []struct {
		name     string
		ref      model.PackageRef
		wantName string
		wantExpr string
		wantErr  bool
	}{
		{
			name:     "simple nixpkgs",
			ref:      model.NixpkgsRef("duckstation"),
			wantName: "duckstation",
			wantExpr: "pkgs.duckstation",
		},
		{
			name:     "nixpkgs with overlay",
			ref:      model.NixpkgsOverlayRef("retroarch-bsnes", "pkgs.wrapRetroArch {}"),
			wantName: "retroarch-bsnes",
			wantExpr: "pkgs.wrapRetroArch {}",
		},
		{
			name: "github appimage",
			ref: model.GitHubAppImageRef(
				"eden", "owner", "repo", "v1.0",
				map[string]string{"x86_64": "asset-x64.AppImage", "aarch64": "asset-arm64.AppImage"},
				map[string]string{"x86_64": "abc123", "aarch64": "def456"},
			),
			wantName: "eden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := packageInfoFromRef(tt.ref)
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
			if tt.wantExpr != "" && info.Expr != tt.wantExpr {
				t.Errorf("Expr = %s, want %s", info.Expr, tt.wantExpr)
			}
		})
	}
}

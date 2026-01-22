package nix

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

// mockEmulatorLookup implements EmulatorLookup for testing.
type mockEmulatorLookup struct {
	emulators map[model.EmulatorID]model.Emulator
}

func newMockLookup() *mockEmulatorLookup {
	return &mockEmulatorLookup{
		emulators: map[model.EmulatorID]model.Emulator{
			model.EmulatorDuckStation: {
				ID:      model.EmulatorDuckStation,
				Name:    "DuckStation",
				Package: model.NixpkgsRef("duckstation"),
			},
			model.EmulatorRetroArchBsnes: {
				ID:      model.EmulatorRetroArchBsnes,
				Name:    "RetroArch (bsnes)",
				Package: model.NixpkgsOverlayRef("retroarch-bsnes", `pkgs.retroarch.override { cores = with pkgs.libretro; [ bsnes ]; }`),
			},
			model.EmulatorTIC80: {
				ID:      model.EmulatorTIC80,
				Name:    "TIC-80",
				Package: model.NixpkgsRef("tic-80"),
			},
			model.EmulatorE2ETest: {
				ID:      model.EmulatorE2ETest,
				Name:    "E2E Test",
				Package: model.NixpkgsRef("hello"),
			},
		},
	}
}

func (m *mockEmulatorLookup) GetEmulator(id model.EmulatorID) (model.Emulator, error) {
	emu, ok := m.emulators[id]
	if !ok {
		return model.Emulator{}, fmt.Errorf("unknown emulator: %s", id)
	}
	return emu, nil
}

func TestFlakeGeneratorGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	fg := NewFlakeGenerator(newMockLookup())

	err := fg.Generate(tmpDir, []model.EmulatorID{model.EmulatorDuckStation})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	flakePath := filepath.Join(tmpDir, "flake.nix")
	if _, err := os.Stat(flakePath); os.IsNotExist(err) {
		t.Fatal("flake.nix was not created")
	}

	content, err := os.ReadFile(flakePath)
	if err != nil {
		t.Fatalf("failed to read flake.nix: %v", err)
	}

	// Check that the flake contains expected content
	if !strings.Contains(string(content), "duckstation") {
		t.Error("flake.nix should contain duckstation package")
	}
	if !strings.Contains(string(content), "nixpkgs.url") {
		t.Error("flake.nix should contain nixpkgs input")
	}
	if !strings.Contains(string(content), "kyaraben-emulators") {
		t.Error("flake.nix should contain combined environment")
	}
}

func TestFlakeGeneratorGenerateMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	fg := NewFlakeGenerator(newMockLookup())

	emulators := []model.EmulatorID{
		model.EmulatorRetroArchBsnes,
		model.EmulatorDuckStation,
		model.EmulatorTIC80,
	}

	err := fg.Generate(tmpDir, emulators)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "flake.nix"))
	if err != nil {
		t.Fatalf("failed to read flake.nix: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "retroarch-bsnes") {
		t.Error("flake.nix should contain retroarch-bsnes package")
	}
	if !strings.Contains(contentStr, "duckstation") {
		t.Error("flake.nix should contain duckstation package")
	}
	if !strings.Contains(contentStr, "tic-80") {
		t.Error("flake.nix should contain tic-80 package")
	}
}

func TestFlakeGeneratorGenerateUnknownEmulator(t *testing.T) {
	tmpDir := t.TempDir()
	fg := NewFlakeGenerator(newMockLookup())

	err := fg.Generate(tmpDir, []model.EmulatorID{"unknown-emulator"})
	if err == nil {
		t.Fatal("expected error for unknown emulator")
	}

	if !strings.Contains(err.Error(), "unknown emulator") {
		t.Errorf("error should mention unknown emulator, got: %v", err)
	}
}

func TestFlakeGeneratorFlakeRef(t *testing.T) {
	fg := NewFlakeGenerator(newMockLookup())

	tests := []struct {
		emuID    model.EmulatorID
		expected string
	}{
		{model.EmulatorDuckStation, "duckstation"},
		{model.EmulatorRetroArchBsnes, "retroarch-bsnes"},
		{model.EmulatorTIC80, "tic-80"},
	}

	for _, test := range tests {
		ref, err := fg.FlakeRef("/tmp/flake", test.emuID)
		if err != nil {
			t.Errorf("FlakeRef(%s) returned error: %v", test.emuID, err)
			continue
		}
		if !strings.HasSuffix(ref, "#"+test.expected) {
			t.Errorf("FlakeRef(%s) = %s, expected suffix #%s", test.emuID, ref, test.expected)
		}
	}
}

func TestFlakeGeneratorDefaultFlakeRef(t *testing.T) {
	fg := NewFlakeGenerator(newMockLookup())

	ref := fg.DefaultFlakeRef("/tmp/flake")
	absPath, _ := filepath.Abs("/tmp/flake")

	if ref != absPath {
		t.Errorf("DefaultFlakeRef() = %s, expected %s", ref, absPath)
	}
}

func TestFlakeGeneratorCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "flake", "dir")
	fg := NewFlakeGenerator(newMockLookup())

	err := fg.Generate(nestedDir, []model.EmulatorID{model.EmulatorE2ETest})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("nested directory should have been created")
	}
}

func TestFlakeGeneratorRetroArchOverride(t *testing.T) {
	tmpDir := t.TempDir()
	fg := NewFlakeGenerator(newMockLookup())

	err := fg.Generate(tmpDir, []model.EmulatorID{model.EmulatorRetroArchBsnes})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "flake.nix"))
	if err != nil {
		t.Fatalf("failed to read flake.nix: %v", err)
	}

	contentStr := string(content)

	// RetroArch should use override syntax
	if !strings.Contains(contentStr, "pkgs.retroarch.override") {
		t.Error("RetroArch should use override syntax for cores")
	}
	if !strings.Contains(contentStr, "bsnes") {
		t.Error("RetroArch override should include bsnes core")
	}
}

func TestNewFlakeGenerator(t *testing.T) {
	lookup := newMockLookup()
	fg := NewFlakeGenerator(lookup)

	if fg == nil {
		t.Fatal("NewFlakeGenerator returned nil")
	}

	if fg.emulators == nil {
		t.Fatal("emulators should be initialized")
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
			ref:      model.NixpkgsOverlayRef("retroarch-bsnes", "pkgs.retroarch.override {}"),
			wantName: "retroarch-bsnes",
			wantExpr: "pkgs.retroarch.override {}",
		},
		{
			name:    "github package",
			ref:     model.GitHubRef("owner", "repo", "asset"),
			wantErr: true,
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
			if info.Expr != tt.wantExpr {
				t.Errorf("Expr = %s, want %s", info.Expr, tt.wantExpr)
			}
		})
	}
}

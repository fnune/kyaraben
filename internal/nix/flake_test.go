package nix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestFlakeGeneratorGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	fg := NewFlakeGenerator()

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
	fg := NewFlakeGenerator()

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
	fg := NewFlakeGenerator()

	err := fg.Generate(tmpDir, []model.EmulatorID{"unknown-emulator"})
	if err == nil {
		t.Fatal("expected error for unknown emulator")
	}

	if !strings.Contains(err.Error(), "unknown emulator") {
		t.Errorf("error should mention unknown emulator, got: %v", err)
	}
}

func TestFlakeGeneratorFlakeRef(t *testing.T) {
	fg := NewFlakeGenerator()

	tests := []struct {
		emuID    model.EmulatorID
		expected string
	}{
		{model.EmulatorDuckStation, "duckstation"},
		{model.EmulatorRetroArchBsnes, "retroarch-bsnes"},
		{model.EmulatorTIC80, "tic-80"},
	}

	for _, test := range tests {
		ref := fg.FlakeRef("/tmp/flake", test.emuID)
		if !strings.HasSuffix(ref, "#"+test.expected) {
			t.Errorf("FlakeRef(%s) = %s, expected suffix #%s", test.emuID, ref, test.expected)
		}
	}
}

func TestFlakeGeneratorDefaultFlakeRef(t *testing.T) {
	fg := NewFlakeGenerator()

	ref := fg.DefaultFlakeRef("/tmp/flake")
	absPath, _ := filepath.Abs("/tmp/flake")

	if ref != absPath {
		t.Errorf("DefaultFlakeRef() = %s, expected %s", ref, absPath)
	}
}

func TestFlakeGeneratorCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "flake", "dir")
	fg := NewFlakeGenerator()

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
	fg := NewFlakeGenerator()

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
	fg := NewFlakeGenerator()

	if fg == nil {
		t.Fatal("NewFlakeGenerator returned nil")
	}

	if fg.emulatorAttrs == nil {
		t.Fatal("emulatorAttrs should be initialized")
	}

	expectedAttrs := []model.EmulatorID{
		model.EmulatorRetroArchBsnes,
		model.EmulatorDuckStation,
		model.EmulatorTIC80,
		model.EmulatorE2ETest,
	}

	for _, emuID := range expectedAttrs {
		if _, ok := fg.emulatorAttrs[emuID]; !ok {
			t.Errorf("emulatorAttrs missing entry for %s", emuID)
		}
	}
}

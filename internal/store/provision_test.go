package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestProvisionCheckerCheck(t *testing.T) {
	tmpDir := t.TempDir()
	store := mustNewUserStore(t, tmpDir)

	// Initialize the store
	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	if err := store.InitializeForEmulator(model.SystemIDPSX, model.EmulatorIDDuckStation, model.StandardPathUsage()); err != nil {
		t.Fatalf("Failed to initialize PSX system: %v", err)
	}

	checker := NewProvisionChecker(store)

	// Test emulator with provisions
	emu := model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Systems: []model.SystemID{model.SystemIDPSX},
		Provisions: []model.Provision{
			{
				ID:       "psx-bios-usa",
				Filename: "scph5501.bin",
				Required: true,
				MD5Hash:  "490f666e1afb15b7362b406ed1cea246",
			},
			{
				ID:       "psx-bios-japan",
				Filename: "scph5500.bin",
				Required: false,
				MD5Hash:  "8dd7d5296a650fac7319bce665a6a53c",
			},
		},
	}

	// Without any BIOS files
	results := checker.Check(emu, model.SystemIDPSX)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Required file should be missing
	if results[0].Status != model.ProvisionMissing {
		t.Errorf("Expected required provision to be missing, got %s", results[0].Status)
	}

	// Optional file should be marked as optional
	if results[1].Status != model.ProvisionOptional {
		t.Errorf("Expected optional provision to be optional, got %s", results[1].Status)
	}

	// HasMissingRequired should return true
	if !HasMissingRequired(results) {
		t.Error("HasMissingRequired should return true when required file is missing")
	}
}

func TestProvisionCheckerWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	store := mustNewUserStore(t, tmpDir)

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	if err := store.InitializeForEmulator(model.SystemIDGBA, model.EmulatorIDMGBA, model.StandardPathUsage()); err != nil {
		t.Fatalf("Failed to initialize GBA system: %v", err)
	}

	checker := NewProvisionChecker(store)

	emu := model.Emulator{
		ID:         model.EmulatorIDMGBA,
		Systems:    []model.SystemID{model.SystemIDGBA},
		Provisions: []model.Provision{},
	}

	results := checker.Check(emu, model.SystemIDGBA)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for emulator with no provisions, got %d", len(results))
	}

	if HasMissingRequired(results) {
		t.Error("HasMissingRequired should return false when there are no provisions")
	}
}

func TestProvisionCheckerHashVerification(t *testing.T) {
	tmpDir := t.TempDir()
	store := mustNewUserStore(t, tmpDir)

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	if err := store.InitializeForEmulator(model.SystemIDPSX, model.EmulatorIDDuckStation, model.StandardPathUsage()); err != nil {
		t.Fatalf("Failed to initialize PSX system: %v", err)
	}

	// Create a fake BIOS file with known content
	biosDir := store.SystemBiosDir(model.SystemIDPSX)
	biosFile := filepath.Join(biosDir, "test.bin")
	content := []byte("test content")
	if err := os.WriteFile(biosFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	// MD5 of "test content" is 9473fdd0d880a43c21b7778d34872157

	checker := NewProvisionChecker(store)

	// Test with correct hash
	emu := model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Systems: []model.SystemID{model.SystemIDPSX},
		Provisions: []model.Provision{
			{
				ID:       "test-bios",
				Filename: "test.bin",
				Required: true,
				MD5Hash:  "9473fdd0d880a43c21b7778d34872157",
			},
		},
	}

	results := checker.Check(emu, model.SystemIDPSX)
	if results[0].Status != model.ProvisionFound {
		t.Errorf("Expected provision to be found with correct hash, got %s", results[0].Status)
	}

	// Test with incorrect hash
	emu.Provisions[0].MD5Hash = "wronghash"
	results = checker.Check(emu, model.SystemIDPSX)
	if results[0].Status != model.ProvisionInvalid {
		t.Errorf("Expected provision to be invalid with wrong hash, got %s", results[0].Status)
	}
}

func TestProvisionCheckerCaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	store := mustNewUserStore(t, tmpDir)

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	if err := store.InitializeForEmulator(model.SystemIDPSX, model.EmulatorIDDuckStation, model.StandardPathUsage()); err != nil {
		t.Fatalf("Failed to initialize PSX system: %v", err)
	}

	// Create file with different case
	biosDir := store.SystemBiosDir(model.SystemIDPSX)
	biosFile := filepath.Join(biosDir, "SCPH5501.BIN")
	if err := os.WriteFile(biosFile, []byte("fake bios"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checker := NewProvisionChecker(store)

	emu := model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Systems: []model.SystemID{model.SystemIDPSX},
		Provisions: []model.Provision{
			{
				ID:       "psx-bios",
				Filename: "scph5501.bin", // lowercase
				Required: true,
				// No hash - just check existence
			},
		},
	}

	results := checker.Check(emu, model.SystemIDPSX)
	if results[0].Status != model.ProvisionFound {
		t.Errorf("Expected provision to be found (case insensitive), got %s", results[0].Status)
	}
}

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

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	if err := store.InitializeForEmulator(model.SystemIDPSX, model.EmulatorIDDuckStation, model.StandardPathUsage()); err != nil {
		t.Fatalf("Failed to initialize PSX system: %v", err)
	}

	checker := NewProvisionChecker(store)

	emu := model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Systems: []model.SystemID{model.SystemIDPSX},
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "At least one regional BIOS required",
			Provisions: []model.Provision{
				{
					Filename:    "scph5501.bin",
					Description: "USA",
					Hashes:      []string{"490f666e1afb15b7362b406ed1cea246"},
				},
				{
					Filename:    "scph5500.bin",
					Description: "Japan",
					Hashes:      []string{"8dd7d5296a650fac7319bce665a6a53c"},
				},
			},
		}},
	}

	results := checker.Check(emu, model.SystemIDPSX)
	if len(results) != 1 {
		t.Errorf("Expected 1 group result, got %d", len(results))
	}

	gr := results[0]
	if gr.IsSatisfied {
		t.Error("Group should not be satisfied when no files present")
	}
	if gr.Satisfied != 0 {
		t.Errorf("Expected 0 satisfied, got %d", gr.Satisfied)
	}

	for _, pr := range gr.Results {
		if pr.Status != model.ProvisionMissing {
			t.Errorf("Expected provision to be missing, got %s", pr.Status)
		}
	}

	if !HasUnsatisfiedRequired(results) {
		t.Error("HasUnsatisfiedRequired should return true when required group is unsatisfied")
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
		ID:              model.EmulatorIDMGBA,
		Systems:         []model.SystemID{model.SystemIDGBA},
		ProvisionGroups: nil,
	}

	results := checker.Check(emu, model.SystemIDGBA)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for emulator with no provisions, got %d", len(results))
	}

	if HasUnsatisfiedRequired(results) {
		t.Error("HasUnsatisfiedRequired should return false when there are no provisions")
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

	biosDir := store.SystemBiosDir(model.SystemIDPSX)
	biosFile := filepath.Join(biosDir, "test.bin")
	content := []byte("test content")
	if err := os.WriteFile(biosFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checker := NewProvisionChecker(store)

	emu := model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Systems: []model.SystemID{model.SystemIDPSX},
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "Test BIOS required",
			Provisions: []model.Provision{{
				Filename: "test.bin",
				Hashes:   []string{"9473fdd0d880a43c21b7778d34872157"},
			}},
		}},
	}

	results := checker.Check(emu, model.SystemIDPSX)
	if results[0].Results[0].Status != model.ProvisionFound {
		t.Errorf("Expected provision to be found with correct hash, got %s", results[0].Results[0].Status)
	}
	if !results[0].IsSatisfied {
		t.Error("Group should be satisfied when file is found")
	}

	emu.ProvisionGroups[0].Provisions[0].Hashes = []string{"wronghash"}
	results = checker.Check(emu, model.SystemIDPSX)
	if results[0].Results[0].Status != model.ProvisionInvalid {
		t.Errorf("Expected provision to be invalid with wrong hash, got %s", results[0].Results[0].Status)
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

	biosDir := store.SystemBiosDir(model.SystemIDPSX)
	biosFile := filepath.Join(biosDir, "SCPH5501.BIN")
	if err := os.WriteFile(biosFile, []byte("fake bios"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checker := NewProvisionChecker(store)

	emu := model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Systems: []model.SystemID{model.SystemIDPSX},
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Provisions: []model.Provision{{
				Filename: "scph5501.bin",
			}},
		}},
	}

	results := checker.Check(emu, model.SystemIDPSX)
	if results[0].Results[0].Status != model.ProvisionFound {
		t.Errorf("Expected provision to be found (case insensitive), got %s", results[0].Results[0].Status)
	}
}

func TestProvisionGroupSatisfaction(t *testing.T) {
	tmpDir := t.TempDir()
	store := mustNewUserStore(t, tmpDir)

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	if err := store.InitializeForEmulator(model.SystemIDPSX, model.EmulatorIDDuckStation, model.StandardPathUsage()); err != nil {
		t.Fatalf("Failed to initialize PSX system: %v", err)
	}

	biosDir := store.SystemBiosDir(model.SystemIDPSX)
	if err := os.WriteFile(filepath.Join(biosDir, "scph5501.bin"), []byte("usa"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checker := NewProvisionChecker(store)

	emu := model.Emulator{
		ID:      model.EmulatorIDDuckStation,
		Systems: []model.SystemID{model.SystemIDPSX},
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "At least one regional BIOS required",
			Provisions: []model.Provision{
				{Filename: "scph5501.bin", Description: "USA"},
				{Filename: "scph5500.bin", Description: "Japan"},
				{Filename: "scph5502.bin", Description: "Europe"},
			},
		}},
	}

	results := checker.Check(emu, model.SystemIDPSX)
	gr := results[0]

	if !gr.IsSatisfied {
		t.Error("Group should be satisfied when at least one file is present")
	}
	if gr.Satisfied != 1 {
		t.Errorf("Expected 1 satisfied, got %d", gr.Satisfied)
	}

	if HasUnsatisfiedRequired(results) {
		t.Error("HasUnsatisfiedRequired should return false when group is satisfied")
	}
}

func TestProvisionCheckerFilePattern(t *testing.T) {
	tmpDir := t.TempDir()
	store := mustNewUserStore(t, tmpDir)

	if err := store.Initialize(); err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	if err := store.InitializeForEmulator(model.SystemIDSwitch, model.EmulatorIDEden, model.StandardPathUsage()); err != nil {
		t.Fatalf("Failed to initialize Switch system: %v", err)
	}

	biosDir := store.SystemBiosDir(model.SystemIDSwitch)
	checker := NewProvisionChecker(store)

	emu := model.Emulator{
		ID:      model.EmulatorIDEden,
		Systems: []model.SystemID{model.SystemIDSwitch},
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 0,
			Message:     "Firmware",
			Provisions: []model.Provision{{
				FilePattern: "*.nca",
				Description: "Switch firmware",
			}},
		}},
	}

	results := checker.Check(emu, model.SystemIDSwitch)
	if results[0].Results[0].Status != model.ProvisionMissing {
		t.Error("Empty bios directory should be missing")
	}

	if err := os.WriteFile(filepath.Join(biosDir, "0100000000000809.nca"), []byte("firmware"), 0644); err != nil {
		t.Fatalf("Failed to create NCA file: %v", err)
	}

	results = checker.Check(emu, model.SystemIDSwitch)
	if results[0].Results[0].Status != model.ProvisionFound {
		t.Error("Bios directory with .nca files should be found")
	}
	if results[0].Results[0].FoundPath != biosDir {
		t.Errorf("Expected FoundPath %s, got %s", biosDir, results[0].Results[0].FoundPath)
	}
}

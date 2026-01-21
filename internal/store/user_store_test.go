package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestUserStoreInitialize(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewUserStore(tmpDir)

	// Initialize should create all directories
	if err := store.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify directories exist
	for _, dir := range []string{"roms", "bios", "saves", "states", "screenshots"} {
		path := filepath.Join(tmpDir, dir)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Directory %s not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}

	// IsInitialized should return true
	if !store.IsInitialized() {
		t.Error("IsInitialized returned false after Initialize")
	}
}

func TestUserStoreInitializeSystem(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewUserStore(tmpDir)

	// Initialize base structure first
	if err := store.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Initialize SNES system
	if err := store.InitializeSystem(model.SystemSNES); err != nil {
		t.Fatalf("InitializeSystem failed: %v", err)
	}

	// Verify system directories exist
	expectedDirs := []string{
		filepath.Join(tmpDir, "roms", "snes"),
		filepath.Join(tmpDir, "bios", "snes"),
		filepath.Join(tmpDir, "saves", "snes"),
		filepath.Join(tmpDir, "states", "snes"),
		filepath.Join(tmpDir, "screenshots", "snes"),
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("System directory not created: %s: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestUserStorePaths(t *testing.T) {
	store := NewUserStore("/home/user/Emulation")

	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{"RomsDir", store.RomsDir, "/home/user/Emulation/roms"},
		{"BiosDir", store.BiosDir, "/home/user/Emulation/bios"},
		{"SavesDir", store.SavesDir, "/home/user/Emulation/saves"},
		{"StatesDir", store.StatesDir, "/home/user/Emulation/states"},
		{"ScreenshotsDir", store.ScreenshotsDir, "/home/user/Emulation/screenshots"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUserStoreSystemPaths(t *testing.T) {
	store := NewUserStore("/home/user/Emulation")

	tests := []struct {
		name   string
		system model.SystemID
		fn     func(model.SystemID) string
		want   string
	}{
		{"SystemRomsDir", model.SystemSNES, store.SystemRomsDir, "/home/user/Emulation/roms/snes"},
		{"SystemBiosDir", model.SystemPSX, store.SystemBiosDir, "/home/user/Emulation/bios/psx"},
		{"SystemSavesDir", model.SystemTIC80, store.SystemSavesDir, "/home/user/Emulation/saves/tic80"},
		{"SystemStatesDir", model.SystemSNES, store.SystemStatesDir, "/home/user/Emulation/states/snes"},
		{"SystemScreenshotsDir", model.SystemPSX, store.SystemScreenshotsDir, "/home/user/Emulation/screenshots/psx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(tt.system)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestUserStoreExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Non-existent directory
	store := NewUserStore(filepath.Join(tmpDir, "nonexistent"))
	if store.Exists() {
		t.Error("Exists returned true for non-existent directory")
	}

	// Existing directory
	store = NewUserStore(tmpDir)
	if !store.Exists() {
		t.Error("Exists returned false for existing directory")
	}

	// File (not directory)
	filePath := filepath.Join(tmpDir, "file")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	store = NewUserStore(filePath)
	if store.Exists() {
		t.Error("Exists returned true for file (not directory)")
	}
}

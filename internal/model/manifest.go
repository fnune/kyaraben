package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fnune/kyaraben/internal/paths"
)

// Manifest tracks what kyaraben has installed and configured.
type Manifest struct {
	Version            int                              `json:"version"`
	LastApplied        time.Time                        `json:"last_applied"`
	InstalledEmulators map[EmulatorID]InstalledEmulator `json:"installed_emulators"`
	ManagedConfigs     []ManagedConfig                  `json:"managed_configs"`
}

// InstalledEmulator tracks an installed emulator.
type InstalledEmulator struct {
	ID        EmulatorID `json:"id"`
	Version   string     `json:"version"`
	StorePath string     `json:"store_path"` // Path in Nix store
	Installed time.Time  `json:"installed"`
}

// ManagedConfig tracks a config file managed by kyaraben.
type ManagedConfig struct {
	Path         string     `json:"path"`          // Path to the config file
	BaselineHash string     `json:"baseline_hash"` // Hash of baseline (what we last wrote)
	LastModified time.Time  `json:"last_modified"`
	EmulatorID   EmulatorID `json:"emulator_id"`
}

// NewManifest creates a new empty manifest.
func NewManifest() *Manifest {
	return &Manifest{
		Version:            1,
		InstalledEmulators: make(map[EmulatorID]InstalledEmulator),
		ManagedConfigs:     make([]ManagedConfig, 0),
	}
}

// DefaultManifestPath returns the default path to the manifest file.
func DefaultManifestPath() (string, error) {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "manifest.json"), nil
}

// LoadManifest loads the manifest from a file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewManifest(), nil
		}
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}

// Save writes the manifest to a file atomically.
// It writes to a temporary file first, then renames to the target path.
// This prevents corruption if the process crashes during write.
func (m *Manifest) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating manifest directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding manifest: %w", err)
	}

	tempFile, err := os.CreateTemp(dir, "manifest-*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tempPath := tempFile.Name()

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("writing manifest: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("renaming manifest: %w", err)
	}

	return nil
}

// AddEmulator records an installed emulator.
func (m *Manifest) AddEmulator(emu InstalledEmulator) {
	m.InstalledEmulators[emu.ID] = emu
}

// AddManagedConfig records a managed config file.
func (m *Manifest) AddManagedConfig(cfg ManagedConfig) {
	// Update existing or append
	for i, existing := range m.ManagedConfigs {
		if existing.Path == cfg.Path {
			m.ManagedConfigs[i] = cfg
			return
		}
	}
	m.ManagedConfigs = append(m.ManagedConfigs, cfg)
}

// GetEmulator returns an installed emulator by ID.
func (m *Manifest) GetEmulator(id EmulatorID) (InstalledEmulator, bool) {
	emu, ok := m.InstalledEmulators[id]
	return emu, ok
}

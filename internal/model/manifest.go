package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fnune/kyaraben/internal/fileutil"
	"github.com/fnune/kyaraben/internal/paths"
)

// Manifest tracks what kyaraben has installed and configured.
type Manifest struct {
	Version            int                              `json:"version"`
	KyarabenVersion    string                           `json:"kyaraben_version,omitempty"`
	LastApplied        time.Time                        `json:"last_applied"`
	InstalledEmulators map[EmulatorID]InstalledEmulator `json:"installed_emulators"`
	InstalledFrontends map[FrontendID]InstalledFrontend `json:"installed_frontends,omitempty"`
	ManagedConfigs     []ManagedConfig                  `json:"managed_configs"`
	DesktopFiles       []string                         `json:"desktop_files,omitempty"`
	IconFiles          []string                         `json:"icon_files,omitempty"`
	KyarabenInstall    *KyarabenInstall                 `json:"kyaraben_install,omitempty"`
}

// KyarabenInstall tracks the kyaraben app installation paths.
type KyarabenInstall struct {
	AppPath     string `json:"app_path,omitempty"`
	CLIPath     string `json:"cli_path,omitempty"`
	DesktopPath string `json:"desktop_path,omitempty"`
}

// InstalledEmulator tracks an installed emulator.
type InstalledEmulator struct {
	ID        EmulatorID `json:"id"`
	Version   string     `json:"version"`
	StorePath string     `json:"store_path"` // Path in Nix store
	Installed time.Time  `json:"installed"`
}

// InstalledFrontend tracks an installed frontend.
type InstalledFrontend struct {
	ID        FrontendID `json:"id"`
	Version   string     `json:"version"`
	StorePath string     `json:"store_path"`
	Installed time.Time  `json:"installed"`
}

type ManagedKey struct {
	Path  []string `json:"path"`
	Value string   `json:"value"`
}

type ManagedConfig struct {
	EmulatorIDs  []EmulatorID `json:"emulator_ids"`
	Target       ConfigTarget `json:"target"`
	BaselineHash string       `json:"baseline_hash"`
	LastModified time.Time    `json:"last_modified"`
	ManagedKeys  []ManagedKey `json:"managed_keys"`
}

// NewManifest creates a new empty manifest.
func NewManifest() *Manifest {
	return &Manifest{
		Version:            1,
		InstalledEmulators: make(map[EmulatorID]InstalledEmulator),
		InstalledFrontends: make(map[FrontendID]InstalledFrontend),
		ManagedConfigs:     make([]ManagedConfig, 0),
	}
}

func DefaultManifestPath() (string, error) {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "build", "manifest.json"), nil
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

func (m *Manifest) SaveWithBackup(path string) error {
	if _, err := os.Stat(path); err == nil {
		_, _ = fileutil.BackupWithTimestamp(path)
	}
	return m.Save(path)
}

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

	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("syncing manifest: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("renaming manifest: %w", err)
	}

	if err := syncDir(dir); err != nil {
		return fmt.Errorf("syncing directory: %w", err)
	}

	return nil
}

func syncDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = d.Close() }()
	return d.Sync()
}

// AddEmulator records an installed emulator.
func (m *Manifest) AddEmulator(emu InstalledEmulator) {
	m.InstalledEmulators[emu.ID] = emu
}

// AddFrontend records an installed frontend.
func (m *Manifest) AddFrontend(fe InstalledFrontend) {
	if m.InstalledFrontends == nil {
		m.InstalledFrontends = make(map[FrontendID]InstalledFrontend)
	}
	m.InstalledFrontends[fe.ID] = fe
}

func (m *Manifest) GetFrontend(id FrontendID) (InstalledFrontend, bool) {
	if m.InstalledFrontends == nil {
		return InstalledFrontend{}, false
	}
	fe, ok := m.InstalledFrontends[id]
	return fe, ok
}

func (m *Manifest) AddManagedConfig(cfg ManagedConfig) error {
	for i, existing := range m.ManagedConfigs {
		if existing.Target == cfg.Target {
			m.ManagedConfigs[i].EmulatorIDs = appendUniqueEmulatorIDs(existing.EmulatorIDs, cfg.EmulatorIDs...)
			m.ManagedConfigs[i].ManagedKeys = cfg.ManagedKeys
			m.ManagedConfigs[i].BaselineHash = cfg.BaselineHash
			m.ManagedConfigs[i].LastModified = cfg.LastModified
			return nil
		}
	}
	m.ManagedConfigs = append(m.ManagedConfigs, cfg)
	return nil
}

func appendUniqueEmulatorIDs(slice []EmulatorID, elems ...EmulatorID) []EmulatorID {
	for _, e := range elems {
		found := false
		for _, s := range slice {
			if s == e {
				found = true
				break
			}
		}
		if !found {
			slice = append(slice, e)
		}
	}
	return slice
}

func (m *Manifest) GetEmulator(id EmulatorID) (InstalledEmulator, bool) {
	emu, ok := m.InstalledEmulators[id]
	return emu, ok
}

func (m *Manifest) GetManagedConfig(target ConfigTarget) (ManagedConfig, bool) {
	for _, cfg := range m.ManagedConfigs {
		if cfg.Target == target {
			return cfg, true
		}
	}
	return ManagedConfig{}, false
}

func (m *Manifest) GetManagedConfigsForEmulator(emuID EmulatorID) []ManagedConfig {
	var configs []ManagedConfig
	for _, cfg := range m.ManagedConfigs {
		for _, id := range cfg.EmulatorIDs {
			if id == emuID {
				configs = append(configs, cfg)
				break
			}
		}
	}
	return configs
}

package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// KyarabenConfig represents the user's kyaraben configuration.
type KyarabenConfig struct {
	Global  GlobalConfig            `toml:"global"`
	Sync    SyncConfig              `toml:"sync"`
	Systems map[SystemID]SystemConf `toml:"systems"`
}

// GlobalConfig holds global settings.
type GlobalConfig struct {
	UserStore string `toml:"user_store"` // Path to emulation directory
}

// SystemConf holds per-system configuration.
type SystemConf struct {
	Emulator string `toml:"emulator"` // "eden" or "eden@v0.1.0"
}

// EmulatorID returns the emulator ID from the Emulator field.
func (s SystemConf) EmulatorID() EmulatorID {
	if idx := strings.Index(s.Emulator, "@"); idx != -1 {
		return EmulatorID(s.Emulator[:idx])
	}
	return EmulatorID(s.Emulator)
}

// EmulatorVersion returns the pinned version, or empty string for default.
func (s SystemConf) EmulatorVersion() string {
	if idx := strings.Index(s.Emulator, "@"); idx != -1 {
		return s.Emulator[idx+1:]
	}
	return ""
}

func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config directory: %w", err)
	}
	return filepath.Join(configDir, "kyaraben", "config.toml"), nil
}

func DefaultUserStore() string {
	return "~/Emulation"
}

// LoadConfig loads the kyaraben configuration from a file.
func LoadConfig(path string) (*KyarabenConfig, error) {
	var cfg KyarabenConfig
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}
	return &cfg, nil
}

// SaveConfig writes the kyaraben configuration to a file.
func SaveConfig(cfg *KyarabenConfig, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}

	encoder := toml.NewEncoder(f)
	encodeErr := encoder.Encode(cfg)
	closeErr := f.Close()

	if encodeErr != nil {
		return fmt.Errorf("encoding config: %w", encodeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing config file: %w", closeErr)
	}
	return nil
}

// ExpandUserStore expands ~ in the user store path.
func (c *KyarabenConfig) ExpandUserStore() (string, error) {
	path := c.Global.UserStore
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expanding home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}
	return path, nil
}

func (c *KyarabenConfig) EnabledSystems() []SystemID {
	systems := make([]SystemID, 0, len(c.Systems))
	for id := range c.Systems {
		systems = append(systems, id)
	}
	return systems
}

// NewDefaultConfig creates a new config with default values.
func NewDefaultConfig() *KyarabenConfig {
	return &KyarabenConfig{
		Global: GlobalConfig{
			UserStore: DefaultUserStore(),
		},
		Sync:    DefaultSyncConfig(),
		Systems: make(map[SystemID]SystemConf),
	}
}

// BuildVersionOverrides returns a map from package names to pinned versions
// based on the emulator versions configured in the systems.
func (c *KyarabenConfig) BuildVersionOverrides(getEmulator func(EmulatorID) (Emulator, error)) (map[string]string, error) {
	overrides := make(map[string]string)
	for _, sysConf := range c.Systems {
		version := sysConf.EmulatorVersion()
		if version == "" {
			continue
		}
		emu, err := getEmulator(sysConf.EmulatorID())
		if err != nil {
			return nil, fmt.Errorf("unknown emulator %q: %w", sysConf.EmulatorID(), err)
		}
		overrides[emu.Package.PackageName()] = version
	}
	return overrides, nil
}

package model

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/fnune/kyaraben/internal/paths"
)

// KyarabenConfig represents the user's kyaraben configuration.
type KyarabenConfig struct {
	Global    GlobalConfig                  `toml:"global"`
	Sync      SyncConfig                    `toml:"sync"`
	Systems   map[SystemID][]EmulatorID     `toml:"systems"`
	Emulators map[EmulatorID]EmulatorConf   `toml:"emulators,omitempty"`
	Frontends map[FrontendID]FrontendConfig `toml:"frontends,omitempty"`
}

// FrontendConfig holds per-frontend configuration.
type FrontendConfig struct {
	Enabled bool   `toml:"enabled"`
	Version string `toml:"version,omitempty"`
}

// GlobalConfig holds global settings.
type GlobalConfig struct {
	UserStore string `toml:"user_store"` // Path to emulation directory
}

// EmulatorConf holds per-emulator configuration.
type EmulatorConf struct {
	Version string `toml:"version,omitempty"`
}

// EmulatorVersion returns the configured version for an emulator, or empty for default.
func (c *KyarabenConfig) EmulatorVersion(id EmulatorID) string {
	if c.Emulators == nil {
		return ""
	}
	if conf, ok := c.Emulators[id]; ok {
		return conf.Version
	}
	return ""
}

func DefaultConfigPath() (string, error) {
	configDir, err := paths.KyarabenConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config directory: %w", err)
	}
	return filepath.Join(configDir, "config.toml"), nil
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
	for id, emulators := range c.Systems {
		if len(emulators) > 0 {
			systems = append(systems, id)
		}
	}
	return systems
}

// EnabledEmulators returns a deduplicated list of all enabled emulator IDs.
func (c *KyarabenConfig) EnabledEmulators() []EmulatorID {
	seen := make(map[EmulatorID]bool)
	var result []EmulatorID
	for _, emulators := range c.Systems {
		for _, eid := range emulators {
			if !seen[eid] {
				seen[eid] = true
				result = append(result, eid)
			}
		}
	}
	return result
}

// EmulatorsForSystem returns the emulators enabled for a specific system.
func (c *KyarabenConfig) EmulatorsForSystem(sys SystemID) []EmulatorID {
	return c.Systems[sys]
}

// SystemsForEmulator returns all systems that have the given emulator enabled.
func (c *KyarabenConfig) SystemsForEmulator(emu EmulatorID) []SystemID {
	var result []SystemID
	for sys, emulators := range c.Systems {
		for _, eid := range emulators {
			if eid == emu {
				result = append(result, sys)
				break
			}
		}
	}
	return result
}

// NewDefaultConfig creates a new config with default values.
func NewDefaultConfig() *KyarabenConfig {
	return &KyarabenConfig{
		Global: GlobalConfig{
			UserStore: DefaultUserStore(),
		},
		Sync: DefaultSyncConfig(),
		Systems: map[SystemID][]EmulatorID{
			SystemIDNES:      {EmulatorIDRetroArchMesen},
			SystemIDSNES:     {EmulatorIDRetroArchBsnes},
			SystemIDN64:      {EmulatorIDRetroArchMupen64Plus},
			SystemIDGB:       {EmulatorIDMGBA},
			SystemIDGBC:      {EmulatorIDMGBA},
			SystemIDGBA:      {EmulatorIDMGBA},
			SystemIDNDS:      {EmulatorIDMelonDS},
			SystemIDPSX:      {EmulatorIDDuckStation},
			SystemIDPS2:      {EmulatorIDPCSX2},
			SystemIDGenesis:  {EmulatorIDRetroArchGenesisPlusGX},
			SystemIDGameCube: {EmulatorIDDolphin},
			SystemIDWii:      {EmulatorIDDolphin},
			SystemIDPSP:      {EmulatorIDPPSSPP},
		},
		Emulators: make(map[EmulatorID]EmulatorConf),
		Frontends: map[FrontendID]FrontendConfig{
			FrontendIDESDE: {Enabled: true},
		},
	}
}

// BuildVersionOverrides returns a map from package names to pinned versions
// based on the emulator versions configured in the emulators section.
func (c *KyarabenConfig) BuildVersionOverrides(getEmulator func(EmulatorID) (Emulator, error)) (map[string]string, error) {
	overrides := make(map[string]string)
	for emuID, emuConf := range c.Emulators {
		if emuConf.Version == "" {
			continue
		}
		emu, err := getEmulator(emuID)
		if err != nil {
			return nil, fmt.Errorf("unknown emulator %q: %w", emuID, err)
		}
		overrides[emu.Package.PackageName()] = emuConf.Version
	}
	return overrides, nil
}

// EnabledFrontends returns a list of frontend IDs that are enabled.
func (c *KyarabenConfig) EnabledFrontends() []FrontendID {
	var result []FrontendID
	for id, conf := range c.Frontends {
		if conf.Enabled {
			result = append(result, id)
		}
	}
	return result
}

// FrontendVersion returns the configured version for a frontend, or empty for default.
func (c *KyarabenConfig) FrontendVersion(id FrontendID) string {
	if c.Frontends == nil {
		return ""
	}
	if conf, ok := c.Frontends[id]; ok {
		return conf.Version
	}
	return ""
}

// BuildFrontendVersionOverrides returns a map from package names to pinned versions
// based on frontend versions.
func (c *KyarabenConfig) BuildFrontendVersionOverrides(getFrontend func(FrontendID) (Frontend, error)) (map[string]string, error) {
	overrides := make(map[string]string)
	for _, frontendID := range c.EnabledFrontends() {
		version := c.FrontendVersion(frontendID)
		if version == "" {
			continue
		}
		frontend, err := getFrontend(frontendID)
		if err != nil {
			return nil, fmt.Errorf("unknown frontend %q: %w", frontendID, err)
		}
		overrides[frontend.Package.PackageName()] = version
	}
	return overrides, nil
}

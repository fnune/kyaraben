package model

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/paths"
)

// KyarabenConfig represents the user's kyaraben configuration.
type KyarabenConfig struct {
	Global     GlobalConfig                  `toml:"global"`
	Sync       SyncConfig                    `toml:"sync"`
	Controller ControllerTomlConfig          `toml:"controller"`
	Systems    map[SystemID][]EmulatorID     `toml:"systems"`
	Emulators  map[EmulatorID]EmulatorConf   `toml:"emulators,omitempty"`
	Frontends  map[FrontendID]FrontendConfig `toml:"frontends,omitempty"`
}

// ControllerTomlConfig is the TOML representation of controller settings.
type ControllerTomlConfig struct {
	Layout  string           `toml:"layout"`
	Hotkeys HotkeyTomlConfig `toml:"hotkeys"`
}

// HotkeyTomlConfig is the TOML representation of hotkey bindings.
// All hotkeys share the same modifier button, with individual action buttons.
type HotkeyTomlConfig struct {
	Modifier         string `toml:"modifier"`
	SaveState        string `toml:"save_state"`
	LoadState        string `toml:"load_state"`
	NextSlot         string `toml:"next_slot"`
	PrevSlot         string `toml:"prev_slot"`
	FastForward      string `toml:"fast_forward"`
	Rewind           string `toml:"rewind"`
	Pause            string `toml:"pause"`
	Screenshot       string `toml:"screenshot"`
	Quit             string `toml:"quit"`
	ToggleFullscreen string `toml:"toggle_fullscreen"`
	OpenMenu         string `toml:"open_menu"`
}

// ResolveControllerConfig validates and resolves the TOML controller config into
// the typed ControllerConfig used by generators.
func (c *KyarabenConfig) ResolveControllerConfig() (*ControllerConfig, error) {
	cc := &ControllerConfig{
		Layout:  LayoutStandard,
		Hotkeys: DefaultHotkeys(),
	}

	if c.Controller.Layout != "" {
		layout, err := ValidateLayoutID(c.Controller.Layout)
		if err != nil {
			return nil, err
		}
		cc.Layout = layout
	}

	hk := c.Controller.Hotkeys

	modifier := SDLButton(ButtonBack)
	if hk.Modifier != "" {
		if !validButtons[SDLButton(hk.Modifier)] {
			return nil, fmt.Errorf("controller.hotkeys.modifier: unknown button %q", hk.Modifier)
		}
		modifier = SDLButton(hk.Modifier)
	}

	resolvers := []struct {
		src  string
		dest *HotkeyBinding
	}{
		{hk.SaveState, &cc.Hotkeys.SaveState},
		{hk.LoadState, &cc.Hotkeys.LoadState},
		{hk.NextSlot, &cc.Hotkeys.NextSlot},
		{hk.PrevSlot, &cc.Hotkeys.PrevSlot},
		{hk.FastForward, &cc.Hotkeys.FastForward},
		{hk.Rewind, &cc.Hotkeys.Rewind},
		{hk.Pause, &cc.Hotkeys.Pause},
		{hk.Screenshot, &cc.Hotkeys.Screenshot},
		{hk.Quit, &cc.Hotkeys.Quit},
		{hk.ToggleFullscreen, &cc.Hotkeys.ToggleFullscreen},
		{hk.OpenMenu, &cc.Hotkeys.OpenMenu},
	}
	for _, r := range resolvers {
		if r.src == "" {
			continue
		}
		action := SDLButton(r.src)
		if !validButtons[action] {
			return nil, fmt.Errorf("controller.hotkeys: unknown button %q", r.src)
		}
		*r.dest = HotkeyBinding{Buttons: []SDLButton{modifier, action}}
	}

	return cc, nil
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

type ConfigStore struct {
	fs vfs.FS
}

func NewConfigStore(fs vfs.FS) *ConfigStore {
	return &ConfigStore{fs: fs}
}

type ConfigValidators struct {
	GetEmulator func(EmulatorID) (Emulator, error)
	GetSystem   func(SystemID) (System, error)
	GetFrontend func(FrontendID) (Frontend, error)
}

func (s *ConfigStore) Load(path string, validators *ConfigValidators) (*KyarabenConfig, error) {
	data, err := s.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg KyarabenConfig
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}
	if validators != nil {
		if err := cfg.validate(validators); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func (c *KyarabenConfig) validate(v *ConfigValidators) error {
	for sysID, emulatorIDs := range c.Systems {
		if _, err := v.GetSystem(sysID); err != nil {
			return fmt.Errorf("unknown system %q", sysID)
		}
		for _, emuID := range emulatorIDs {
			if _, err := v.GetEmulator(emuID); err != nil {
				return fmt.Errorf("system %s: unknown emulator %q", sysID, emuID)
			}
		}
	}
	for feID := range c.Frontends {
		if _, err := v.GetFrontend(feID); err != nil {
			return fmt.Errorf("unknown frontend %q", feID)
		}
	}
	return nil
}

func (s *ConfigStore) Save(cfg *KyarabenConfig, path string) error {
	dir := filepath.Dir(path)
	if err := vfs.MkdirAll(s.fs, dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	f, err := s.fs.Create(path)
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

func LoadConfig(path string, validators *ConfigValidators) (*KyarabenConfig, error) {
	return NewConfigStore(vfs.OSFS).Load(path, validators)
}

func SaveConfig(cfg *KyarabenConfig, path string) error {
	return NewConfigStore(vfs.OSFS).Save(cfg, path)
}

func (c *KyarabenConfig) ExpandUserStoreWith(homeDir string) string {
	path := c.Global.UserStore
	if len(path) > 0 && path[0] == '~' && homeDir != "" {
		path = filepath.Join(homeDir, path[1:])
	}
	return path
}

func (c *KyarabenConfig) ExpandUserStore() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expanding home directory: %w", err)
	}
	return c.ExpandUserStoreWith(home), nil
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
		Controller: ControllerTomlConfig{
			Layout: string(LayoutStandard),
			Hotkeys: HotkeyTomlConfig{
				Modifier:         string(ButtonBack),
				SaveState:        string(ButtonRightShoulder),
				LoadState:        string(ButtonLeftShoulder),
				NextSlot:         string(ButtonDPadRight),
				PrevSlot:         string(ButtonDPadLeft),
				FastForward:      string(ButtonY),
				Rewind:           string(ButtonX),
				Pause:            string(ButtonA),
				Screenshot:       string(ButtonB),
				Quit:             string(ButtonStart),
				ToggleFullscreen: string(ButtonLeftStick),
				OpenMenu:         string(ButtonRightStick),
			},
		},
		Systems: map[SystemID][]EmulatorID{
			SystemIDNES:      {EmulatorIDRetroArchMesen},
			SystemIDSNES:     {EmulatorIDRetroArchBsnes},
			SystemIDN64:      {EmulatorIDRetroArchMupen64Plus},
			SystemIDGB:       {EmulatorIDRetroArchMGBA},
			SystemIDGBC:      {EmulatorIDRetroArchMGBA},
			SystemIDGBA:      {EmulatorIDRetroArchMGBA},
			SystemIDNDS:      {EmulatorIDRetroArchMelonDS},
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
// based on the emulator and frontend versions configured in the config.
func (c *KyarabenConfig) BuildVersionOverrides(
	getEmulator func(EmulatorID) (Emulator, error),
	getFrontend func(FrontendID) (Frontend, error),
) (map[string]string, error) {
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
	for feID, feConf := range c.Frontends {
		if feConf.Version == "" {
			continue
		}
		fe, err := getFrontend(feID)
		if err != nil {
			return nil, fmt.Errorf("unknown frontend %q: %w", feID, err)
		}
		overrides[fe.Package.PackageName()] = feConf.Version
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

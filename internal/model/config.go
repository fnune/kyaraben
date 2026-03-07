package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/paths"
)

// ConfigWarning represents a non-fatal issue found during config loading.
type ConfigWarning struct {
	Field   string
	Message string
}

func (w ConfigWarning) String() string {
	if w.Field != "" {
		return fmt.Sprintf("%s: %s", w.Field, w.Message)
	}
	return w.Message
}

// ConfigWarnings collects multiple warnings during config loading.
type ConfigWarnings []ConfigWarning

func (w ConfigWarnings) Error() string {
	if len(w) == 0 {
		return ""
	}
	var msgs []string
	for _, warn := range w {
		msgs = append(msgs, warn.String())
	}
	return fmt.Sprintf("config has %d warning(s):\n  - %s", len(w), strings.Join(msgs, "\n  - "))
}

func (w ConfigWarnings) HasWarnings() bool {
	return len(w) > 0
}

// KyarabenConfig represents the user's kyaraben configuration.
type KyarabenConfig struct {
	Global     GlobalConfig                  `toml:"global"`
	Graphics   GraphicsConfig                `toml:"graphics"`
	Savestate  SavestateConfig               `toml:"savestate"`
	Sync       SyncConfig                    `toml:"sync"`
	Controller ControllerTomlConfig          `toml:"controller"`
	Systems    map[SystemID][]EmulatorID     `toml:"systems"`
	Emulators  map[EmulatorID]EmulatorConf   `toml:"emulators,omitempty"`
	Frontends  map[FrontendID]FrontendConfig `toml:"frontends,omitempty"`
}

// GraphicsConfig holds graphics-related default settings.
type GraphicsConfig struct {
	Preset string `toml:"preset,omitempty"`
	Bezels *bool  `toml:"bezels,omitempty"`
	Target string `toml:"target,omitempty"`
}

// DisplayPreset values for graphics.preset.
const (
	PresetClean  = "clean"
	PresetRetro  = "retro"
	PresetManual = "manual"
)

// TargetDevice values for graphics.target.
const (
	TargetAuto      = "auto"
	TargetDesktop   = "desktop"
	TargetSteamDeck = "steam-deck"
)

// Per-emulator preset override values.
const (
	EmulatorPresetClean  = "clean"
	EmulatorPresetRetro  = "retro"
	EmulatorPresetManual = "manual"
)

// ConfigInput for GraphicsConfig fields
const (
	ConfigInputPreset ConfigInput = "graphics.preset"
	ConfigInputBezels ConfigInput = "graphics.bezels"
	ConfigInputTarget ConfigInput = "graphics.target"
)

// SavestateConfig holds savestate/resume-related settings.
type SavestateConfig struct {
	Resume string `toml:"resume,omitempty"`
}

// Resume setting values for savestate.resume (global setting).
// "recommended" enables resume only on emulators that recommend it.
const (
	ResumeRecommended = "recommended"
	ResumeOff         = "off"
	ResumeManual      = "manual"
)

// Per-emulator resume override values.
const (
	EmulatorResumeOn     = "on"
	EmulatorResumeOff    = "off"
	EmulatorResumeManual = "manual"
)

// ConfigInput for SavestateConfig.Resume
const ConfigInputResume ConfigInput = "savestate.resume"

// ControllerTomlConfig is the TOML representation of controller settings.
type ControllerTomlConfig struct {
	NintendoConfirm string           `toml:"nintendo_confirm"`
	Hotkeys         HotkeyTomlConfig `toml:"hotkeys"`
}

// ConfigInput for ControllerTomlConfig.NintendoConfirm
const ConfigInputNintendoConfirm ConfigInput = "controller.nintendo_confirm"

// ConfigInput for ControllerTomlConfig.Hotkeys
const ConfigInputHotkeys ConfigInput = "controller.hotkeys"

// ConfigInput for GlobalConfig.Collection
const ConfigInputCollection ConfigInput = "global.collection"

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

// ControllerConfigResult contains the resolved controller config and any warnings.
type ControllerConfigResult struct {
	Config   *ControllerConfig
	Warnings ConfigWarnings
}

// ResolveControllerConfig validates and resolves the TOML controller config into
// the typed ControllerConfig used by generators. Invalid values are replaced with
// defaults and warnings are returned.
func (c *KyarabenConfig) ResolveControllerConfig() (*ControllerConfig, error) {
	result := c.ResolveControllerConfigWithWarnings()
	return result.Config, nil
}

// ResolveControllerConfigWithWarnings is like ResolveControllerConfig but also
// returns warnings for invalid values that were replaced with defaults.
func (c *KyarabenConfig) ResolveControllerConfigWithWarnings() *ControllerConfigResult {
	var warnings ConfigWarnings

	cc := &ControllerConfig{
		NintendoConfirm: NintendoConfirmEast,
		Hotkeys:         DefaultHotkeys(),
	}

	if c.Controller.NintendoConfirm != "" {
		confirm, err := ValidateNintendoConfirmButton(c.Controller.NintendoConfirm)
		if err != nil {
			warnings = append(warnings, ConfigWarning{
				Field:   "controller.nintendo_confirm",
				Message: fmt.Sprintf("invalid value %q (valid: %q, %q), using default %q", c.Controller.NintendoConfirm, NintendoConfirmSouth, NintendoConfirmEast, NintendoConfirmEast),
			})
		} else {
			cc.NintendoConfirm = confirm
		}
	}

	hk := c.Controller.Hotkeys

	modifier := SDLButton(ButtonBack)
	modifierValid := true
	if hk.Modifier != "" {
		if !validButtons[SDLButton(hk.Modifier)] {
			warnings = append(warnings, ConfigWarning{
				Field:   "controller.hotkeys.modifier",
				Message: fmt.Sprintf("invalid button %q, using default %q", hk.Modifier, ButtonBack),
			})
			modifierValid = false
		} else {
			modifier = SDLButton(hk.Modifier)
		}
	}

	hotkeyFields := []struct {
		src   string
		field string
		dest  *HotkeyBinding
	}{
		{hk.SaveState, "save_state", &cc.Hotkeys.SaveState},
		{hk.LoadState, "load_state", &cc.Hotkeys.LoadState},
		{hk.NextSlot, "next_slot", &cc.Hotkeys.NextSlot},
		{hk.PrevSlot, "prev_slot", &cc.Hotkeys.PrevSlot},
		{hk.FastForward, "fast_forward", &cc.Hotkeys.FastForward},
		{hk.Rewind, "rewind", &cc.Hotkeys.Rewind},
		{hk.Pause, "pause", &cc.Hotkeys.Pause},
		{hk.Screenshot, "screenshot", &cc.Hotkeys.Screenshot},
		{hk.Quit, "quit", &cc.Hotkeys.Quit},
		{hk.ToggleFullscreen, "toggle_fullscreen", &cc.Hotkeys.ToggleFullscreen},
		{hk.OpenMenu, "open_menu", &cc.Hotkeys.OpenMenu},
	}
	for _, r := range hotkeyFields {
		if r.src == "" {
			continue
		}
		action := SDLButton(r.src)
		if !validButtons[action] {
			warnings = append(warnings, ConfigWarning{
				Field:   fmt.Sprintf("controller.hotkeys.%s", r.field),
				Message: fmt.Sprintf("invalid button %q, using default", r.src),
			})
			continue
		}
		if modifierValid {
			*r.dest = HotkeyBinding{Buttons: []SDLButton{modifier, action}}
		}
	}

	return &ControllerConfigResult{Config: cc, Warnings: warnings}
}

// FrontendConfig holds per-frontend configuration.
type FrontendConfig struct {
	Enabled bool   `toml:"enabled"`
	Version string `toml:"version,omitempty"`
}

// GlobalConfig holds global settings.
type GlobalConfig struct {
	Collection string `toml:"collection"`
}

// EmulatorConf holds per-emulator configuration.
type EmulatorConf struct {
	Version string  `toml:"version,omitempty"`
	Preset  *string `toml:"preset,omitempty"`
	Resume  *string `toml:"resume,omitempty"`
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

// EmulatorPreset returns the resolved preset for an emulator.
// Resolution order:
//  1. Per-emulator override
//  2. Global preset setting
//  3. Default: pseudo-authentic
func (c *KyarabenConfig) EmulatorPreset(id EmulatorID) string {
	if c.Emulators != nil {
		if conf, ok := c.Emulators[id]; ok && conf.Preset != nil {
			return *conf.Preset
		}
	}
	if c.Graphics.Preset != "" {
		return c.Graphics.Preset
	}
	return PresetClean
}

// EmulatorPresetOverride returns the per-emulator preset override, or nil if using default.
func (c *KyarabenConfig) EmulatorPresetOverride(id EmulatorID) *string {
	if c.Emulators == nil {
		return nil
	}
	if conf, ok := c.Emulators[id]; ok {
		return conf.Preset
	}
	return nil
}

// GraphicsBezels returns whether bezels are enabled.
func (c *KyarabenConfig) GraphicsBezels() bool {
	if c.Graphics.Bezels != nil {
		return *c.Graphics.Bezels
	}
	return true
}

// GraphicsTarget returns the target device setting.
func (c *KyarabenConfig) GraphicsTarget() string {
	if c.Graphics.Target != "" {
		return c.Graphics.Target
	}
	return TargetAuto
}

// EmulatorResume returns the resolved resume setting for an emulator.
// Resolution order:
//  1. Per-emulator override (on/off/manual)
//  2. If global = "recommended" and emulator recommends resume -> "on"
//  3. If global = "off" -> "off"
//  4. Otherwise -> "manual"
func (c *KyarabenConfig) EmulatorResume(id EmulatorID, resumeRecommended bool) string {
	if c.Emulators != nil {
		if conf, ok := c.Emulators[id]; ok && conf.Resume != nil {
			return *conf.Resume
		}
	}
	switch c.Savestate.Resume {
	case ResumeRecommended:
		if resumeRecommended {
			return EmulatorResumeOn
		}
		return EmulatorResumeManual
	case ResumeOff:
		return EmulatorResumeOff
	default:
		return EmulatorResumeManual
	}
}

// EmulatorResumeOverride returns the per-emulator resume override, or nil if using default.
func (c *KyarabenConfig) EmulatorResumeOverride(id EmulatorID) *string {
	if c.Emulators == nil {
		return nil
	}
	if conf, ok := c.Emulators[id]; ok {
		return conf.Resume
	}
	return nil
}

func DefaultConfigPath() (string, error) {
	configDir, err := paths.KyarabenConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting config directory: %w", err)
	}
	return filepath.Join(configDir, "config.toml"), nil
}

func DefaultCollection() string {
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

// LoadResult contains the loaded config and any warnings encountered.
type LoadResult struct {
	Config   *KyarabenConfig
	Warnings ConfigWarnings
}

func (s *ConfigStore) Load(path string, validators *ConfigValidators) (*KyarabenConfig, error) {
	result, err := s.LoadWithWarnings(path, validators)
	if err != nil {
		return nil, err
	}
	return result.Config, nil
}

func (s *ConfigStore) LoadWithWarnings(path string, validators *ConfigValidators) (*LoadResult, error) {
	data, err := s.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg KyarabenConfig
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}
	cfg.applyDefaults()

	var warnings ConfigWarnings
	if validators != nil {
		warnings = cfg.validateAndFilter(validators)
	}

	return &LoadResult{Config: &cfg, Warnings: warnings}, nil
}

func (c *KyarabenConfig) applyDefaults() {
	if c.Controller.NintendoConfirm == "" {
		c.Controller.NintendoConfirm = string(NintendoConfirmEast)
	}
}

func (c *KyarabenConfig) validateAndFilter(v *ConfigValidators) ConfigWarnings {
	var warnings ConfigWarnings

	newSystems := make(map[SystemID][]EmulatorID)
	for sysID, emulatorIDs := range c.Systems {
		if _, err := v.GetSystem(sysID); err != nil {
			warnings = append(warnings, ConfigWarning{
				Field:   fmt.Sprintf("systems.%s", sysID),
				Message: fmt.Sprintf("unknown system %q, skipping", sysID),
			})
			continue
		}

		var validEmulators []EmulatorID
		for _, emuID := range emulatorIDs {
			if _, err := v.GetEmulator(emuID); err != nil {
				warnings = append(warnings, ConfigWarning{
					Field:   fmt.Sprintf("systems.%s", sysID),
					Message: fmt.Sprintf("unknown emulator %q, skipping", emuID),
				})
				continue
			}
			validEmulators = append(validEmulators, emuID)
		}
		if len(validEmulators) > 0 {
			newSystems[sysID] = validEmulators
		}
	}
	c.Systems = newSystems

	newFrontends := make(map[FrontendID]FrontendConfig)
	for feID, feConf := range c.Frontends {
		if _, err := v.GetFrontend(feID); err != nil {
			warnings = append(warnings, ConfigWarning{
				Field:   fmt.Sprintf("frontends.%s", feID),
				Message: fmt.Sprintf("unknown frontend %q, skipping", feID),
			})
			continue
		}
		newFrontends[feID] = feConf
	}
	c.Frontends = newFrontends

	return warnings
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

func LoadConfigWithWarnings(path string, validators *ConfigValidators) (*LoadResult, error) {
	return NewConfigStore(vfs.OSFS).LoadWithWarnings(path, validators)
}

func SaveConfig(cfg *KyarabenConfig, path string) error {
	return NewConfigStore(vfs.OSFS).Save(cfg, path)
}

func (c *KyarabenConfig) ExpandCollectionWith(homeDir string) string {
	path := c.Global.Collection
	if len(path) > 0 && path[0] == '~' && homeDir != "" {
		path = filepath.Join(homeDir, path[1:])
	}
	return path
}

func (c *KyarabenConfig) ExpandCollection() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expanding home directory: %w", err)
	}
	return c.ExpandCollectionWith(home), nil
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
			Collection: DefaultCollection(),
		},
		Sync: DefaultSyncConfig(),
		Controller: ControllerTomlConfig{
			NintendoConfirm: string(NintendoConfirmEast),
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

// BuildVersionOverrides returns a map from package names (or core names for
// RetroArch cores) to pinned versions based on the emulator and frontend
// versions configured in the config.
func (c *KyarabenConfig) BuildVersionOverrides(
	getEmulator func(EmulatorID) (Emulator, error),
	getFrontend func(FrontendID) (Frontend, error),
) (map[string]string, error) {
	overrides := make(map[string]string)
	for emuID, emuConf := range c.Emulators {
		if emuConf.Version == "" {
			continue
		}
		if coreName := emuID.RetroArchCoreName(); coreName != "" {
			overrides[coreName] = emuConf.Version
		} else {
			emu, err := getEmulator(emuID)
			if err != nil {
				return nil, fmt.Errorf("unknown emulator %q: %w", emuID, err)
			}
			overrides[emu.Package.PackageName()] = emuConf.Version
		}
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

package rpcs3

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/fnune/kyaraben/internal/model"
)

type Definition struct{}

func (Definition) Emulator() model.Emulator {
	return model.Emulator{
		ID:      model.EmulatorIDRPCS3,
		Name:    "RPCS3",
		Systems: []model.SystemID{model.SystemIDPS3},
		Package: model.AppImageRef("rpcs3"),
		ProvisionGroups: []model.ProvisionGroup{{
			MinRequired: 1,
			Message:     "Firmware required (provides system libraries and OS)",
			BaseDir: func(store model.StoreReader, sys model.SystemID) string {
				configDir, err := os.UserConfigDir()
				if err != nil {
					return ""
				}
				return filepath.Join(configDir, "rpcs3")
			},
			Provisions: []model.Provision{{
				Kind:        model.ProvisionFirmware,
				Description: "Official firmware",
				Strategy: model.ImportStrategy{
					Pattern:             "dev_flash/sys/*",
					Filename:            "PS3UPDAT.PUP",
					VerifiedDescription: "firmware installed",
					Instructions:        "Import via File > Install Firmware",
				},
				ImportViaUI: true,
			}},
		}},
		StateKinds: []model.StateKind{
			model.StateSaves,
			model.StateSavestates,
			model.StateScreenshots,
		},
		Launcher: model.LauncherInfo{
			Binary:      "rpcs3",
			GenericName: "PlayStation 3 Emulator",
			Categories:  []string{"Game", "Emulator"},
			RomCommand: func(opts model.RomLaunchOptions) string {
				cmd := opts.BinaryPath + " --no-gui"
				if opts.Fullscreen {
					cmd += " --fullscreen"
				}
				cmd += " %ROM%"
				return cmd
			},
		},
		PathUsage: model.PathUsage{
			UsesSavesDir:       true,
			UsesStatesDir:      true,
			UsesScreenshotsDir: true,
		},
	}
}

func (Definition) ConfigGenerator() model.ConfigGenerator {
	return &Config{}
}

var vfsTarget = model.ConfigTarget{
	RelPath: "rpcs3/vfs.yml",
	Format:  model.ConfigFormatYAML,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var guiTarget = model.ConfigTarget{
	RelPath: "rpcs3/GuiConfigs/CurrentSettings.ini",
	Format:  model.ConfigFormatINI,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var inputProfileTarget = model.ConfigTarget{
	RelPath: "rpcs3/input_configs/global/Default.yml",
	Format:  model.ConfigFormatRaw,
	BaseDir: model.ConfigBaseDirUserConfig,
}

var activeProfilesTarget = model.ConfigTarget{
	RelPath: "rpcs3/input_configs/active_profiles.yml",
	Format:  model.ConfigFormatRaw,
	BaseDir: model.ConfigBaseDirUserConfig,
}

const activeProfilesYAML = `Active Profiles:
  global: Default
`

func generateInputProfileYAML() string {
	return generatePlayersYAML(4) + generateNullPlayersYAML(5, 7)
}

func generatePlayersYAML(count int) string {
	var result string
	for i := 0; i < count; i++ {
		result += generatePlayerYAML(i + 1)
	}
	return result
}

func generatePlayerYAML(playerNum int) string {
	return `Player ` + itoa(playerNum) + ` Input:
  Handler: Evdev
  Device: Microsoft X-Box 360 pad ` + itoa(playerNum-1) + `
  Config:
    Left Stick Left: LX-
    Left Stick Down: LY+
    Left Stick Right: LX+
    Left Stick Up: LY-
    Right Stick Left: RX-
    Right Stick Down: RY+
    Right Stick Right: RX+
    Right Stick Up: RY-
    Start: Start
    Select: Select
    PS Button: Mode
    Square: X
    Cross: A
    Circle: B
    Triangle: Y
    Left: Hat0 X-
    Down: Hat0 Y+
    Right: Hat0 X+
    Up: Hat0 Y-
    R1: TR
    R2: RZ+
    R3: Thumb R
    L1: TL
    L2: LZ+
    L3: Thumb L
    Motion Sensor X:
      Axis: X
      Mirrored: false
      Shift: 0
    Motion Sensor Y:
      Axis: Y
      Mirrored: false
      Shift: 0
    Motion Sensor Z:
      Axis: Z
      Mirrored: false
      Shift: 0
    Motion Sensor G:
      Axis: RX
      Mirrored: false
      Shift: 0
    Pressure Intensity Button: ""
    Pressure Intensity Percent: 50
    Left Stick Multiplier: 100
    Right Stick Multiplier: 100
    Left Stick Deadzone: 30
    Right Stick Deadzone: 30
    Left Trigger Threshold: 0
    Right Trigger Threshold: 0
    Left Pad Squircling Factor: 5000
    Right Pad Squircling Factor: 5000
    Color Value R: 0
    Color Value G: 0
    Color Value B: 20
    Blink LED when battery is below 20%: true
    Use LED as a battery indicator: false
    LED battery indicator brightness: 10
    Enable Large Vibration Motor: true
    Enable Small Vibration Motor: true
    Switch Vibration Motors: false
    Mouse Movement Mode: Relative
    Mouse Deadzone X Axis: 60
    Mouse Deadzone Y Axis: 60
    Mouse Acceleration X Axis: 200
    Mouse Acceleration Y Axis: 250
    Left Stick Lerp Factor: 100
    Right Stick Lerp Factor: 100
    Analog Button Lerp Factor: 100
    Trigger Lerp Factor: 100
    Device Class Type: 0
    Vendor ID: 1356
    Product ID: 616
  Buddy Device: "Null"
`
}

func generateNullPlayersYAML(start, end int) string {
	var result string
	for i := start; i <= end; i++ {
		result += generateNullPlayerYAML(i)
	}
	return result
}

func generateNullPlayerYAML(playerNum int) string {
	return `Player ` + itoa(playerNum) + ` Input:
  Handler: "Null"
  Device: "Null"
  Config:
    Left Stick Left: ""
    Left Stick Down: ""
    Left Stick Right: ""
    Left Stick Up: ""
    Right Stick Left: ""
    Right Stick Down: ""
    Right Stick Right: ""
    Right Stick Up: ""
    Start: ""
    Select: ""
    PS Button: ""
    Square: ""
    Cross: ""
    Circle: ""
    Triangle: ""
    Left: ""
    Down: ""
    Right: ""
    Up: ""
    R1: ""
    R2: ""
    R3: ""
    L1: ""
    L2: ""
    L3: ""
    Motion Sensor X:
      Axis: ""
      Mirrored: false
      Shift: 0
    Motion Sensor Y:
      Axis: ""
      Mirrored: false
      Shift: 0
    Motion Sensor Z:
      Axis: ""
      Mirrored: false
      Shift: 0
    Motion Sensor G:
      Axis: ""
      Mirrored: false
      Shift: 0
    Pressure Intensity Button: ""
    Pressure Intensity Percent: 50
    Left Stick Multiplier: 100
    Right Stick Multiplier: 100
    Left Stick Deadzone: 0
    Right Stick Deadzone: 0
    Left Trigger Threshold: 0
    Right Trigger Threshold: 0
    Left Pad Squircling Factor: 0
    Right Pad Squircling Factor: 0
    Color Value R: 0
    Color Value G: 0
    Color Value B: 0
    Blink LED when battery is below 20%: true
    Use LED as a battery indicator: false
    LED battery indicator brightness: 50
    Enable Large Vibration Motor: true
    Enable Small Vibration Motor: true
    Switch Vibration Motors: false
    Mouse Movement Mode: Relative
    Mouse Deadzone X Axis: 60
    Mouse Deadzone Y Axis: 60
    Mouse Acceleration X Axis: 200
    Mouse Acceleration Y Axis: 250
    Left Stick Lerp Factor: 100
    Right Stick Lerp Factor: 100
    Analog Button Lerp Factor: 100
    Trigger Lerp Factor: 100
    Device Class Type: 0
    Vendor ID: 0
    Product ID: 0
  Buddy Device: "Null"
`
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

type Config struct{}

func (c *Config) Generate(ctx model.GenerateContext) (model.GenerateResult, error) {
	store := ctx.Store

	patches := []model.ConfigPatch{
		{
			Target: vfsTarget,
			Entries: []model.ConfigEntry{
				{Path: []string{"/dev_hdd0/"}, Value: store.SystemSavesDir(model.SystemIDPS3) + "/"},
				{Path: []string{"/games/"}, Value: store.SystemRomsDir(model.SystemIDPS3)},
			},
		},
		{
			Target: guiTarget,
			Entries: []model.ConfigEntry{
				{Path: []string{"main_window", "infoBoxEnabledWelcome"}, Value: "false"},
				{Path: []string{"main_window", "confirmationBoxExitGame"}, Value: "false"},
				{Path: []string{"Meta", "checkUpdateStart"}, Value: "false"},
			},
		},
		{
			Target:  inputProfileTarget,
			Entries: []model.ConfigEntry{{Value: generateInputProfileYAML(), DefaultOnly: true}},
		},
		{
			Target:  activeProfilesTarget,
			Entries: []model.ConfigEntry{{Value: activeProfilesYAML, DefaultOnly: true}},
		},
	}

	configDir, err := ctx.BaseDirResolver.UserConfigDir()
	if err != nil {
		return model.GenerateResult{}, err
	}
	rpcs3Dir := filepath.Join(configDir, "rpcs3")

	symlinks := []model.SymlinkSpec{
		{Source: filepath.Join(rpcs3Dir, "savestates"), Target: store.EmulatorStatesDir(model.EmulatorIDRPCS3)},
		{Source: filepath.Join(rpcs3Dir, "screenshots"), Target: store.EmulatorScreenshotsDir(model.EmulatorIDRPCS3)},
	}

	return model.GenerateResult{
		Patches:  patches,
		Symlinks: symlinks,
	}, nil
}

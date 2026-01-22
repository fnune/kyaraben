package model

// ConfigFormat indicates the format of an emulator config file.
type ConfigFormat string

const (
	ConfigFormatINI  ConfigFormat = "ini"
	ConfigFormatTOML ConfigFormat = "toml"
	ConfigFormatCFG  ConfigFormat = "cfg"  // RetroArch-style key=value
	ConfigFormatXML  ConfigFormat = "xml"
	ConfigFormatJSON ConfigFormat = "json"
)

// EmulatorConfig represents a configuration file managed by kyaraben.
type EmulatorConfig struct {
	Path       string       // Absolute path to the config file
	Format     ConfigFormat // File format
	EmulatorID EmulatorID   // Which emulator this config belongs to
}

// ConfigEntry represents a single configuration key-value pair.
type ConfigEntry struct {
	Section string // Section name (for INI files), empty for flat configs
	Key     string
	Value   string
}

// ConfigPatch represents changes kyaraben wants to make to a config.
type ConfigPatch struct {
	Config  EmulatorConfig
	Entries []ConfigEntry
}

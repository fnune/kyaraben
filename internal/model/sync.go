package model

type SyncConfig struct {
	Enabled   bool             `toml:"enabled"`
	Autostart bool             `toml:"autostart"`
	Relays    []string         `toml:"relays,omitempty"`
	Syncthing SyncthingConfig  `toml:"syncthing"`
	Ignore    SyncIgnoreConfig `toml:"ignore"`
}

type SyncthingConfig struct {
	ListenPort             int    `toml:"listen_port"`
	DiscoveryPort          int    `toml:"discovery_port"`
	GUIPort                int    `toml:"gui_port"`
	RelayEnabled           bool   `toml:"relay_enabled"`
	GlobalDiscoveryEnabled bool   `toml:"global_discovery_enabled"`
	IgnoreDeleteROMs       *bool  `toml:"ignore_delete_roms,omitempty"`
	BaseURL                string `toml:"base_url,omitempty"`
}

func (c SyncthingConfig) IgnoreDeleteROMsEnabled() bool {
	return c.IgnoreDeleteROMs == nil || *c.IgnoreDeleteROMs
}

type SyncIgnoreConfig struct {
	Patterns []string `toml:"patterns"`
}

func DefaultSyncConfig() SyncConfig {
	return SyncConfig{
		Enabled:   false,
		Autostart: true,
		Syncthing: SyncthingConfig{
			ListenPort:    22100,
			DiscoveryPort: 21127,
			GUIPort:       8484,
			RelayEnabled:  true,
		},
		Ignore: SyncIgnoreConfig{
			Patterns: []string{
				"**/shader_cache/**",
				"**/cache/**",
				"**/*.tmp",
				".DS_Store",
				"Thumbs.db",
				"/installed",
			},
		},
	}
}

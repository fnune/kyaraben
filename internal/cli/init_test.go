package cli

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

func TestInitCmd_Headless_UsesDefaults(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{})
	require.NoError(t, err)
	defer cleanup()

	configPath := "/tmp/kyaraben-test/config.toml"

	ctx := &Context{
		FS:          fs,
		Paths:       paths.NewPaths(""),
		ConfigStore: model.NewConfigStore(fs),
		ConfigPath:  configPath,
	}

	cmd := &InitCmd{
		Collection: "/mnt/storage/Emulation",
		Headless:   true,
	}

	err = cmd.Run(ctx)
	require.NoError(t, err)

	data, err := fs.ReadFile(configPath)
	require.NoError(t, err)

	var cfg model.KyarabenConfig
	_, err = toml.Decode(string(data), &cfg)
	require.NoError(t, err)

	assert.True(t, cfg.Global.Headless, "headless should be true")
	assert.True(t, cfg.Sync.Enabled, "sync should be enabled")
	assert.Equal(t, "/mnt/storage/Emulation", cfg.Global.Collection)
	assert.Empty(t, cfg.Systems, "headless config should have no systems")
	assert.Equal(t, 22100, cfg.Sync.Syncthing.ListenPort, "should use default listen port")
	assert.Equal(t, 21127, cfg.Sync.Syncthing.DiscoveryPort, "should use default discovery port")
	assert.Equal(t, 8484, cfg.Sync.Syncthing.GUIPort, "should use default GUI port")
	assert.True(t, cfg.Sync.Syncthing.RelayEnabled, "relay should be enabled by default")
	assert.True(t, cfg.Sync.Autostart, "autostart should be enabled by default")
}

func TestInitCmd_Regular_UsesDefaults(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{})
	require.NoError(t, err)
	defer cleanup()

	configPath := "/tmp/kyaraben-test/config.toml"

	ctx := &Context{
		FS:          fs,
		Paths:       paths.NewPaths(""),
		ConfigStore: model.NewConfigStore(fs),
		ConfigPath:  configPath,
	}

	cmd := &InitCmd{
		Collection: "~/Emulation",
		Headless:   false,
	}

	err = cmd.Run(ctx)
	require.NoError(t, err)

	data, err := fs.ReadFile(configPath)
	require.NoError(t, err)

	var cfg model.KyarabenConfig
	_, err = toml.Decode(string(data), &cfg)
	require.NoError(t, err)

	assert.NotEmpty(t, cfg.Systems, "regular config should have default systems")
	assert.False(t, cfg.Global.Headless, "headless should be false")
	assert.Equal(t, 22100, cfg.Sync.Syncthing.ListenPort, "should use default listen port")
}

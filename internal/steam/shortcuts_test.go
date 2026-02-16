package steam

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fnune/kyaraben/internal/testutil"
)

func TestManager_IsAvailable_NoSteam(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.config/.keep": "",
	})

	manager := NewManager(fs, "/home/user")
	assert.False(t, manager.IsAvailable())
}

func TestManager_IsAvailable_WithSteam(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/.keep": "",
	})

	manager := NewManager(fs, "/home/user")
	assert.True(t, manager.IsAvailable())
}

func TestManager_Sync_NoSteam(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.config/.keep": "",
	})

	manager := NewManager(fs, "/home/user")
	err := manager.Sync([]ShortcutEntry{{
		AppName: "Test",
		Exe:     "/usr/bin/test",
	}})
	require.NoError(t, err)
}

func TestManager_Sync_CreatesShortcut(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/.keep": "",
	})

	manager := NewManager(fs, "/home/user")

	entries := []ShortcutEntry{{
		AppName:       "ES-DE",
		Exe:           "/home/user/.local/state/kyaraben/bin/es-de",
		StartDir:      "/home/user/.local/state/kyaraben/bin",
		LaunchOptions: "--fullscreen",
		Tags:          []string{"Kyaraben"},
	}}

	err := manager.Sync(entries)
	require.NoError(t, err)

	shortcutsPath := "/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf"
	data, err := fs.ReadFile(shortcutsPath)
	require.NoError(t, err)

	shortcuts, err := ParseShortcuts(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, shortcuts, 1)

	s := shortcuts[0]
	assert.Equal(t, "ES-DE", s.AppName)
	assert.Equal(t, "/home/user/.local/state/kyaraben/bin/es-de", s.Exe)
	assert.Equal(t, "/home/user/.local/state/kyaraben/bin", s.StartDir)
	assert.Equal(t, "--fullscreen", s.LaunchOptions)
	assert.Equal(t, []string{"Kyaraben"}, s.Tags)
	expectedAppID := GenerateAppID(entries[0].Exe, entries[0].AppName)
	assert.Equal(t, expectedAppID, s.AppID)
}

func TestManager_Sync_PreservesExistingShortcuts(t *testing.T) {
	existingShortcuts := []Shortcut{{
		AppID:   123456789,
		AppName: "User's Game",
		Exe:     "/home/user/games/game.exe",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteShortcuts(&buf, existingShortcuts))

	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf": buf.Bytes(),
	})

	manager := NewManager(fs, "/home/user")

	entries := []ShortcutEntry{{
		AppName: "ES-DE",
		Exe:     "/home/user/.local/state/kyaraben/bin/es-de",
	}}

	err := manager.Sync(entries)
	require.NoError(t, err)

	shortcutsPath := "/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf"
	data, err := fs.ReadFile(shortcutsPath)
	require.NoError(t, err)

	shortcuts, err := ParseShortcuts(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, shortcuts, 2)

	assert.Equal(t, "User's Game", shortcuts[0].AppName)
	assert.Equal(t, "ES-DE", shortcuts[1].AppName)
}

func TestManager_Sync_UpdatesExistingManagedShortcut(t *testing.T) {
	exe := "/home/user/.local/state/kyaraben/bin/es-de"
	appName := "ES-DE"
	appID := GenerateAppID(exe, appName)

	existingShortcuts := []Shortcut{{
		AppID:         appID,
		AppName:       appName,
		Exe:           exe,
		LaunchOptions: "--old-option",
	}}
	var buf bytes.Buffer
	require.NoError(t, WriteShortcuts(&buf, existingShortcuts))

	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf": buf.Bytes(),
	})

	manager := NewManager(fs, "/home/user")

	entries := []ShortcutEntry{{
		AppName:       appName,
		Exe:           exe,
		LaunchOptions: "--new-option",
	}}

	err := manager.Sync(entries)
	require.NoError(t, err)

	shortcutsPath := "/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf"
	data, err := fs.ReadFile(shortcutsPath)
	require.NoError(t, err)

	shortcuts, err := ParseShortcuts(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, shortcuts, 1)

	assert.Equal(t, "--new-option", shortcuts[0].LaunchOptions)
}

func TestManager_Sync_MultipleUsers(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/11111111/config/.keep": "",
		"/home/user/.steam/steam/userdata/22222222/config/.keep": "",
	})

	manager := NewManager(fs, "/home/user")

	entries := []ShortcutEntry{{
		AppName: "ES-DE",
		Exe:     "/home/user/.local/state/kyaraben/bin/es-de",
	}}

	err := manager.Sync(entries)
	require.NoError(t, err)

	for _, userID := range []string{"11111111", "22222222"} {
		shortcutsPath := "/home/user/.steam/steam/userdata/" + userID + "/config/shortcuts.vdf"
		data, err := fs.ReadFile(shortcutsPath)
		require.NoError(t, err)

		shortcuts, err := ParseShortcuts(bytes.NewReader(data))
		require.NoError(t, err)
		require.Len(t, shortcuts, 1, "user %s should have 1 shortcut", userID)
	}
}

func TestManager_Sync_WithGridAssets(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/.keep": "",
	})

	manager := NewManager(fs, "/home/user")

	gridData := []byte("PNG grid data")
	heroData := []byte("PNG hero data")

	entries := []ShortcutEntry{{
		AppName: "ES-DE",
		Exe:     "/home/user/.local/state/kyaraben/bin/es-de",
		GridAssets: &GridAssets{
			Grid: gridData,
			Hero: heroData,
		},
	}}

	err := manager.Sync(entries)
	require.NoError(t, err)

	appID := GenerateAppID(entries[0].Exe, entries[0].AppName)
	gridDir := "/home/user/.steam/steam/userdata/12345678/config/grid"

	data, err := fs.ReadFile(gridDir + "/" + itoa(appID) + ".png")
	require.NoError(t, err)
	assert.Equal(t, gridData, data)

	data, err = fs.ReadFile(gridDir + "/" + itoa(appID) + "_hero.png")
	require.NoError(t, err)
	assert.Equal(t, heroData, data)
}

func TestManager_RemoveShortcuts(t *testing.T) {
	exe := "/home/user/.local/state/kyaraben/bin/es-de"
	appName := "ES-DE"
	appID := GenerateAppID(exe, appName)

	existingShortcuts := []Shortcut{
		{
			AppID:   appID,
			AppName: appName,
			Exe:     exe,
		},
		{
			AppID:   123456789,
			AppName: "User's Game",
			Exe:     "/home/user/games/game.exe",
		},
	}
	var buf bytes.Buffer
	require.NoError(t, WriteShortcuts(&buf, existingShortcuts))

	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf": buf.Bytes(),
	})

	manager := NewManager(fs, "/home/user")

	err := manager.RemoveShortcuts([]uint32{appID})
	require.NoError(t, err)

	shortcutsPath := "/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf"
	data, err := fs.ReadFile(shortcutsPath)
	require.NoError(t, err)

	shortcuts, err := ParseShortcuts(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, shortcuts, 1)

	assert.Equal(t, "User's Game", shortcuts[0].AppName)
}

func TestManager_Sync_RemovesPreviouslyManagedShortcut(t *testing.T) {
	exe := "/home/user/.local/state/kyaraben/bin/es-de"
	appName := "ES-DE"
	appID := GenerateAppID(exe, appName)

	existingShortcuts := []Shortcut{
		{
			AppID:   appID,
			AppName: appName,
			Exe:     exe,
		},
		{
			AppID:   123456789,
			AppName: "User's Game",
			Exe:     "/home/user/games/game.exe",
		},
	}
	var buf bytes.Buffer
	require.NoError(t, WriteShortcuts(&buf, existingShortcuts))

	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf": buf.Bytes(),
	})

	manager := NewManager(fs, "/home/user")

	err := manager.Sync([]ShortcutEntry{})
	require.NoError(t, err)

	shortcutsPath := "/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf"
	data, err := fs.ReadFile(shortcutsPath)
	require.NoError(t, err)

	shortcuts, err := ParseShortcuts(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, shortcuts, 2)
}

func itoa(n uint32) string {
	return fmt.Sprintf("%d", n)
}

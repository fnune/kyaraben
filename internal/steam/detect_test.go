package steam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fnune/kyaraben/internal/testutil"
)

func TestDetector_Detect_NativeSteam(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/.keep": "",
		"/home/user/.steam/steam/userdata/87654321/config/.keep": "",
		"/home/user/.steam/steam/userdata/invalid_user/.keep":    "",
		"/home/user/.steam/steam/userdata/99999999/other/.keep":  "",
	})

	detector := NewDetector(fs, "/home/user")

	install, err := detector.Detect()
	require.NoError(t, err)
	require.NotNil(t, install)

	assert.Equal(t, "/home/user/.steam/steam", install.RootPath)
	assert.Equal(t, []string{"12345678", "87654321"}, install.UserIDs)
}

func TestDetector_Detect_FlatpakSteam(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.local/share/Steam/userdata/12345678/config/.keep": "",
	})

	detector := NewDetector(fs, "/home/user")

	install, err := detector.Detect()
	require.NoError(t, err)
	require.NotNil(t, install)

	assert.Equal(t, "/home/user/.local/share/Steam", install.RootPath)
	assert.Equal(t, []string{"12345678"}, install.UserIDs)
}

func TestDetector_Detect_NoSteam(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.config/.keep": "",
	})

	detector := NewDetector(fs, "/home/user")

	install, err := detector.Detect()
	require.NoError(t, err)
	assert.Nil(t, install)
}

func TestDetector_Detect_SteamExistsButNoUsers(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/.keep": "",
	})

	detector := NewDetector(fs, "/home/user")

	install, err := detector.Detect()
	require.NoError(t, err)
	assert.Nil(t, install)
}

func TestDetector_Detect_PrefersNativeOverFlatpak(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/11111111/config/.keep":       "",
		"/home/user/.local/share/Steam/userdata/22222222/config/.keep": "",
	})

	detector := NewDetector(fs, "/home/user")

	install, err := detector.Detect()
	require.NoError(t, err)
	require.NotNil(t, install)

	assert.Equal(t, "/home/user/.steam/steam", install.RootPath)
	assert.Equal(t, []string{"11111111"}, install.UserIDs)
}

func TestDetector_Detect_NoHomeDir(t *testing.T) {
	fs := testutil.NewTestFS(t, map[string]any{
		"/home/user/.steam/steam/userdata/12345678/config/.keep": "",
	})

	detector := NewDetector(fs, "")

	install, err := detector.Detect()
	require.NoError(t, err)
	assert.Nil(t, install)
}

func TestSteamInstall_ShortcutsPath(t *testing.T) {
	install := &SteamInstall{
		RootPath: "/home/user/.steam/steam",
		UserIDs:  []string{"12345678"},
	}

	path := install.ShortcutsPath("12345678")
	assert.Equal(t, "/home/user/.steam/steam/userdata/12345678/config/shortcuts.vdf", path)
}

func TestSteamInstall_GridDir(t *testing.T) {
	install := &SteamInstall{
		RootPath: "/home/user/.steam/steam",
		UserIDs:  []string{"12345678"},
	}

	path := install.GridDir("12345678")
	assert.Equal(t, "/home/user/.steam/steam/userdata/12345678/config/grid", path)
}

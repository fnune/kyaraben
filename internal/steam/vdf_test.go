package steam

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseShortcuts_Empty(t *testing.T) {
	shortcuts, err := ParseShortcuts(bytes.NewReader(nil))
	require.NoError(t, err)
	assert.Empty(t, shortcuts)
}

func TestParseShortcuts_SingleEntry(t *testing.T) {
	data := buildShortcutsVDF(t, []Shortcut{{
		AppID:              2684466978,
		AppName:            "ES-DE",
		Exe:                "/home/user/.local/state/kyaraben/bin/esde",
		StartDir:           "/home/user/.local/state/kyaraben/bin",
		Icon:               "",
		LaunchOptions:      "",
		IsHidden:           0,
		AllowDesktopConfig: 1,
		AllowOverlay:       1,
		Tags:               []string{"Kyaraben"},
	}})

	shortcuts, err := ParseShortcuts(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, shortcuts, 1)

	s := shortcuts[0]
	assert.Equal(t, uint32(2684466978), s.AppID)
	assert.Equal(t, "ES-DE", s.AppName)
	assert.Equal(t, "/home/user/.local/state/kyaraben/bin/esde", s.Exe)
	assert.Equal(t, "/home/user/.local/state/kyaraben/bin", s.StartDir)
	assert.Equal(t, uint32(1), s.AllowDesktopConfig)
	assert.Equal(t, uint32(1), s.AllowOverlay)
	assert.Equal(t, []string{"Kyaraben"}, s.Tags)
}

func TestParseShortcuts_MultipleEntries(t *testing.T) {
	data := buildShortcutsVDF(t, []Shortcut{
		{
			AppID:   123456789,
			AppName: "Game One",
			Exe:     "/usr/bin/game1",
			Tags:    []string{"Games", "Action"},
		},
		{
			AppID:   987654321,
			AppName: "Game Two",
			Exe:     "/usr/bin/game2",
			Tags:    nil,
		},
	})

	shortcuts, err := ParseShortcuts(bytes.NewReader(data))
	require.NoError(t, err)
	require.Len(t, shortcuts, 2)

	assert.Equal(t, "Game One", shortcuts[0].AppName)
	assert.Equal(t, []string{"Games", "Action"}, shortcuts[0].Tags)
	assert.Equal(t, "Game Two", shortcuts[1].AppName)
	assert.Empty(t, shortcuts[1].Tags)
}

func TestWriteShortcuts_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := WriteShortcuts(&buf, nil)
	require.NoError(t, err)

	shortcuts, err := ParseShortcuts(&buf)
	require.NoError(t, err)
	assert.Empty(t, shortcuts)
}

func TestWriteShortcuts_RoundTrip(t *testing.T) {
	original := []Shortcut{
		{
			AppID:              2684466978,
			AppName:            "ES-DE",
			Exe:                "/home/user/.local/state/kyaraben/bin/esde",
			StartDir:           "/home/user/.local/state/kyaraben/bin",
			Icon:               "/path/to/icon.png",
			LaunchOptions:      "--fullscreen",
			IsHidden:           0,
			AllowDesktopConfig: 1,
			AllowOverlay:       1,
			Tags:               []string{"Kyaraben", "Frontend"},
		},
		{
			AppID:              123456789,
			AppName:            "RetroArch",
			Exe:                "/usr/bin/retroarch",
			StartDir:           "/usr/bin",
			IsHidden:           1,
			AllowDesktopConfig: 0,
			AllowOverlay:       0,
			Tags:               []string{"Emulator"},
		},
	}

	var buf bytes.Buffer
	err := WriteShortcuts(&buf, original)
	require.NoError(t, err)

	parsed, err := ParseShortcuts(&buf)
	require.NoError(t, err)
	require.Len(t, parsed, 2)

	for i, orig := range original {
		p := parsed[i]
		assert.Equal(t, orig.AppID, p.AppID, "AppID mismatch at index %d", i)
		assert.Equal(t, orig.AppName, p.AppName, "AppName mismatch at index %d", i)
		assert.Equal(t, orig.Exe, p.Exe, "Exe mismatch at index %d", i)
		assert.Equal(t, orig.StartDir, p.StartDir, "StartDir mismatch at index %d", i)
		assert.Equal(t, orig.Icon, p.Icon, "Icon mismatch at index %d", i)
		assert.Equal(t, orig.LaunchOptions, p.LaunchOptions, "LaunchOptions mismatch at index %d", i)
		assert.Equal(t, orig.IsHidden, p.IsHidden, "IsHidden mismatch at index %d", i)
		assert.Equal(t, orig.AllowDesktopConfig, p.AllowDesktopConfig, "AllowDesktopConfig mismatch at index %d", i)
		assert.Equal(t, orig.AllowOverlay, p.AllowOverlay, "AllowOverlay mismatch at index %d", i)
		assert.Equal(t, orig.Tags, p.Tags, "Tags mismatch at index %d", i)
	}
}

func TestWriteShortcuts_WithAllFields(t *testing.T) {
	original := []Shortcut{{
		AppID:               2684466978,
		AppName:             "Test Game",
		Exe:                 "/path/to/game",
		StartDir:            "/path/to",
		Icon:                "/path/to/icon.png",
		ShortcutPath:        "/shortcut/path",
		LaunchOptions:       "--fullscreen --vsync",
		IsHidden:            1,
		AllowDesktopConfig:  0,
		AllowOverlay:        0,
		OpenVR:              1,
		Devkit:              1,
		DevkitGameID:        "devkit123",
		DevkitOverrideAppID: 999,
		LastPlayTime:        1700000000,
		FlatpakAppID:        "org.example.game",
		Tags:                []string{"Tag1", "Tag2", "Tag3"},
	}}

	var buf bytes.Buffer
	err := WriteShortcuts(&buf, original)
	require.NoError(t, err)

	parsed, err := ParseShortcuts(&buf)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	p := parsed[0]
	assert.Equal(t, original[0].AppID, p.AppID)
	assert.Equal(t, original[0].AppName, p.AppName)
	assert.Equal(t, original[0].Exe, p.Exe)
	assert.Equal(t, original[0].StartDir, p.StartDir)
	assert.Equal(t, original[0].Icon, p.Icon)
	assert.Equal(t, original[0].ShortcutPath, p.ShortcutPath)
	assert.Equal(t, original[0].LaunchOptions, p.LaunchOptions)
	assert.Equal(t, original[0].IsHidden, p.IsHidden)
	assert.Equal(t, original[0].AllowDesktopConfig, p.AllowDesktopConfig)
	assert.Equal(t, original[0].AllowOverlay, p.AllowOverlay)
	assert.Equal(t, original[0].OpenVR, p.OpenVR)
	assert.Equal(t, original[0].Devkit, p.Devkit)
	assert.Equal(t, original[0].DevkitGameID, p.DevkitGameID)
	assert.Equal(t, original[0].DevkitOverrideAppID, p.DevkitOverrideAppID)
	assert.Equal(t, original[0].LastPlayTime, p.LastPlayTime)
	assert.Equal(t, original[0].FlatpakAppID, p.FlatpakAppID)
	assert.Equal(t, original[0].Tags, p.Tags)
}

func TestGenerateAppID(t *testing.T) {
	tests := []struct {
		exe     string
		appName string
	}{
		{"/home/user/.local/state/kyaraben/bin/esde", "ES-DE"},
		{"/usr/bin/retroarch", "RetroArch"},
		{"/home/user/Games/game.exe", "My Game"},
	}

	for _, tc := range tests {
		appID := GenerateAppID(tc.exe, tc.appName)
		assert.True(t, appID&0x80000000 != 0, "high bit should be set")

		appID2 := GenerateAppID(tc.exe, tc.appName)
		assert.Equal(t, appID, appID2, "same inputs should produce same AppID")
	}
}

func TestGenerateAppID_DifferentInputs(t *testing.T) {
	id1 := GenerateAppID("/path/to/game1", "Game")
	id2 := GenerateAppID("/path/to/game2", "Game")
	id3 := GenerateAppID("/path/to/game1", "Different")

	assert.NotEqual(t, id1, id2, "different exe paths should produce different IDs")
	assert.NotEqual(t, id1, id3, "different app names should produce different IDs")
}

func TestParseShortcuts_InvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"wrong root type", []byte{typeString, 's', 'h', 'o', 'r', 't', 'c', 'u', 't', 's', 0}},
		{"truncated", []byte{typeMap, 's', 'h', 'o', 'r', 't'}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseShortcuts(bytes.NewReader(tc.data))
			assert.Error(t, err)
		})
	}
}

func TestWriteShortcuts_QuotesExePath(t *testing.T) {
	shortcuts := []Shortcut{{
		AppID:   123,
		AppName: "Test",
		Exe:     "/path/to/exe",
	}}

	var buf bytes.Buffer
	err := WriteShortcuts(&buf, shortcuts)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `"/path/to/exe"`)
}

func TestParseShortcuts_UnquotesExePath(t *testing.T) {
	shortcuts := []Shortcut{{
		AppID:   123,
		AppName: "Test",
		Exe:     "/path/to/exe",
	}}

	var buf bytes.Buffer
	err := WriteShortcuts(&buf, shortcuts)
	require.NoError(t, err)

	parsed, err := ParseShortcuts(&buf)
	require.NoError(t, err)
	require.Len(t, parsed, 1)
	assert.Equal(t, "/path/to/exe", parsed[0].Exe)
}

func buildShortcutsVDF(t *testing.T, shortcuts []Shortcut) []byte {
	var buf bytes.Buffer
	err := WriteShortcuts(&buf, shortcuts)
	require.NoError(t, err)
	return buf.Bytes()
}

package emulators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/emulators/cemu"
	"github.com/fnune/kyaraben/internal/emulators/dolphin"
	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/eden"
	"github.com/fnune/kyaraben/internal/emulators/flycast"
	"github.com/fnune/kyaraben/internal/emulators/melonds"
	"github.com/fnune/kyaraben/internal/emulators/mgba"
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbeetlesaturn"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/rpcs3"
	"github.com/fnune/kyaraben/internal/emulators/vita3k"
	"github.com/fnune/kyaraben/internal/model"
)

type fakeStoreReader struct {
	root string
}

func (f *fakeStoreReader) RomsDir() string {
	return filepath.Join(f.root, "roms")
}

func (f *fakeStoreReader) BiosDir() string {
	return filepath.Join(f.root, "bios")
}

func (f *fakeStoreReader) SystemBiosDir(sys model.SystemID) string {
	return filepath.Join(f.root, "bios", string(sys))
}

func (f *fakeStoreReader) SystemSavesDir(sys model.SystemID) string {
	return filepath.Join(f.root, "saves", string(sys))
}

func (f *fakeStoreReader) EmulatorSavesDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "saves", string(emu))
}

func (f *fakeStoreReader) EmulatorStatesDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "states", string(emu))
}

func (f *fakeStoreReader) EmulatorScreenshotsDir(emu model.EmulatorID) string {
	name := string(emu)
	if strings.HasPrefix(name, "retroarch:") {
		name = "retroarch"
	}
	return filepath.Join(f.root, "screenshots", name)
}

func (f *fakeStoreReader) SystemRomsDir(sys model.SystemID) string {
	return filepath.Join(f.root, "roms", string(sys))
}

func (f *fakeStoreReader) EmulatorOpaqueDir(emu model.EmulatorID) string {
	return filepath.Join(f.root, "opaque", string(emu))
}

func TestDuckStationGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := duckstation.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatINI {
		t.Errorf("expected INI format, got %s", patch.Target.Format)
	}

	if patch.Target.BaseDir != model.ConfigBaseDirUserConfig {
		t.Errorf("expected UserConfig base dir, got %s", patch.Target.BaseDir)
	}

	if !strings.Contains(patch.Target.RelPath, "duckstation") {
		t.Errorf("expected RelPath to contain 'duckstation', got %s", patch.Target.RelPath)
	}

	path, err := patch.Target.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	configDir, _ := os.UserConfigDir()
	expectedPath := filepath.Join(configDir, "duckstation", "settings.ini")
	if path != expectedPath {
		t.Errorf("resolved path = %q, want %q", path, expectedPath)
	}

	expectedKeys := map[string]bool{
		"SearchDirectory": false,
		"Directory":       false,
		"SaveStates":      false,
		"Screenshots":     false,
		"RecursivePaths":  false,
	}

	for _, entry := range patch.Entries {
		if _, ok := expectedKeys[entry.Key()]; ok {
			expectedKeys[entry.Key()] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("expected key %q not found in entries", key)
		}
	}
}

func TestRetroArchCoresGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}

	gen := retroarchbsnes.Definition{}.ConfigGenerator()
	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch (shared config), got %d", len(patches))
	}

	shared := patches[0]
	if shared.Target.RelPath != "retroarch/retroarch.cfg" {
		t.Errorf("expected shared config path, got %s", shared.Target.RelPath)
	}

	sharedKeys := collectKeys(shared.Entries)
	if !sharedKeys["libretro_directory"] {
		t.Error("shared config missing libretro_directory")
	}
	if !sharedKeys["rgui_browser_directory"] {
		t.Error("shared config missing rgui_browser_directory")
	}
}

func collectKeys(entries []model.ConfigEntry) map[string]bool {
	keys := make(map[string]bool)
	for _, e := range entries {
		keys[e.Key()] = true
	}
	return keys
}

func TestMelonDSGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := melonds.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatTOML {
		t.Errorf("expected TOML format, got %s", patch.Target.Format)
	}

	if !strings.Contains(patch.Target.RelPath, "melonDS") {
		t.Errorf("expected RelPath to contain 'melonDS', got %s", patch.Target.RelPath)
	}

	expectedKeys := map[string]bool{
		"BIOS9Path":      false,
		"BIOS7Path":      false,
		"SaveFilePath":   false,
		"SavestatePath":  false,
		"ScreenshotPath": false,
		"LastROMFolder":  false,
	}

	for _, entry := range patch.Entries {
		if _, ok := expectedKeys[entry.Key()]; ok {
			expectedKeys[entry.Key()] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("expected key %q not found in entries", key)
		}
	}
}

func TestFlycastGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := flycast.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatINI {
		t.Errorf("expected INI format, got %s", patch.Target.Format)
	}

	if !strings.Contains(patch.Target.RelPath, "flycast") {
		t.Errorf("expected RelPath to contain 'flycast', got %s", patch.Target.RelPath)
	}

	// Verify Flycast has all required path settings
	expectedKeys := map[string]bool{
		"Flycast.DataPath":        false,
		"Dreamcast.BiosPath":      false,
		"Dreamcast.ContentPath":   false,
		"Dreamcast.SavePath":      false,
		"Dreamcast.VMUPath":       false,
		"Dreamcast.SavestatePath": false,
	}

	for _, entry := range patch.Entries {
		if _, ok := expectedKeys[entry.Key()]; ok {
			expectedKeys[entry.Key()] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("expected key %q not found in entries", key)
		}
	}
}

func TestDolphinGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := dolphin.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatINI {
		t.Errorf("expected INI format, got %s", patch.Target.Format)
	}

	if patch.Target.BaseDir != model.ConfigBaseDirUserConfig {
		t.Errorf("expected user config base dir, got %s", patch.Target.BaseDir)
	}

	if !strings.Contains(patch.Target.RelPath, "dolphin-emu") {
		t.Errorf("expected RelPath to contain 'dolphin-emu', got %s", patch.Target.RelPath)
	}

	foundISOPath := false
	foundDumpPath := false
	for _, entry := range patch.Entries {
		if entry.Key() == "ISOPath0" {
			foundISOPath = true
		}
		if entry.Key() == "DumpPath" {
			foundDumpPath = true
		}
	}

	if !foundISOPath {
		t.Error("expected ISOPath0 entry not found")
	}
	if !foundDumpPath {
		t.Error("expected DumpPath entry not found")
	}
}

func TestDolphinSymlinks(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	resolver := fakeBaseDirResolver{root: "/home/user"}
	gen := dolphin.Definition{}.ConfigGenerator()

	provider, ok := gen.(model.SymlinkProvider)
	if !ok {
		t.Fatal("Dolphin config generator should implement SymlinkProvider")
	}

	specs, err := provider.Symlinks(store, resolver)
	if err != nil {
		t.Fatalf("Symlinks() error = %v", err)
	}

	if len(specs) != 4 {
		t.Fatalf("expected 4 symlink specs, got %d", len(specs))
	}

	expectedSources := map[string]bool{
		"/home/user/.local/share/dolphin-emu/GC":          false,
		"/home/user/.local/share/dolphin-emu/Wii":         false,
		"/home/user/.local/share/dolphin-emu/StateSaves":  false,
		"/home/user/.local/share/dolphin-emu/ScreenShots": false,
	}

	for _, spec := range specs {
		if _, ok := expectedSources[spec.Source]; ok {
			expectedSources[spec.Source] = true
		}
	}

	for source, found := range expectedSources {
		if !found {
			t.Errorf("expected symlink source %q not found", source)
		}
	}
}

func TestMGBAGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := mgba.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatINI {
		t.Errorf("expected INI format, got %s", patch.Target.Format)
	}

	// Verify mGBA sets the expected BIOS and Qt settings
	expectedFullPaths := map[string]bool{
		"bios":                    false,
		"gb.bios":                 false,
		"gbc.bios":                false,
		"sgb.bios":                false,
		"gba.bios":                false,
		"ports.qt.bios":           false,
		"ports.qt.gb.bios":        false,
		"ports.qt.gbc.bios":       false,
		"ports.qt.sgb.bios":       false,
		"ports.qt.gba.bios":       false,
		"ports.qt.useBios":        false,
		"ports.qt.savegamePath":   false,
		"ports.qt.savestatePath":  false,
		"ports.qt.screenshotPath": false,
		"ports.qt.showLibrary":    false,
	}

	for _, entry := range patch.Entries {
		if _, ok := expectedFullPaths[entry.FullPath()]; ok {
			expectedFullPaths[entry.FullPath()] = true
		}
	}

	for fullPath, found := range expectedFullPaths {
		if !found {
			t.Errorf("expected config entry %q not found", fullPath)
		}
	}
}

func TestRetroArchCoreOverrideContainsSystemDirectory(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := retroarchbeetlesaturn.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 2 {
		t.Fatalf("expected 2 patches (shared + override), got %d", len(patches))
	}

	override := patches[1]
	expectedPath := retroarch.CoreOverrideTarget("mednafen_saturn").RelPath
	if override.Target.RelPath != expectedPath {
		t.Errorf("expected override path %q, got %q", expectedPath, override.Target.RelPath)
	}

	for _, entry := range override.Entries {
		switch entry.Key() {
		case "savefile_directory", "savestate_directory", "screenshot_directory":
			t.Errorf("override should not contain %s (symlinks handle directories)", entry.Key())
		}
	}

	found := false
	for _, entry := range override.Entries {
		if entry.Key() == "system_directory" {
			found = true
			if !strings.HasSuffix(entry.Value, "/saturn") {
				t.Errorf("system_directory should end in /saturn, got %s", entry.Value)
			}
		}
	}
	if !found {
		t.Error("expected system_directory entry not found")
	}
}

func TestRetroArchSharedConfigEnablesSorting(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := retroarchbsnes.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) < 1 {
		t.Fatal("expected at least 1 patch")
	}

	shared := patches[0]
	if shared.Target != retroarch.MainConfigTarget {
		t.Fatalf("first patch should be shared config, got %v", shared.Target)
	}

	enabledSortSettings := map[string]bool{
		"sort_savefiles_enable":  false,
		"sort_savestates_enable": false,
	}

	disabledSortSettings := map[string]bool{
		"sort_savefiles_by_content_enable":  false,
		"sort_savestates_by_content_enable": false,
	}

	for _, entry := range shared.Entries {
		if _, ok := enabledSortSettings[entry.Key()]; ok {
			if entry.Value != "true" {
				t.Errorf("%s should be true, got %s", entry.Key(), entry.Value)
			}
			enabledSortSettings[entry.Key()] = true
		}
		if _, ok := disabledSortSettings[entry.Key()]; ok {
			if entry.Value != "false" {
				t.Errorf("%s should be false, got %s", entry.Key(), entry.Value)
			}
			disabledSortSettings[entry.Key()] = true
		}
	}

	for key, found := range enabledSortSettings {
		if !found {
			t.Errorf("missing sort setting: %s", key)
		}
	}
	for key, found := range disabledSortSettings {
		if !found {
			t.Errorf("missing sort setting: %s", key)
		}
	}
}

func TestVita3KGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := vita3k.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatYAML {
		t.Errorf("expected YAML format, got %s", patch.Target.Format)
	}

	// Vita3K now uses -c CLI arg to set config location, so config is inside opaque dir
	if patch.Target.BaseDir != model.ConfigBaseDirOpaqueDir {
		t.Errorf("expected opaque dir base dir, got %s", patch.Target.BaseDir)
	}

	if !strings.Contains(patch.Target.RelPath, "opaque/vita3k") {
		t.Errorf("expected RelPath to contain 'opaque/vita3k', got %s", patch.Target.RelPath)
	}

	foundPrefPath := false
	for _, entry := range patch.Entries {
		if entry.Key() == "pref-path" {
			foundPrefPath = true
			if !strings.Contains(entry.Value, "vita3k") {
				t.Errorf("pref-path should contain 'vita3k', got %s", entry.Value)
			}
		}
	}

	if !foundPrefPath {
		t.Error("expected pref-path entry not found")
	}

	// Test LaunchArgsProvider interface
	provider, ok := gen.(model.LaunchArgsProvider)
	if !ok {
		t.Fatal("Vita3K config generator should implement LaunchArgsProvider")
	}
	args := provider.LaunchArgs(store)
	if len(args) != 2 || args[0] != "-c" {
		t.Errorf("expected LaunchArgs to return [-c, <path>], got %v", args)
	}
}

func TestRPCS3Generate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := rpcs3.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatYAML {
		t.Errorf("expected YAML format, got %s", patch.Target.Format)
	}

	// RPCS3 uses vfs.yml for Virtual File System path mappings
	if !strings.Contains(patch.Target.RelPath, "vfs.yml") {
		t.Errorf("expected RelPath to contain 'vfs.yml', got %s", patch.Target.RelPath)
	}

	foundEmulatorDir := false
	foundGamesDir := false
	for _, entry := range patch.Entries {
		if strings.Contains(entry.Key(), "EmulatorDir") {
			foundEmulatorDir = true
			if !strings.Contains(entry.Value, "rpcs3") {
				t.Errorf("EmulatorDir should contain 'rpcs3', got %s", entry.Value)
			}
		}
		if entry.Key() == "/games/" {
			foundGamesDir = true
			if !strings.Contains(entry.Value, "roms/ps3") {
				t.Errorf("/games/ should point to ps3 roms, got %s", entry.Value)
			}
		}
	}

	if !foundEmulatorDir {
		t.Error("expected $(EmulatorDir) entry not found")
	}
	if !foundGamesDir {
		t.Error("expected /games/ entry not found")
	}
}

func TestGeneratedEntriesContainStorePaths(t *testing.T) {
	store := &fakeStoreReader{root: "/test/emulation"}
	gen := duckstation.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := patches[0]

	foundStorePath := false
	for _, entry := range patch.Entries {
		if strings.Contains(entry.Value, "/test/emulation") {
			foundStorePath = true
			break
		}
	}

	if !foundStorePath {
		t.Error("no entry contains the store path")
	}
}

func TestCemuSymlinks(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	resolver := fakeBaseDirResolver{root: "/home/user"}
	gen := cemu.Definition{}.ConfigGenerator()

	provider, ok := gen.(model.SymlinkProvider)
	if !ok {
		t.Fatal("Cemu config generator should implement SymlinkProvider")
	}

	specs, err := provider.Symlinks(store, resolver)
	if err != nil {
		t.Fatalf("Symlinks() error = %v", err)
	}

	if len(specs) != 2 {
		t.Fatalf("expected 2 symlink specs, got %d", len(specs))
	}

	expectedSources := map[string]bool{
		"/home/user/.local/share/Cemu/mlc01/usr/save/00050000": false,
		"/home/user/.local/share/Cemu/screenshots":             false,
	}

	for _, spec := range specs {
		if _, ok := expectedSources[spec.Source]; ok {
			expectedSources[spec.Source] = true
		}
	}

	for source, found := range expectedSources {
		if !found {
			t.Errorf("expected symlink source %q not found", source)
		}
	}
}

func TestEdenGenerate(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	gen := eden.Definition{}.ConfigGenerator()

	patches, err := gen.Generate(store)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]

	if patch.Target.Format != model.ConfigFormatINI {
		t.Errorf("expected INI format, got %s", patch.Target.Format)
	}

	if patch.Target.BaseDir != model.ConfigBaseDirUserConfig {
		t.Errorf("expected user config base dir, got %s", patch.Target.BaseDir)
	}

	if patch.Target.RelPath != "eden/qt-config.ini" {
		t.Errorf("expected RelPath 'eden/qt-config.ini', got %s", patch.Target.RelPath)
	}
}

func TestEdenSymlinks(t *testing.T) {
	store := &fakeStoreReader{root: "/emulation"}
	resolver := fakeBaseDirResolver{root: "/home/user"}
	gen := eden.Definition{}.ConfigGenerator()

	provider, ok := gen.(model.SymlinkProvider)
	if !ok {
		t.Fatal("Eden config generator should implement SymlinkProvider")
	}

	specs, err := provider.Symlinks(store, resolver)
	if err != nil {
		t.Fatalf("Symlinks() error = %v", err)
	}

	if len(specs) != 4 {
		t.Fatalf("expected 4 symlink specs, got %d", len(specs))
	}

	expectedSources := map[string]bool{
		"/home/user/.local/share/eden/keys":                            false,
		"/home/user/.local/share/eden/nand/system/Contents/registered": false,
		"/home/user/.local/share/eden/screenshots":                     false,
		"/home/user/.local/share/eden/nand/user/save":                  false,
	}

	for _, spec := range specs {
		if _, ok := expectedSources[spec.Source]; ok {
			expectedSources[spec.Source] = true
		}
	}

	for source, found := range expectedSources {
		if !found {
			t.Errorf("expected symlink source %q not found", source)
		}
	}
}

type fakeBaseDirResolver struct {
	root string
}

func (f fakeBaseDirResolver) UserConfigDir() (string, error) {
	return filepath.Join(f.root, ".config"), nil
}

func (f fakeBaseDirResolver) UserHomeDir() (string, error) {
	return f.root, nil
}

func (f fakeBaseDirResolver) UserDataDir() (string, error) {
	return filepath.Join(f.root, ".local", "share"), nil
}

func TestUnmanagedEntriesPreserveExisting(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewConfigWriter(fakeBaseDirResolver{root: tmpDir})

	t.Run("CFG format", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.cfg")
		if err := os.WriteFile(path, []byte("menu_driver = \"ozone\"\n"), 0644); err != nil {
			t.Fatal(err)
		}

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: path,
				Format:  model.ConfigFormatCFG,
				BaseDir: model.ConfigBaseDirOpaqueDir,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"menu_driver"}, Value: "rgui", Unmanaged: true},
				{Path: []string{"system_directory"}, Value: "/bios"},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), `menu_driver = "ozone"`) {
			t.Errorf("unmanaged entry was overwritten: %s", content)
		}
		if !strings.Contains(string(content), `system_directory = "/bios"`) {
			t.Errorf("managed entry was not written: %s", content)
		}
	})

	t.Run("CFG format fresh file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "fresh.cfg")

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: path,
				Format:  model.ConfigFormatCFG,
				BaseDir: model.ConfigBaseDirOpaqueDir,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"menu_driver"}, Value: "rgui", Unmanaged: true},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), `menu_driver = "rgui"`) {
			t.Errorf("unmanaged entry was not written to fresh file: %s", content)
		}
	})

	t.Run("INI format", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.ini")
		if err := os.WriteFile(path, []byte("[Section]\nkey = existing\n"), 0644); err != nil {
			t.Fatal(err)
		}

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: path,
				Format:  model.ConfigFormatINI,
				BaseDir: model.ConfigBaseDirOpaqueDir,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"Section", "key"}, Value: "new", Unmanaged: true},
				{Path: []string{"Section", "other"}, Value: "value"},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), "key = existing") {
			t.Errorf("unmanaged entry was overwritten: %s", content)
		}
		if !strings.Contains(string(content), "other = value") {
			t.Errorf("managed entry was not written: %s", content)
		}
	})

	t.Run("YAML format", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.yaml")
		if err := os.WriteFile(path, []byte("nested:\n  key: existing\n"), 0644); err != nil {
			t.Fatal(err)
		}

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: path,
				Format:  model.ConfigFormatYAML,
				BaseDir: model.ConfigBaseDirOpaqueDir,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"nested", "key"}, Value: "new", Unmanaged: true},
				{Path: []string{"nested", "other"}, Value: "value"},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), "key: existing") {
			t.Errorf("unmanaged entry was overwritten: %s", content)
		}
		if !strings.Contains(string(content), "other: value") {
			t.Errorf("managed entry was not written: %s", content)
		}
	})

	t.Run("XML format", func(t *testing.T) {
		path := filepath.Join(tmpDir, "test.xml")
		if err := os.WriteFile(path, []byte("<root><key>existing</key></root>"), 0644); err != nil {
			t.Fatal(err)
		}

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: path,
				Format:  model.ConfigFormatXML,
				BaseDir: model.ConfigBaseDirOpaqueDir,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"root", "key"}, Value: "new", Unmanaged: true},
				{Path: []string{"root", "other"}, Value: "value"},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), "<key>existing</key>") {
			t.Errorf("unmanaged entry was overwritten: %s", content)
		}
		if !strings.Contains(string(content), "<other>value</other>") {
			t.Errorf("managed entry was not written: %s", content)
		}
	})
}

func TestConfigTargetResolveIntegration(t *testing.T) {
	targets := []model.ConfigTarget{
		{RelPath: "test/config.ini", Format: model.ConfigFormatINI, BaseDir: model.ConfigBaseDirUserConfig},
		{RelPath: "test/config.cfg", Format: model.ConfigFormatCFG, BaseDir: model.ConfigBaseDirUserData},
		{RelPath: ".testrc", Format: model.ConfigFormatINI, BaseDir: model.ConfigBaseDirHome},
	}

	for _, target := range targets {
		path, err := target.Resolve()
		if err != nil {
			t.Errorf("Resolve() for %s failed: %v", target.RelPath, err)
			continue
		}

		if !filepath.IsAbs(path) {
			t.Errorf("Resolve() returned non-absolute path: %s", path)
		}

		if !strings.HasSuffix(path, target.RelPath) {
			t.Errorf("Resolve() path %q doesn't end with RelPath %q", path, target.RelPath)
		}
	}
}

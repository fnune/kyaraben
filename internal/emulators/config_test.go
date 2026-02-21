package emulators

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/emulators/azahar"
	"github.com/fnune/kyaraben/internal/emulators/cemu"
	"github.com/fnune/kyaraben/internal/emulators/dolphin"
	"github.com/fnune/kyaraben/internal/emulators/duckstation"
	"github.com/fnune/kyaraben/internal/emulators/eden"
	"github.com/fnune/kyaraben/internal/emulators/flycast"
	"github.com/fnune/kyaraben/internal/emulators/melonds"
	"github.com/fnune/kyaraben/internal/emulators/mgba"
	"github.com/fnune/kyaraben/internal/emulators/pcsx2"
	"github.com/fnune/kyaraben/internal/emulators/ppsspp"
	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbeetlesaturn"
	"github.com/fnune/kyaraben/internal/emulators/retroarchbsnes"
	"github.com/fnune/kyaraben/internal/emulators/rpcs3"
	"github.com/fnune/kyaraben/internal/emulators/vita3k"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/testutil"
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

func (f *fakeStoreReader) CoresDir() string {
	return "/state/cores"
}

func TestDuckStationGenerate(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := duckstation.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result.Patches))
	}

	patch := result.Patches[0]

	if patch.Target.Format != model.ConfigFormatINI {
		t.Errorf("expected INI format, got %s", patch.Target.Format)
	}

	if patch.Target.BaseDir != model.ConfigBaseDirUserConfig {
		t.Errorf("expected UserConfig base dir, got %s", patch.Target.BaseDir)
	}

	if !strings.Contains(patch.Target.RelPath, "duckstation") {
		t.Errorf("expected RelPath to contain 'duckstation', got %s", patch.Target.RelPath)
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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}

	gen := retroarchbsnes.Definition{}.ConfigGenerator()
	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 1 {
		t.Fatalf("expected 1 patch (shared config), got %d", len(result.Patches))
	}

	shared := result.Patches[0]
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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := melonds.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result.Patches))
	}

	patch := result.Patches[0]

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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := flycast.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result.Patches))
	}

	patch := result.Patches[0]

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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := dolphin.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result.Patches))
	}

	patch := result.Patches[0]

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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := dolphin.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	specs := result.Symlinks

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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := mgba.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result.Patches))
	}

	patch := result.Patches[0]

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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := retroarchbeetlesaturn.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 2 {
		t.Fatalf("expected 2 patches (shared + override), got %d", len(result.Patches))
	}

	override := result.Patches[1]
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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := retroarchbsnes.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) < 1 {
		t.Fatal("expected at least 1 patch")
	}

	shared := result.Patches[0]
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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := vita3k.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 2 {
		t.Fatalf("expected 2 patches, got %d", len(result.Patches))
	}

	configPatch := result.Patches[0]
	userPatch := result.Patches[1]

	if configPatch.Target.Format != model.ConfigFormatYAML {
		t.Errorf("expected YAML format for config, got %s", configPatch.Target.Format)
	}

	if configPatch.Target.BaseDir != model.ConfigBaseDirUserConfig {
		t.Errorf("expected user config base dir, got %s", configPatch.Target.BaseDir)
	}

	if configPatch.Target.RelPath != "Vita3K/config.yml" {
		t.Errorf("expected RelPath 'Vita3K/config.yml', got %s", configPatch.Target.RelPath)
	}

	expectedEntries := map[string]string{
		"show-welcome":      "false",
		"check-for-updates": "false",
		"user-auto-connect": "true",
		"bgm-volume":        "0",
	}

	for _, entry := range configPatch.Entries {
		expected, ok := expectedEntries[entry.Key()]
		if !ok {
			t.Errorf("unexpected config entry: %s", entry.Key())
			continue
		}
		if entry.Value != expected {
			t.Errorf("expected %s=%s, got %s", entry.Key(), expected, entry.Value)
		}
		delete(expectedEntries, entry.Key())
	}

	for key := range expectedEntries {
		t.Errorf("missing expected config entry: %s", key)
	}

	if userPatch.Target.Format != model.ConfigFormatRaw {
		t.Errorf("expected Raw format for user.xml, got %s", userPatch.Target.Format)
	}

	if userPatch.Target.BaseDir != model.ConfigBaseDirUserData {
		t.Errorf("expected user data base dir for user.xml, got %s", userPatch.Target.BaseDir)
	}

	if !strings.Contains(userPatch.Target.RelPath, "user.xml") {
		t.Errorf("expected RelPath to contain 'user.xml', got %s", userPatch.Target.RelPath)
	}

	if len(userPatch.Entries) != 1 || !strings.Contains(userPatch.Entries[0].Value, "Kyaraben") {
		t.Error("user.xml should contain 'Kyaraben' user")
	}
}

func TestVita3KSymlinks(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}

	gen := vita3k.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	specs := result.Symlinks

	if len(specs) != 3 {
		t.Fatalf("expected 3 symlink specs, got %d", len(specs))
	}

	expectedSources := map[string]bool{
		"/home/user/.local/share/Vita3K/Vita3K/ux0/user/00/savedata": false,
		"/home/user/.local/share/Vita3K/screenshots":                 false,
		"/emulation/roms/psvita/installed":                           false,
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

func TestRPCS3Generate(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := rpcs3.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 2 {
		t.Fatalf("expected 2 patches, got %d", len(result.Patches))
	}

	vfsPatch := result.Patches[0]
	guiPatch := result.Patches[1]

	if vfsPatch.Target.Format != model.ConfigFormatYAML {
		t.Errorf("expected YAML format for vfs, got %s", vfsPatch.Target.Format)
	}

	if !strings.Contains(vfsPatch.Target.RelPath, "vfs.yml") {
		t.Errorf("expected RelPath to contain 'vfs.yml', got %s", vfsPatch.Target.RelPath)
	}

	foundDevHdd0 := false
	foundGamesDir := false
	for _, entry := range vfsPatch.Entries {
		if entry.Key() == "/dev_hdd0/" {
			foundDevHdd0 = true
			if !strings.Contains(entry.Value, "saves/ps3") {
				t.Errorf("/dev_hdd0/ should point to ps3 saves, got %s", entry.Value)
			}
		}
		if entry.Key() == "/games/" {
			foundGamesDir = true
			if !strings.Contains(entry.Value, "roms/ps3") {
				t.Errorf("/games/ should point to ps3 roms, got %s", entry.Value)
			}
		}
	}

	if !foundDevHdd0 {
		t.Error("expected /dev_hdd0/ entry not found")
	}
	if !foundGamesDir {
		t.Error("expected /games/ entry not found")
	}

	if guiPatch.Target.Format != model.ConfigFormatINI {
		t.Errorf("expected INI format for gui, got %s", guiPatch.Target.Format)
	}

	if !strings.Contains(guiPatch.Target.RelPath, "CurrentSettings.ini") {
		t.Errorf("expected RelPath to contain 'CurrentSettings.ini', got %s", guiPatch.Target.RelPath)
	}
}

func TestGeneratedEntriesContainStorePaths(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/test/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := duckstation.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := result.Patches[0]

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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := cemu.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	specs := result.Symlinks

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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := eden.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(result.Patches))
	}

	patch := result.Patches[0]

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
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := eden.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	specs := result.Symlinks

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

func TestUnmanagedEntriesPreserveExisting(t *testing.T) {
	t.Parallel()

	t.Run("CFG format", func(t *testing.T) {
		t.Parallel()

		fs := testutil.NewTestFS(t, map[string]any{
			"/config/test.cfg": "menu_driver = \"ozone\"\n",
		})

		resolver := testutil.FakeResolver{ConfigDir: "/config"}
		writer := NewConfigWriter(fs, resolver)

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: "test.cfg",
				Format:  model.ConfigFormatCFG,
				BaseDir: model.ConfigBaseDirUserConfig,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"menu_driver"}, Value: "rgui", Unmanaged: true},
				{Path: []string{"system_directory"}, Value: "/bios"},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := fs.ReadFile("/config/test.cfg")
		if !strings.Contains(string(content), `menu_driver = "ozone"`) {
			t.Errorf("unmanaged entry was overwritten: %s", content)
		}
		if !strings.Contains(string(content), `system_directory = "/bios"`) {
			t.Errorf("managed entry was not written: %s", content)
		}
	})

	t.Run("CFG format fresh file", func(t *testing.T) {
		t.Parallel()

		fs := testutil.NewTestFS(t, map[string]any{
			"/config": &vfst.Dir{Perm: 0755},
		})

		resolver := testutil.FakeResolver{ConfigDir: "/config"}
		writer := NewConfigWriter(fs, resolver)

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: "fresh.cfg",
				Format:  model.ConfigFormatCFG,
				BaseDir: model.ConfigBaseDirUserConfig,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"menu_driver"}, Value: "rgui", Unmanaged: true},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := fs.ReadFile("/config/fresh.cfg")
		if !strings.Contains(string(content), `menu_driver = "rgui"`) {
			t.Errorf("unmanaged entry was not written to fresh file: %s", content)
		}
	})

	t.Run("INI format", func(t *testing.T) {
		t.Parallel()

		fs := testutil.NewTestFS(t, map[string]any{
			"/config/test.ini": "[Section]\nkey = existing\n",
		})

		resolver := testutil.FakeResolver{ConfigDir: "/config"}
		writer := NewConfigWriter(fs, resolver)

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: "test.ini",
				Format:  model.ConfigFormatINI,
				BaseDir: model.ConfigBaseDirUserConfig,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"Section", "key"}, Value: "new", Unmanaged: true},
				{Path: []string{"Section", "other"}, Value: "value"},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := fs.ReadFile("/config/test.ini")
		if !strings.Contains(string(content), "key = existing") {
			t.Errorf("unmanaged entry was overwritten: %s", content)
		}
		if !strings.Contains(string(content), "other = value") {
			t.Errorf("managed entry was not written: %s", content)
		}
	})

	t.Run("YAML format", func(t *testing.T) {
		t.Parallel()

		fs := testutil.NewTestFS(t, map[string]any{
			"/config/test.yaml": "nested:\n  key: existing\n",
		})

		resolver := testutil.FakeResolver{ConfigDir: "/config"}
		writer := NewConfigWriter(fs, resolver)

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: "test.yaml",
				Format:  model.ConfigFormatYAML,
				BaseDir: model.ConfigBaseDirUserConfig,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"nested", "key"}, Value: "new", Unmanaged: true},
				{Path: []string{"nested", "other"}, Value: "value"},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := fs.ReadFile("/config/test.yaml")
		if !strings.Contains(string(content), "key: existing") {
			t.Errorf("unmanaged entry was overwritten: %s", content)
		}
		if !strings.Contains(string(content), "other: value") {
			t.Errorf("managed entry was not written: %s", content)
		}
	})

	t.Run("XML format", func(t *testing.T) {
		t.Parallel()

		fs := testutil.NewTestFS(t, map[string]any{
			"/config/test.xml": "<root><key>existing</key></root>",
		})

		resolver := testutil.FakeResolver{ConfigDir: "/config"}
		writer := NewConfigWriter(fs, resolver)

		patch := model.ConfigPatch{
			Target: model.ConfigTarget{
				RelPath: "test.xml",
				Format:  model.ConfigFormatXML,
				BaseDir: model.ConfigBaseDirUserConfig,
			},
			Entries: []model.ConfigEntry{
				{Path: []string{"root", "key"}, Value: "new", Unmanaged: true},
				{Path: []string{"root", "other"}, Value: "value"},
			},
		}

		if _, err := writer.Apply(patch); err != nil {
			t.Fatal(err)
		}

		content, _ := fs.ReadFile("/config/test.xml")
		if !strings.Contains(string(content), "<key>existing</key>") {
			t.Errorf("unmanaged entry was overwritten: %s", content)
		}
		if !strings.Contains(string(content), "<other>value</other>") {
			t.Errorf("managed entry was not written: %s", content)
		}
	})
}

func TestConfigTargetResolve(t *testing.T) {
	t.Parallel()

	resolver := testutil.FakeResolver{
		ConfigDir: "/home/user/.config",
		HomeDir:   "/home/user",
		DataDir:   "/home/user/.local/share",
	}

	tests := []struct {
		target   model.ConfigTarget
		expected string
	}{
		{
			target:   model.ConfigTarget{RelPath: "test/config.ini", BaseDir: model.ConfigBaseDirUserConfig},
			expected: "/home/user/.config/test/config.ini",
		},
		{
			target:   model.ConfigTarget{RelPath: "test/config.cfg", BaseDir: model.ConfigBaseDirUserData},
			expected: "/home/user/.local/share/test/config.cfg",
		},
		{
			target:   model.ConfigTarget{RelPath: ".testrc", BaseDir: model.ConfigBaseDirHome},
			expected: "/home/user/.testrc",
		},
	}

	for _, tt := range tests {
		path, err := tt.target.ResolveWith(resolver)
		if err != nil {
			t.Errorf("ResolveWith() for %s failed: %v", tt.target.RelPath, err)
			continue
		}

		if path != tt.expected {
			t.Errorf("ResolveWith() = %q, want %q", path, tt.expected)
		}
	}
}

func defaultControllerConfig() *model.ControllerConfig {
	return &model.ControllerConfig{
		Layout:  model.LayoutStandard,
		Hotkeys: model.DefaultHotkeys(),
		GUIDs:   model.BuiltinGUIDs,
	}
}

func TestDuckStationControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := duckstation.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := result.Patches[0]
	keys := collectKeys(patch.Entries)

	for _, key := range []string{"Cross", "Circle", "Square", "Triangle", "L1", "R1", "Start", "Select"} {
		if !keys[key] {
			t.Errorf("missing pad key %q", key)
		}
	}
	for _, key := range []string{"SaveSelectedSaveState", "LoadSelectedSaveState", "ToggleFastForward"} {
		if !keys[key] {
			t.Errorf("missing hotkey %q", key)
		}
	}

	foundPad1Cross := false
	for _, entry := range patch.Entries {
		if entry.Key() == "Cross" && strings.Contains(strings.Join(entry.Path, "."), "Pad1") {
			foundPad1Cross = true
			if !strings.HasPrefix(entry.Value, "SDL-0/") {
				t.Errorf("Pad1 Cross should start with SDL-0/, got %s", entry.Value)
			}
		}
	}
	if !foundPad1Cross {
		t.Error("Pad1.Cross entry not found")
	}
}

func TestPCSX2ControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := pcsx2.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := result.Patches[0]
	keys := collectKeys(patch.Entries)
	if !keys["SaveStateToSlot"] {
		t.Error("missing PCSX2 hotkey SaveStateToSlot")
	}
	if !keys["ToggleTurbo"] {
		t.Error("missing PCSX2 hotkey ToggleTurbo")
	}
}

func TestDolphinControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := dolphin.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 3 {
		t.Fatalf("expected 3 patches (Dolphin.ini, GCPadNew.ini, Hotkeys.ini), got %d", len(result.Patches))
	}

	gcPad := result.Patches[1]
	if !strings.Contains(gcPad.Target.RelPath, "GCPadNew") {
		t.Errorf("second patch should be GCPadNew.ini, got %s", gcPad.Target.RelPath)
	}
	keys := collectKeys(gcPad.Entries)
	if !keys["Buttons/A"] {
		t.Error("missing GCPad Buttons/A")
	}

	hotkeys := result.Patches[2]
	if !strings.Contains(hotkeys.Target.RelPath, "Hotkeys") {
		t.Errorf("third patch should be Hotkeys.ini, got %s", hotkeys.Target.RelPath)
	}
}

func TestMGBAControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := mgba.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := result.Patches[0]
	keys := collectKeys(patch.Entries)
	for _, key := range []string{"keyA", "keyB", "keyL", "keyR", "keySelect", "keyStart"} {
		if !keys[key] {
			t.Errorf("missing mGBA pad key %q", key)
		}
	}
}

func TestMelonDSControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := melonds.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := result.Patches[0]
	keys := collectKeys(patch.Entries)
	for _, key := range []string{"Joy_A", "Joy_B", "Joy_X", "Joy_Y", "Joy_L", "Joy_R"} {
		if !keys[key] {
			t.Errorf("missing melonDS pad key %q", key)
		}
	}
}

func TestPPSSPPControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := ppsspp.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 2 {
		t.Fatalf("expected 2 patches (ppsspp.ini, controls.ini), got %d", len(result.Patches))
	}

	controls := result.Patches[1]
	if !strings.Contains(controls.Target.RelPath, "controls.ini") {
		t.Errorf("second patch should be controls.ini, got %s", controls.Target.RelPath)
	}

	keys := collectKeys(controls.Entries)
	if !keys["Cross"] {
		t.Error("missing PPSSPP Cross mapping")
	}
	if !keys["Save State"] {
		t.Error("missing PPSSPP Save State hotkey")
	}
}

func TestFlycastControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := flycast.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Patches) != 2 {
		t.Fatalf("expected 2 patches (emu.cfg, mapping), got %d", len(result.Patches))
	}

	mapping := result.Patches[1]
	if !strings.Contains(mapping.Target.RelPath, "mappings") {
		t.Errorf("second patch should be mappings file, got %s", mapping.Target.RelPath)
	}

	found := false
	for _, entry := range mapping.Entries {
		if strings.Contains(entry.Value, "btn_a") {
			found = true
			break
		}
	}
	if !found {
		t.Error("mapping should contain btn_a binding")
	}
}

func TestEdenControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := eden.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := result.Patches[0]
	foundButtonA := false
	for _, entry := range patch.Entries {
		if strings.Contains(entry.Key(), "button_a") && strings.Contains(entry.Value, "engine:sdl") {
			foundButtonA = true
			break
		}
	}
	if !foundButtonA {
		t.Error("Eden should have GUID-based button_a entry")
	}
}

func TestAzaharControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := azahar.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := result.Patches[0]
	foundProfile := false
	foundButtonA := false
	for _, entry := range patch.Entries {
		if entry.Key() == "profile" {
			foundProfile = true
		}
		if strings.Contains(entry.Key(), "button_a") && strings.Contains(entry.Value, "engine:sdl") {
			foundButtonA = true
		}
	}
	if !foundProfile {
		t.Error("Azahar should set active profile")
	}
	if !foundButtonA {
		t.Error("Azahar should have GUID-based button_a entry")
	}
}

func TestRetroArchControllerConfig(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := retroarchbsnes.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{
		Store:            store,
		BaseDirResolver:  resolver,
		ControllerConfig: defaultControllerConfig(),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	shared := result.Patches[0]
	keys := collectKeys(shared.Entries)
	if !keys["input_joypad_driver"] {
		t.Error("missing input_joypad_driver in shared config")
	}
	if !keys["input_autodetect_enable"] {
		t.Error("missing input_autodetect_enable in shared config")
	}
	if !keys["input_enable_hotkey_btn"] {
		t.Error("missing input_enable_hotkey_btn in shared config")
	}
}

func TestNintendoLayoutSwapsFaceButtons(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}

	standardCC := &model.ControllerConfig{
		Layout:  model.LayoutStandard,
		Hotkeys: model.DefaultHotkeys(),
		GUIDs:   model.BuiltinGUIDs,
	}
	nintendoCC := &model.ControllerConfig{
		Layout:  model.LayoutNintendo,
		Hotkeys: model.DefaultHotkeys(),
		GUIDs:   model.BuiltinGUIDs,
	}

	gen := duckstation.Definition{}.ConfigGenerator()

	standardResult, _ := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver, ControllerConfig: standardCC})
	nintendoResult, _ := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver, ControllerConfig: nintendoCC})

	findPad1Value := func(entries []model.ConfigEntry, key string) string {
		for _, e := range entries {
			if len(e.Path) == 2 && e.Path[0] == "Pad1" && e.Path[1] == key {
				return e.Value
			}
		}
		return ""
	}

	stdCross := findPad1Value(standardResult.Patches[0].Entries, "Cross")
	ninCross := findPad1Value(nintendoResult.Patches[0].Entries, "Cross")

	if stdCross == ninCross {
		t.Errorf("Nintendo layout should produce different Cross mapping, both got %s", stdCross)
	}
}

func TestNoControllerConfigWhenNil(t *testing.T) {
	t.Parallel()

	store := &fakeStoreReader{root: "/emulation"}
	resolver := testutil.FakeResolver{ConfigDir: "/home/user/.config", HomeDir: "/home/user", DataDir: "/home/user/.local/share"}
	gen := duckstation.Definition{}.ConfigGenerator()

	result, err := gen.Generate(model.GenerateContext{Store: store, BaseDirResolver: resolver})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	patch := result.Patches[0]
	keys := collectKeys(patch.Entries)
	if keys["Cross"] || keys["Circle"] {
		t.Error("controller entries should not be present when ControllerConfig is nil")
	}
}

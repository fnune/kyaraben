package emulators

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
)

type fakeResolver struct {
	configDir string
	dataDir   string
	homeDir   string
}

func (r fakeResolver) UserConfigDir() (string, error) { return r.configDir, nil }
func (r fakeResolver) UserDataDir() (string, error)   { return r.dataDir, nil }
func (r fakeResolver) UserHomeDir() (string, error)   { return r.homeDir, nil }

func sha256sum(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func TestComputeDiff_NewFile(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"audio", "volume"}, Value: "80"},
		},
	}

	diff, err := dc.ComputeDiff(patch)
	if err != nil {
		t.Fatalf("ComputeDiff: %v", err)
	}

	if !diff.IsNewFile {
		t.Error("expected IsNewFile to be true")
	}
	if len(diff.Changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(diff.Changes))
	}
	for _, c := range diff.Changes {
		if c.Type != ChangeAdd {
			t.Errorf("expected ChangeAdd, got %v", c.Type)
		}
	}
}

func TestComputeDiff_NoChanges(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": "[video]\nresolution = 1920x1080\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
		},
	}

	diff, err := dc.ComputeDiff(patch)
	if err != nil {
		t.Fatalf("ComputeDiff: %v", err)
	}

	if diff.IsNewFile {
		t.Error("expected IsNewFile to be false")
	}
	if diff.HasChanges() {
		t.Error("expected no changes")
	}
}

func TestComputeDiff_Modifications(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": "[video]\nresolution = 1280x720\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
		},
	}

	diff, err := dc.ComputeDiff(patch)
	if err != nil {
		t.Fatalf("ComputeDiff: %v", err)
	}

	if diff.IsNewFile {
		t.Error("expected IsNewFile to be false")
	}
	if !diff.HasChanges() {
		t.Error("expected changes")
	}
	if len(diff.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(diff.Changes))
	}

	c := diff.Changes[0]
	if c.Type != ChangeModify {
		t.Errorf("expected ChangeModify, got %v", c.Type)
	}
	if c.OldValue != "1280x720" {
		t.Errorf("expected old value 1280x720, got %s", c.OldValue)
	}
	if c.NewValue != "1920x1080" {
		t.Errorf("expected new value 1920x1080, got %s", c.NewValue)
	}
}

func TestComputeDiff_AddToExistingFile(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": "[video]\nresolution = 1920x1080\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"audio", "volume"}, Value: "80"},
		},
	}

	diff, err := dc.ComputeDiff(patch)
	if err != nil {
		t.Fatalf("ComputeDiff: %v", err)
	}

	if len(diff.Changes) != 1 {
		t.Fatalf("expected 1 change (the add), got %d", len(diff.Changes))
	}
	if diff.Changes[0].Type != ChangeAdd {
		t.Errorf("expected ChangeAdd, got %v", diff.Changes[0].Type)
	}
}

func TestComputeDiff_Stats(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": "[video]\nresolution = 1280x720\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"audio", "volume"}, Value: "80"},
		},
	}

	diff, err := dc.ComputeDiff(patch)
	if err != nil {
		t.Fatalf("ComputeDiff: %v", err)
	}

	adds, modifies, removes := diff.Stats()
	if adds != 1 {
		t.Errorf("expected 1 add, got %d", adds)
	}
	if modifies != 1 {
		t.Errorf("expected 1 modify, got %d", modifies)
	}
	if removes != 0 {
		t.Errorf("expected 0 removes, got %d", removes)
	}
}

func TestComputeDiffWithBaseline_DetectsUserModifiedKeys(t *testing.T) {
	originalContent := "[video]\nresolution = 1920x1080\nmonitor = 1\n"
	modifiedContent := "[video]\nresolution = 1920x1080\nmonitor = 2\n"

	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": modifiedContent,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		BaselineHash: sha256sum(originalContent),
		ManagedKeys: []model.ManagedKey{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"video", "monitor"}, Value: "1"},
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"video", "monitor"}, Value: "1"},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.UserModified {
		t.Error("expected UserModified to be true")
	}
	if len(diff.UserChanges) != 1 {
		t.Fatalf("expected 1 user change, got %d", len(diff.UserChanges))
	}

	uc := diff.UserChanges[0]
	if uc.Path[len(uc.Path)-1] != "monitor" {
		t.Errorf("expected changed key monitor, got %s", uc.Path[len(uc.Path)-1])
	}
	if uc.BaselineValue != "1" {
		t.Errorf("expected baseline value 1, got %s", uc.BaselineValue)
	}
	if uc.CurrentValue != "2" {
		t.Errorf("expected current value 2, got %s", uc.CurrentValue)
	}
}

func TestComputeDiffWithBaseline_NoChangesWhenHashMatches(t *testing.T) {
	content := "[video]\nresolution = 1920x1080\n"

	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": content,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		BaselineHash: sha256sum(content),
		ManagedKeys: []model.ManagedKey{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if diff.UserModified {
		t.Error("expected UserModified to be false when hash matches")
	}
	if len(diff.UserChanges) != 0 {
		t.Errorf("expected 0 user changes, got %d", len(diff.UserChanges))
	}
}

func TestComputeDiffWithBaseline_NilBaseline(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": "[video]\nresolution = 1280x720\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, nil)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if diff.UserModified {
		t.Error("expected UserModified to be false with nil baseline")
	}
	if !diff.HasChanges() {
		t.Error("expected changes from diff")
	}
}

func TestFormatWithColor_PlainText(t *testing.T) {
	diff := &ConfigDiff{
		Path:      "/config/test.ini",
		IsNewFile: true,
		Changes: []ConfigChange{
			{Type: ChangeAdd, Path: []string{"video", "resolution"}, NewValue: "1920x1080"},
		},
	}

	output := diff.FormatWithColor(false)

	if output == "" {
		t.Error("expected non-empty output")
	}
	if !containsSubstring(output, "CREATE") {
		t.Errorf("expected CREATE in output, got: %s", output)
	}
	if !containsSubstring(output, "/config/test.ini") {
		t.Errorf("expected path in output, got: %s", output)
	}
}

func TestFormatWithColor_ModifiedFileWithUserChanges(t *testing.T) {
	diff := &ConfigDiff{
		Path:         "/config/test.ini",
		IsNewFile:    false,
		UserModified: true,
		UserChanges: []UserChange{
			{Path: []string{"video", "monitor"}, BaselineValue: "1", CurrentValue: "2"},
		},
		Changes: []ConfigChange{
			{Type: ChangeModify, Path: []string{"video", "resolution"}, OldValue: "1280x720", NewValue: "1920x1080"},
		},
	}

	output := diff.FormatWithColor(false)

	if !containsSubstring(output, "MODIFY") {
		t.Errorf("expected MODIFY in output, got: %s", output)
	}
	if !containsSubstring(output, "will be overwritten") {
		t.Errorf("expected overwrite warning in output, got: %s", output)
	}
	if !containsSubstring(output, "monitor") {
		t.Errorf("expected monitor key in output, got: %s", output)
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestComputeDiff_TildePathNormalization(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": "[paths]\nsaves = /home/testuser/Emulation/saves\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config", homeDir: "/home/testuser"}
	dc := NewDiffComputer(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"paths", "saves"}, Value: "~/Emulation/saves"},
		},
	}

	diff, err := dc.ComputeDiff(patch)
	if err != nil {
		t.Fatalf("ComputeDiff: %v", err)
	}

	if len(diff.Changes) != 0 {
		t.Errorf("expected no changes (tilde should match expanded path), got %d: %v", len(diff.Changes), diff.Changes)
	}
}

func TestComputeDiffWithBaseline_TildePathNormalization(t *testing.T) {
	originalContent := "[paths]\nsaves = ~/Emulation/saves\n"
	modifiedContent := "[paths]\nsaves = /home/testuser/Emulation/saves\n"

	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/test.ini": modifiedContent,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config", homeDir: "/home/testuser"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		BaselineHash: sha256sum(originalContent),
		ManagedKeys: []model.ManagedKey{
			{Path: []string{"paths", "saves"}, Value: "~/Emulation/saves"},
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"paths", "saves"}, Value: "~/Emulation/saves"},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if len(diff.UserChanges) != 0 {
		t.Errorf("expected no user changes (tilde should match expanded path), got %d: %v", len(diff.UserChanges), diff.UserChanges)
	}
}

func TestComputeDiffWithBaseline_DetectsUserModifiedKeys_XML(t *testing.T) {
	originalContent := `<settings>
  <video>
    <resolution>1920x1080</resolution>
    <monitor>1</monitor>
  </video>
</settings>`
	modifiedContent := `<settings>
  <video>
    <resolution>1920x1080</resolution>
    <monitor>2</monitor>
  </video>
</settings>`

	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/settings.xml": modifiedContent,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		BaselineHash: sha256sum(originalContent),
		ManagedKeys: []model.ManagedKey{
			{Path: []string{"settings", "video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"settings", "video", "monitor"}, Value: "1"},
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "settings.xml",
			Format:  model.ConfigFormatXML,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"settings", "video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"settings", "video", "monitor"}, Value: "1"},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.UserModified {
		t.Error("expected UserModified to be true")
	}
	if len(diff.UserChanges) != 1 {
		t.Fatalf("expected 1 user change, got %d", len(diff.UserChanges))
	}

	uc := diff.UserChanges[0]
	if uc.Path[len(uc.Path)-1] != "monitor" {
		t.Errorf("expected changed key monitor, got %s", uc.Path[len(uc.Path)-1])
	}
}

func TestComputeDiffWithBaseline_DetectsUserModifiedKeys_YAML(t *testing.T) {
	originalContent := `video:
  resolution: 1920x1080
  monitor: "1"
`
	modifiedContent := `video:
  resolution: 1920x1080
  monitor: "2"
`

	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/config.yaml": modifiedContent,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		BaselineHash: sha256sum(originalContent),
		ManagedKeys: []model.ManagedKey{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"video", "monitor"}, Value: "1"},
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "config.yaml",
			Format:  model.ConfigFormatYAML,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"video", "monitor"}, Value: "1"},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.UserModified {
		t.Error("expected UserModified to be true")
	}
	if len(diff.UserChanges) != 1 {
		t.Fatalf("expected 1 user change, got %d", len(diff.UserChanges))
	}
}

func TestComputeDiffWithBaseline_DetectsUserModifiedKeys_TOML(t *testing.T) {
	originalContent := `[video]
resolution = "1920x1080"
monitor = "1"
`
	modifiedContent := `[video]
resolution = "1920x1080"
monitor = "2"
`

	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/config.toml": modifiedContent,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		BaselineHash: sha256sum(originalContent),
		ManagedKeys: []model.ManagedKey{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"video", "monitor"}, Value: "1"},
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "config.toml",
			Format:  model.ConfigFormatTOML,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"video", "resolution"}, Value: "1920x1080"},
			{Path: []string{"video", "monitor"}, Value: "1"},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.UserModified {
		t.Error("expected UserModified to be true")
	}
	if len(diff.UserChanges) != 1 {
		t.Fatalf("expected 1 user change, got %d", len(diff.UserChanges))
	}
}

func TestComputeDiffWithBaseline_DetectsUserModifiedKeys_CFG(t *testing.T) {
	originalContent := `savefile_directory = "/home/user/saves"
savestate_directory = "/home/user/states"
`
	modifiedContent := `savefile_directory = "/home/user/saves"
savestate_directory = "/custom/states"
`

	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config/retroarch.cfg": modifiedContent,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	resolver := fakeResolver{configDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		BaselineHash: sha256sum(originalContent),
		ManagedKeys: []model.ManagedKey{
			{Path: []string{"savefile_directory"}, Value: "/home/user/saves"},
			{Path: []string{"savestate_directory"}, Value: "/home/user/states"},
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "retroarch.cfg",
			Format:  model.ConfigFormatCFG,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"savefile_directory"}, Value: "/home/user/saves"},
			{Path: []string{"savestate_directory"}, Value: "/home/user/states"},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.UserModified {
		t.Error("expected UserModified to be true")
	}
	if len(diff.UserChanges) != 1 {
		t.Fatalf("expected 1 user change, got %d", len(diff.UserChanges))
	}

	uc := diff.UserChanges[0]
	if uc.Path[len(uc.Path)-1] != "savestate_directory" {
		t.Errorf("expected changed key savestate_directory, got %s", uc.Path[len(uc.Path)-1])
	}
	if uc.BaselineValue != "/home/user/states" {
		t.Errorf("expected baseline value /home/user/states, got %s", uc.BaselineValue)
	}
	if uc.CurrentValue != `"/custom/states"` {
		t.Errorf("expected current value \"/custom/states\", got %s", uc.CurrentValue)
	}
}

package emulators

import (
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestComputeDiff_NewFile(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
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
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1920x1080\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
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
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1280x720\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
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
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1920x1080\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
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
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1280x720\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
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
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1920x1080\nmonitor = 2\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{
			"video.resolution": "1920x1080",
			"video.monitor":    "1",
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

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, nil)
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
	if uc.Key != "monitor" {
		t.Errorf("expected changed key monitor, got %s", uc.Key)
	}
	if uc.WrittenValue != "1" {
		t.Errorf("expected written value 1, got %s", uc.WrittenValue)
	}
	if uc.CurrentValue != "2" {
		t.Errorf("expected current value 2, got %s", uc.CurrentValue)
	}
}

func TestComputeDiffWithBaseline_NoChangesWhenValuesMatch(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1920x1080\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{
			"video.resolution": "1920x1080",
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

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, nil)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if diff.UserModified {
		t.Error("expected UserModified to be false when values match")
	}
	if len(diff.UserChanges) != 0 {
		t.Errorf("expected 0 user changes, got %d", len(diff.UserChanges))
	}
}

func TestComputeDiffWithBaseline_NilBaseline(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1280x720\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
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

	diff, err := dc.ComputeDiffWithBaseline(patch, nil, nil)
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
	t.Parallel()
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
	t.Parallel()
	diff := &ConfigDiff{
		Path:         "/config/test.ini",
		IsNewFile:    false,
		UserModified: true,
		UserChanges: []UserChange{
			{Key: "monitor", WrittenValue: "1", CurrentValue: "2"},
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
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[paths]\nsaves = /home/testuser/Emulation/saves\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config", HomeDir: "/home/testuser"}
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
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[paths]\nsaves = /home/testuser/Emulation/saves\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config", HomeDir: "/home/testuser"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{
			"paths.saves": "~/Emulation/saves",
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

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, nil)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if len(diff.UserChanges) != 0 {
		t.Errorf("expected no user changes (tilde should match expanded path), got %d: %v", len(diff.UserChanges), diff.UserChanges)
	}
}

func TestComputeDiffWithBaseline_DetectsVersionUpgrade(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1280x720\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{
			"video.resolution": "1280x720",
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

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, nil)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.KyarabenChanged {
		t.Error("expected KyarabenChanged to be true")
	}
	if len(diff.VersionUpgrades) != 1 {
		t.Fatalf("expected 1 version upgrade, got %d", len(diff.VersionUpgrades))
	}

	vu := diff.VersionUpgrades[0]
	if vu.Key != "resolution" {
		t.Errorf("expected changed key resolution, got %s", vu.Key)
	}
	if vu.OldValue != "1280x720" {
		t.Errorf("expected old value 1280x720, got %s", vu.OldValue)
	}
	if vu.NewValue != "1920x1080" {
		t.Errorf("expected new value 1920x1080, got %s", vu.NewValue)
	}

	if diff.UserModified {
		t.Error("expected UserModified to be false when only kyaraben changed")
	}
}

func TestComputeDiffWithBaseline_UIDrivenChangeNotReportedAsVersionUpgrade(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[Controls]\nbutton_a = 0\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{
			"Controls.button_a": "0",
		},
		ConfigInputsWhenWritten: map[string]string{
			string(model.ConfigInputNintendoConfirm): "east",
		},
	}

	ctx := &DiffContext{
		CurrentConfigInputs: map[string]string{
			string(model.ConfigInputNintendoConfirm): "south",
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{
				Path:      []string{"Controls", "button_a"},
				Value:     "1",
				DependsOn: []model.ConfigInput{model.ConfigInputNintendoConfirm},
			},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, ctx)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if diff.KyarabenChanged {
		t.Error("expected KyarabenChanged to be false for UI-driven changes")
	}
	if len(diff.VersionUpgrades) != 0 {
		t.Errorf("expected 0 version upgrades for UI-driven change, got %d", len(diff.VersionUpgrades))
	}
}

func TestComputeDiffWithBaseline_NoConfigInputsStillDetectsVersionUpgrade(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[Controls]\nbutton_a = 0\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{
			"Controls.button_a": "0",
		},
	}

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "test.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{
				Path:      []string{"Controls", "button_a"},
				Value:     "1",
				DependsOn: []model.ConfigInput{model.ConfigInputNintendoConfirm},
			},
		},
	}

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, nil)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.KyarabenChanged {
		t.Error("expected KyarabenChanged to be true when no context provided")
	}
	if len(diff.VersionUpgrades) != 1 {
		t.Errorf("expected 1 version upgrade, got %d", len(diff.VersionUpgrades))
	}
}

func TestComputeDiffWithBaseline_BothUserAndVersionChanges(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1280x720\nmonitor = 2\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{
			"video.resolution": "1280x720",
			"video.monitor":    "1",
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

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, nil)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.UserModified {
		t.Error("expected UserModified to be true")
	}
	if len(diff.UserChanges) != 1 {
		t.Fatalf("expected 1 user change, got %d", len(diff.UserChanges))
	}
	if diff.UserChanges[0].Key != "monitor" {
		t.Errorf("expected user change for monitor, got %s", diff.UserChanges[0].Key)
	}

	if !diff.KyarabenChanged {
		t.Error("expected KyarabenChanged to be true")
	}
	if len(diff.VersionUpgrades) != 1 {
		t.Fatalf("expected 1 version upgrade, got %d", len(diff.VersionUpgrades))
	}
	if diff.VersionUpgrades[0].Key != "resolution" {
		t.Errorf("expected version upgrade for resolution, got %s", diff.VersionUpgrades[0].Key)
	}
}

func TestComputeDiffWithBaseline_SameKeyUserModifiedAndVersionUpgrade(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = user_custom_value\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{
			"video.resolution": "1280x720",
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

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, nil)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if !diff.UserModified {
		t.Error("expected UserModified to be true")
	}
	if len(diff.UserChanges) != 1 {
		t.Fatalf("expected 1 user change, got %d", len(diff.UserChanges))
	}
	if diff.UserChanges[0].Key != "resolution" {
		t.Errorf("expected user change for resolution, got %s", diff.UserChanges[0].Key)
	}

	if len(diff.VersionUpgrades) != 0 {
		t.Errorf("expected 0 version upgrades when key is user-modified, got %d: %+v",
			len(diff.VersionUpgrades), diff.VersionUpgrades)
	}
}

func TestComputeDiffWithBaseline_EmptyWrittenEntries(t *testing.T) {
	t.Parallel()
	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[video]\nresolution = 1920x1080\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	dc := NewDiffComputer(fs, resolver)

	baseline := &model.ManagedConfig{
		WrittenEntries: map[string]string{},
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

	diff, err := dc.ComputeDiffWithBaseline(patch, baseline, nil)
	if err != nil {
		t.Fatalf("ComputeDiffWithBaseline: %v", err)
	}

	if diff.UserModified {
		t.Error("expected UserModified to be false with empty written entries")
	}
	if diff.KyarabenChanged {
		t.Error("expected KyarabenChanged to be false with empty written entries")
	}
}

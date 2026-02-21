package emulators

import (
	"strings"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestFileRegionWritesFromScratch(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/duckstation/inputprofiles":                        &vfst.Dir{Perm: 0755},
		"/config/duckstation/inputprofiles/kyaraben-steamdeck.ini": "[Pad1]\nCross = old-value\nOldKey = should-disappear\n",
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	writer := NewConfigWriter(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "duckstation/inputprofiles/kyaraben-steamdeck.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"Pad1", "Cross"}, Value: "new-value"},
			{Path: []string{"Pad1", "Circle"}, Value: "SDL-0/East"},
		},
		ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
	}

	result, err := writer.Apply(patch)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path == "" {
		t.Error("expected non-empty path")
	}

	content, _ := fs.ReadFile("/config/duckstation/inputprofiles/kyaraben-steamdeck.ini")
	s := string(content)

	if !strings.Contains(s, "Cross=new-value") {
		t.Errorf("should contain new Cross value, got:\n%s", s)
	}
	if !strings.Contains(s, "Circle=SDL-0/East") {
		t.Errorf("should contain Circle, got:\n%s", s)
	}
	if strings.Contains(s, "old-value") {
		t.Errorf("should not contain old values (written from scratch), got:\n%s", s)
	}
	if strings.Contains(s, "OldKey") {
		t.Errorf("should not contain old keys (written from scratch), got:\n%s", s)
	}
}

func TestFileRegionCreatesDirectories(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	writer := NewConfigWriter(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "duckstation/inputprofiles/kyaraben-steamdeck.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"Pad1", "Cross"}, Value: "SDL-0/South"},
		},
		ManagedRegions: []model.ManagedRegion{model.FileRegion{}},
	}

	_, err := writer.Apply(patch)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	content, err := fs.ReadFile("/config/duckstation/inputprofiles/kyaraben-steamdeck.ini")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if !strings.Contains(string(content), "Cross=SDL-0/South") {
		t.Errorf("should contain Cross entry, got:\n%s", content)
	}
}

func TestSectionRegionWithConfigWriter(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/azahar-emu/qt-config.ini": `[Controls]
profile = 2
profiles\size = 2
profiles\1\name = old
profiles\1\button_a = old-a
profiles\1\button_b = old-b
profiles\2\name = user-custom
profiles\2\button_a = user-a
`,
	})

	resolver := testutil.FakeResolver{ConfigDir: "/config"}
	writer := NewConfigWriter(fs, resolver)

	patch := model.ConfigPatch{
		Target: model.ConfigTarget{
			RelPath: "azahar-emu/qt-config.ini",
			Format:  model.ConfigFormatINI,
			BaseDir: model.ConfigBaseDirUserConfig,
		},
		Entries: []model.ConfigEntry{
			{Path: []string{"Controls", "profile"}, Value: "1", DefaultOnly: true},
			{Path: []string{"Controls", `profiles\size`}, Value: "1", DefaultOnly: true},
			{Path: []string{"Controls", `profiles\1\name`}, Value: "kyaraben-steamdeck"},
			{Path: []string{"Controls", `profiles\1\button_a`}, Value: "new-a"},
		},
		ManagedRegions: []model.ManagedRegion{
			model.SectionRegion{Section: "Controls", KeyPrefix: `profiles\1\`},
		},
	}

	_, err := writer.Apply(patch)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	content, _ := fs.ReadFile("/config/azahar-emu/qt-config.ini")
	s := string(content)

	// DefaultOnly entries preserve user's values.
	if !strings.Contains(s, "profile=2") {
		t.Errorf("default-only 'profile' should keep user value 2, got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\size=2`) {
		t.Errorf("default-only 'profiles\\size' should keep user value 2, got:\n%s", s)
	}

	// Managed region is deleted and rewritten.
	if !strings.Contains(s, `profiles\1\name=kyaraben-steamdeck`) {
		t.Errorf("managed region should have new name, got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\1\button_a=new-a`) {
		t.Errorf("managed region should have new button_a, got:\n%s", s)
	}
	if strings.Contains(s, `profiles\1\button_b`) {
		t.Errorf("old button_b should be deleted from managed region, got:\n%s", s)
	}

	// User profile is untouched.
	if !strings.Contains(s, `profiles\2\name=user-custom`) {
		t.Errorf("user profile 2 should be preserved, got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\2\button_a=user-a`) {
		t.Errorf("user profile 2 button_a should be preserved, got:\n%s", s)
	}
}

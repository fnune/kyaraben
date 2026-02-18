package configformat

import (
	"strings"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestINI_SectionRegionDeletesMatchingKeys(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": `[Controls]
profile = 2
profiles\size = 2
profiles\1\name = old-profile
profiles\1\button_a = old-value-a
profiles\1\button_b = old-value-b
profiles\2\name = user-profile
profiles\2\button_a = user-value-a
`,
	})

	handler := NewHandler(fs, model.ConfigFormatINI)

	entries := []model.ConfigEntry{
		{Path: []string{"Controls", `profiles\1\name`}, Value: "kyaraben-steamdeck"},
		{Path: []string{"Controls", `profiles\1\button_a`}, Value: "new-a"},
	}
	regions := []model.ManagedRegion{
		model.SectionRegion{Section: "Controls", KeyPrefix: `profiles\1\`},
	}

	_, err := handler.Apply("/config/test.ini", entries, regions)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	content, _ := fs.ReadFile("/config/test.ini")
	s := string(content)

	if !strings.Contains(s, `profiles\1\name = kyaraben-steamdeck`) {
		t.Errorf("should contain new profile name, got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\1\button_a = new-a`) {
		t.Errorf("should contain new button_a, got:\n%s", s)
	}
	if strings.Contains(s, "old-value") {
		t.Errorf("should not contain old values from managed region, got:\n%s", s)
	}
	if strings.Contains(s, `profiles\1\button_b`) {
		t.Errorf("old button_b should be deleted (was in managed region), got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\2\name = user-profile`) {
		t.Errorf("should preserve user profile outside managed region, got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\2\button_a = user-value-a`) {
		t.Errorf("should preserve user keys outside managed region, got:\n%s", s)
	}
}

func TestINI_SectionRegionPreservesDefaultOnlyOutsideRegion(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": `[Controls]
profile = 2
profiles\size = 2
profiles\1\name = old-profile
profiles\1\button_a = old-a
`,
	})

	handler := NewHandler(fs, model.ConfigFormatINI)

	entries := []model.ConfigEntry{
		// DefaultOnly entries outside the managed region.
		{Path: []string{"Controls", "profile"}, Value: "1", DefaultOnly: true},
		{Path: []string{"Controls", `profiles\size`}, Value: "1", DefaultOnly: true},
		// Managed entries inside the managed region.
		{Path: []string{"Controls", `profiles\1\name`}, Value: "kyaraben-steamdeck"},
		{Path: []string{"Controls", `profiles\1\button_a`}, Value: "new-a"},
	}
	regions := []model.ManagedRegion{
		model.SectionRegion{Section: "Controls", KeyPrefix: `profiles\1\`},
	}

	_, err := handler.Apply("/config/test.ini", entries, regions)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	content, _ := fs.ReadFile("/config/test.ini")
	s := string(content)

	// User changed profile to 2 before apply. Since it's default-only and existed,
	// it should be preserved.
	if !strings.Contains(s, "profile = 2") {
		t.Errorf("default-only 'profile' should be preserved at user value, got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\size = 2`) {
		t.Errorf("default-only 'profiles\\size' should be preserved at user value, got:\n%s", s)
	}
	// Managed region keys should be rewritten.
	if !strings.Contains(s, `profiles\1\name = kyaraben-steamdeck`) {
		t.Errorf("managed profile name should be written, got:\n%s", s)
	}
}

func TestINI_SectionRegionEmptyPrefixDeletesEntireSection(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": `[ControlMapping]
Up = 10-19
Down = 10-20
Left = 10-21
`,
	})

	handler := NewHandler(fs, model.ConfigFormatINI)

	entries := []model.ConfigEntry{
		{Path: []string{"ControlMapping", "Up"}, Value: "10-19,1-38"},
		{Path: []string{"ControlMapping", "Down"}, Value: "10-20,1-40"},
	}
	regions := []model.ManagedRegion{
		model.SectionRegion{Section: "ControlMapping"},
	}

	_, err := handler.Apply("/config/test.ini", entries, regions)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	content, _ := fs.ReadFile("/config/test.ini")
	s := string(content)

	if !strings.Contains(s, "Up = 10-19,1-38") {
		t.Errorf("should contain new Up value, got:\n%s", s)
	}
	if !strings.Contains(s, "Down = 10-20,1-40") {
		t.Errorf("should contain new Down value, got:\n%s", s)
	}
	if strings.Contains(s, "Left") {
		t.Errorf("Left should be deleted (entire section managed), got:\n%s", s)
	}
}

func TestINI_SectionRegionDefaultOnlyOnFreshFile(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	handler := NewHandler(fs, model.ConfigFormatINI)

	entries := []model.ConfigEntry{
		{Path: []string{"Controls", "profile"}, Value: "1", DefaultOnly: true},
		{Path: []string{"Controls", `profiles\size`}, Value: "1", DefaultOnly: true},
		{Path: []string{"Controls", `profiles\1\name`}, Value: "kyaraben-steamdeck"},
		{Path: []string{"Controls", `profiles\1\button_a`}, Value: "new-a"},
	}
	regions := []model.ManagedRegion{
		model.SectionRegion{Section: "Controls", KeyPrefix: `profiles\1\`},
	}

	_, err := handler.Apply("/config/test.ini", entries, regions)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	content, _ := fs.ReadFile("/config/test.ini")
	s := string(content)

	// On fresh file, default-only entries should be written (nothing to preserve).
	if !strings.Contains(s, "profile = 1") {
		t.Errorf("default-only 'profile' should be written on fresh file, got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\size = 1`) {
		t.Errorf("default-only 'profiles\\size' should be written on fresh file, got:\n%s", s)
	}
	if !strings.Contains(s, `profiles\1\name = kyaraben-steamdeck`) {
		t.Errorf("managed profile name should be written, got:\n%s", s)
	}
}

func TestCFG_SectionRegionDeletesMatchingKeys(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.cfg": `input_enable_hotkey_btn = "4"
input_save_state_btn = "10"
input_save_state = "f2"
menu_driver = "ozone"
`,
	})

	handler := NewHandler(fs, model.ConfigFormatCFG)

	entries := []model.ConfigEntry{
		{Path: []string{"input_enable_hotkey_btn"}, Value: "5"},
		{Path: []string{"input_save_state_btn"}, Value: "11"},
	}
	regions := []model.ManagedRegion{
		model.SectionRegion{KeyPrefix: "input_enable_hotkey_btn"},
		model.SectionRegion{KeyPrefix: "input_save_state_btn"},
	}

	_, err := handler.Apply("/config/test.cfg", entries, regions)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	content, _ := fs.ReadFile("/config/test.cfg")
	s := string(content)

	if !strings.Contains(s, `input_enable_hotkey_btn = "5"`) {
		t.Errorf("should contain new hotkey value, got:\n%s", s)
	}
	if !strings.Contains(s, `input_save_state_btn = "11"`) {
		t.Errorf("should contain new save state value, got:\n%s", s)
	}
	// Keyboard key should be preserved (not in managed region).
	if !strings.Contains(s, `input_save_state = "f2"`) {
		t.Errorf("keyboard key should be preserved, got:\n%s", s)
	}
	// Non-input key should be preserved.
	if !strings.Contains(s, `menu_driver = "ozone"`) {
		t.Errorf("menu_driver should be preserved, got:\n%s", s)
	}
}

func TestINI_NoManagedRegions(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/test.ini": "[Section]\nexisting = value\n",
	})

	handler := NewHandler(fs, model.ConfigFormatINI)

	entries := []model.ConfigEntry{
		{Path: []string{"Section", "new"}, Value: "entry"},
	}

	_, err := handler.Apply("/config/test.ini", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	content, _ := fs.ReadFile("/config/test.ini")
	s := string(content)

	if !strings.Contains(s, "existing = value") {
		t.Errorf("should preserve existing when no managed regions, got:\n%s", s)
	}
	if !strings.Contains(s, "new = entry") {
		t.Errorf("should add new entry, got:\n%s", s)
	}
}

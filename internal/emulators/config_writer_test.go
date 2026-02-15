package emulators

import (
	"strings"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/configformat"
	"github.com/fnune/kyaraben/internal/model"
)

func TestApplyYAML(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/test": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	configPath := "/test/config.yml"

	entries := []model.ConfigEntry{
		{Path: []string{"pref-path"}, Value: "/home/user/data"},
		{Path: []string{"VFS", "$(EmulatorDir)"}, Value: "/home/user/emulator"},
	}

	handler := configformat.NewHandler(fs, model.ConfigFormatYAML)
	result, err := handler.Apply(configPath, entries)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if result.Path != configPath {
		t.Errorf("Path = %q, want %q", result.Path, configPath)
	}

	content, err := fs.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "pref-path: /home/user/data") {
		t.Errorf("config should contain pref-path, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "$(EmulatorDir): /home/user/emulator") {
		t.Errorf("config should contain nested VFS.$(EmulatorDir), got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "VFS:") {
		t.Errorf("config should contain VFS section, got:\n%s", contentStr)
	}
}

func TestApplyXML(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/test": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	configPath := "/test/config.xml"

	entries := []model.ConfigEntry{
		{Path: []string{"content", "GamePaths", "Entry"}, Value: "/home/user/roms"},
		{Path: []string{"content", "mlc_path"}, Value: "/home/user/mlc"},
	}

	handler := configformat.NewHandler(fs, model.ConfigFormatXML)
	result, err := handler.Apply(configPath, entries)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if result.Path != configPath {
		t.Errorf("Path = %q, want %q", result.Path, configPath)
	}

	content, err := fs.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "<Entry>/home/user/roms</Entry>") {
		t.Errorf("config should contain Entry element, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "<mlc_path>/home/user/mlc</mlc_path>") {
		t.Errorf("config should contain mlc_path element, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "<GamePaths>") {
		t.Errorf("config should contain GamePaths element, got:\n%s", contentStr)
	}
}

func TestApplyYAMLPreservesExisting(t *testing.T) {
	existingContent := `existing-key: existing-value
other:
  nested: data
`
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/config.yml": existingContent,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	entries := []model.ConfigEntry{
		{Path: []string{"new-key"}, Value: "new-value"},
	}

	handler := configformat.NewHandler(fs, model.ConfigFormatYAML)
	_, err = handler.Apply("/config.yml", entries)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	content, err := fs.ReadFile("/config.yml")
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "existing-key: existing-value") {
		t.Errorf("should preserve existing-key, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "new-key: new-value") {
		t.Errorf("should add new-key, got:\n%s", contentStr)
	}
}

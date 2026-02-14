package emulators

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fnune/kyaraben/internal/configformat"
	"github.com/fnune/kyaraben/internal/model"
)

func TestApplyYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test", "config.yml")

	entries := []model.ConfigEntry{
		{Path: []string{"pref-path"}, Value: "/home/user/data"},
		{Path: []string{"VFS", "$(EmulatorDir)"}, Value: "/home/user/emulator"},
	}

	handler := configformat.GetHandler(model.ConfigFormatYAML)
	result, err := handler.Apply(configPath, entries)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if result.Path != configPath {
		t.Errorf("Path = %q, want %q", result.Path, configPath)
	}

	content, err := os.ReadFile(configPath)
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test", "config.xml")

	entries := []model.ConfigEntry{
		{Path: []string{"content", "GamePaths", "Entry"}, Value: "/home/user/roms"},
		{Path: []string{"content", "mlc_path"}, Value: "/home/user/mlc"},
	}

	handler := configformat.GetHandler(model.ConfigFormatXML)
	result, err := handler.Apply(configPath, entries)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if result.Path != configPath {
		t.Errorf("Path = %q, want %q", result.Path, configPath)
	}

	content, err := os.ReadFile(configPath)
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	existingContent := `existing-key: existing-value
other:
  nested: data
`
	if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("writing existing config: %v", err)
	}

	entries := []model.ConfigEntry{
		{Path: []string{"new-key"}, Value: "new-value"},
	}

	handler := configformat.GetHandler(model.ConfigFormatYAML)
	_, err := handler.Apply(configPath, entries)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
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

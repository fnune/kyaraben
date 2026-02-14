package configformat

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestINIHandler_Read(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.ini")

	content := `; Comment
[section1]
key1 = value1
key2 = value2

[section2]
key3 = value3
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	handler := &iniHandler{}
	result, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result["section1"]["key1"] != "value1" {
		t.Errorf("expected section1.key1=value1, got %s", result["section1"]["key1"])
	}
	if result["section1"]["key2"] != "value2" {
		t.Errorf("expected section1.key2=value2, got %s", result["section1"]["key2"])
	}
	if result["section2"]["key3"] != "value3" {
		t.Errorf("expected section2.key3=value3, got %s", result["section2"]["key3"])
	}
}

func TestINIHandler_Apply(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.ini")

	handler := &iniHandler{}
	entries := []model.ConfigEntry{
		{Path: []string{"section1", "key1"}, Value: "value1"},
		{Path: []string{"section1", "key2"}, Value: "value2"},
	}

	result, err := handler.Apply(path, entries)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != path {
		t.Errorf("expected path %s, got %s", path, result.Path)
	}
	if result.BaselineHash == "" {
		t.Error("expected non-empty hash")
	}

	readResult, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult["section1"]["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", readResult["section1"]["key1"])
	}
}

func TestCFGHandler_Read(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.cfg")

	content := `# Comment
key1 = "value1"
key2 = "value2"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	handler := &cfgHandler{}
	result, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result[""]["key1"] != `"value1"` {
		t.Errorf("expected key1=\"value1\", got %s", result[""]["key1"])
	}
	if result[""]["key2"] != `"value2"` {
		t.Errorf("expected key2=\"value2\", got %s", result[""]["key2"])
	}
}

func TestCFGHandler_Apply(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.cfg")

	handler := &cfgHandler{}
	entries := []model.ConfigEntry{
		{Path: []string{"key1"}, Value: "value1"},
		{Path: []string{"key2"}, Value: "value2"},
	}

	result, err := handler.Apply(path, entries)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != path {
		t.Errorf("expected path %s, got %s", path, result.Path)
	}

	readResult, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult[""]["key1"] != `"value1"` {
		t.Errorf("expected key1=\"value1\", got %s", readResult[""]["key1"])
	}
}

func TestTOMLHandler_Read(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.toml")

	content := `# Comment
[section1]
key1 = "value1"
key2 = "value2"

[section2]
key3 = "value3"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	handler := &tomlHandler{}
	result, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result["section1"]["key1"] != "value1" {
		t.Errorf("expected section1.key1=value1, got %s", result["section1"]["key1"])
	}
	if result["section2"]["key3"] != "value3" {
		t.Errorf("expected section2.key3=value3, got %s", result["section2"]["key3"])
	}
}

func TestTOMLHandler_Apply(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.toml")

	handler := &tomlHandler{}
	entries := []model.ConfigEntry{
		{Path: []string{"section1", "key1"}, Value: "value1"},
		{Path: []string{"section1", "key2"}, Value: "value2"},
	}

	result, err := handler.Apply(path, entries)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != path {
		t.Errorf("expected path %s, got %s", path, result.Path)
	}

	readResult, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult["section1"]["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", readResult["section1"]["key1"])
	}
}

func TestYAMLHandler_Read(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yaml")

	content := `# Comment
section1:
  key1: value1
  key2: value2
section2:
  key3: value3
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	handler := &yamlHandler{}
	result, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result["section1"]["key1"] != "value1" {
		t.Errorf("expected section1.key1=value1, got %s", result["section1"]["key1"])
	}
	if result["section2"]["key3"] != "value3" {
		t.Errorf("expected section2.key3=value3, got %s", result["section2"]["key3"])
	}
}

func TestYAMLHandler_Apply(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.yaml")

	handler := &yamlHandler{}
	entries := []model.ConfigEntry{
		{Path: []string{"section1", "key1"}, Value: "value1"},
		{Path: []string{"section1", "key2"}, Value: "value2"},
	}

	result, err := handler.Apply(path, entries)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != path {
		t.Errorf("expected path %s, got %s", path, result.Path)
	}

	readResult, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult["section1"]["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", readResult["section1"]["key1"])
	}
}

func TestXMLHandler_Read(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.xml")

	content := `<config>
  <section1>
    <key1>value1</key1>
    <key2>value2</key2>
  </section1>
  <section2>
    <key3>value3</key3>
  </section2>
</config>`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	handler := &xmlHandler{}
	result, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if result["config.section1"]["key1"] != "value1" {
		t.Errorf("expected config.section1.key1=value1, got %s", result["config.section1"]["key1"])
	}
	if result["config.section2"]["key3"] != "value3" {
		t.Errorf("expected config.section2.key3=value3, got %s", result["config.section2"]["key3"])
	}
}

func TestXMLHandler_Apply(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.xml")

	handler := &xmlHandler{}
	entries := []model.ConfigEntry{
		{Path: []string{"config", "section1", "key1"}, Value: "value1"},
		{Path: []string{"config", "section1", "key2"}, Value: "value2"},
	}

	result, err := handler.Apply(path, entries)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != path {
		t.Errorf("expected path %s, got %s", path, result.Path)
	}

	readResult, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult["config.section1"]["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", readResult["config.section1"]["key1"])
	}
}

func TestRawHandler_Read(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.raw")

	content := "raw content here"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing file: %v", err)
	}

	handler := &rawHandler{}
	result, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(result) != 1 || result[""] == nil {
		t.Error("expected empty result map with empty section")
	}
}

func TestRawHandler_Apply(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.raw")

	handler := &rawHandler{}
	entries := []model.ConfigEntry{
		{Path: []string{}, Value: "raw content"},
	}

	result, err := handler.Apply(path, entries)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != path {
		t.Errorf("expected path %s, got %s", path, result.Path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if string(data) != "raw content" {
		t.Errorf("expected 'raw content', got '%s'", string(data))
	}
}

func TestGetHandler(t *testing.T) {
	tests := []struct {
		format model.ConfigFormat
	}{
		{model.ConfigFormatINI},
		{model.ConfigFormatCFG},
		{model.ConfigFormatTOML},
		{model.ConfigFormatYAML},
		{model.ConfigFormatXML},
		{model.ConfigFormatRaw},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			handler := GetHandler(tt.format)
			if handler == nil {
				t.Fatal("expected non-nil handler")
			}
		})
	}
}

func TestFlattenNestedMap(t *testing.T) {
	nested := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"key": "value",
			},
		},
		"toplevel": "topvalue",
	}

	result := make(map[string]map[string]string)
	flattenNestedMap(nested, nil, result)

	if result["level1.level2"]["key"] != "value" {
		t.Errorf("expected level1.level2.key=value, got %s", result["level1.level2"]["key"])
	}
	if result[""]["toplevel"] != "topvalue" {
		t.Errorf("expected toplevel=topvalue, got %s", result[""]["toplevel"])
	}
}

func TestSetNestedValue(t *testing.T) {
	m := make(map[string]interface{})
	setNestedValue(m, []string{"a", "b", "c"}, "value")

	a, ok := m["a"].(map[string]interface{})
	if !ok {
		t.Fatal("expected map at key 'a'")
	}
	b, ok := a["b"].(map[string]interface{})
	if !ok {
		t.Fatal("expected map at key 'b'")
	}
	if b["c"] != "value" {
		t.Errorf("expected c=value, got %v", b["c"])
	}
}

func TestHasNestedValue(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "value",
		},
	}

	if !hasNestedValue(m, []string{"a", "b"}) {
		t.Error("expected to find a.b")
	}
	if hasNestedValue(m, []string{"a", "c"}) {
		t.Error("expected not to find a.c")
	}
	if hasNestedValue(m, []string{"x"}) {
		t.Error("expected not to find x")
	}
}

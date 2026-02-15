package configformat

import (
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/testutil"
)

func TestINIHandler_Read(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/test.ini": `; Comment
[section1]
key1 = value1
key2 = value2

[section2]
key3 = value3
`,
	})

	handler := NewHandler(fs, model.ConfigFormatINI)
	result, err := handler.Read("/test.ini")
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	handler := NewHandler(fs, model.ConfigFormatINI)
	entries := []model.ConfigEntry{
		{Path: []string{"section1", "key1"}, Value: "value1"},
		{Path: []string{"section1", "key2"}, Value: "value2"},
	}

	result, err := handler.Apply("/config/test.ini", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != "/config/test.ini" {
		t.Errorf("expected path /config/test.ini, got %s", result.Path)
	}
	if result.BaselineHash == "" {
		t.Error("expected non-empty hash")
	}

	readResult, err := handler.Read("/config/test.ini")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult["section1"]["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", readResult["section1"]["key1"])
	}
}

func TestCFGHandler_Read(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/test.cfg": `# Comment
key1 = "value1"
key2 = "value2"
`,
	})

	handler := NewHandler(fs, model.ConfigFormatCFG)
	result, err := handler.Read("/test.cfg")
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	handler := NewHandler(fs, model.ConfigFormatCFG)
	entries := []model.ConfigEntry{
		{Path: []string{"key1"}, Value: "value1"},
		{Path: []string{"key2"}, Value: "value2"},
	}

	result, err := handler.Apply("/config/test.cfg", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != "/config/test.cfg" {
		t.Errorf("expected path /config/test.cfg, got %s", result.Path)
	}

	readResult, err := handler.Read("/config/test.cfg")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult[""]["key1"] != `"value1"` {
		t.Errorf("expected key1=\"value1\", got %s", readResult[""]["key1"])
	}
}

func TestTOMLHandler_Read(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/test.toml": `# Comment
[section1]
key1 = "value1"
key2 = "value2"

[section2]
key3 = "value3"
`,
	})

	handler := NewHandler(fs, model.ConfigFormatTOML)
	result, err := handler.Read("/test.toml")
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	handler := NewHandler(fs, model.ConfigFormatTOML)
	entries := []model.ConfigEntry{
		{Path: []string{"section1", "key1"}, Value: "value1"},
		{Path: []string{"section1", "key2"}, Value: "value2"},
	}

	result, err := handler.Apply("/config/test.toml", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != "/config/test.toml" {
		t.Errorf("expected path /config/test.toml, got %s", result.Path)
	}

	readResult, err := handler.Read("/config/test.toml")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult["section1"]["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", readResult["section1"]["key1"])
	}
}

func TestYAMLHandler_Read(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/test.yaml": `# Comment
section1:
  key1: value1
  key2: value2
section2:
  key3: value3
`,
	})

	handler := NewHandler(fs, model.ConfigFormatYAML)
	result, err := handler.Read("/test.yaml")
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	handler := NewHandler(fs, model.ConfigFormatYAML)
	entries := []model.ConfigEntry{
		{Path: []string{"section1", "key1"}, Value: "value1"},
		{Path: []string{"section1", "key2"}, Value: "value2"},
	}

	result, err := handler.Apply("/config/test.yaml", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != "/config/test.yaml" {
		t.Errorf("expected path /config/test.yaml, got %s", result.Path)
	}

	readResult, err := handler.Read("/config/test.yaml")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult["section1"]["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", readResult["section1"]["key1"])
	}
}

func TestXMLHandler_Read(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/test.xml": `<config>
  <section1>
    <key1>value1</key1>
    <key2>value2</key2>
  </section1>
  <section2>
    <key3>value3</key3>
  </section2>
</config>`,
	})

	handler := NewHandler(fs, model.ConfigFormatXML)
	result, err := handler.Read("/test.xml")
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	handler := NewHandler(fs, model.ConfigFormatXML)
	entries := []model.ConfigEntry{
		{Path: []string{"config", "section1", "key1"}, Value: "value1"},
		{Path: []string{"config", "section1", "key2"}, Value: "value2"},
	}

	result, err := handler.Apply("/config/test.xml", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != "/config/test.xml" {
		t.Errorf("expected path /config/test.xml, got %s", result.Path)
	}

	readResult, err := handler.Read("/config/test.xml")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if readResult["config.section1"]["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", readResult["config.section1"]["key1"])
	}
}

func TestRawHandler_Read(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/test.raw": "raw content here",
	})

	handler := NewHandler(fs, model.ConfigFormatRaw)
	result, err := handler.Read("/test.raw")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(result) != 1 || result[""] == nil {
		t.Error("expected empty result map with empty section")
	}
}

func TestRawHandler_Apply(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	handler := NewHandler(fs, model.ConfigFormatRaw)
	entries := []model.ConfigEntry{
		{Path: []string{}, Value: "raw content"},
	}

	result, err := handler.Apply("/config/test.raw", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if result.Path != "/config/test.raw" {
		t.Errorf("expected path /config/test.raw, got %s", result.Path)
	}

	data, err := fs.ReadFile("/config/test.raw")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if string(data) != "raw content" {
		t.Errorf("expected 'raw content', got '%s'", string(data))
	}
}

func TestRawHandler_ApplyUnmanaged(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config/existing.raw": "user content",
	})

	handler := NewHandler(fs, model.ConfigFormatRaw)
	entries := []model.ConfigEntry{
		{Value: "new content", Unmanaged: true},
	}

	_, err := handler.Apply("/config/existing.raw", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := fs.ReadFile("/config/existing.raw")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if string(data) != "user content" {
		t.Errorf("expected existing content preserved, got '%s'", string(data))
	}
}

func TestRawHandler_ApplyUnmanagedCreatesIfMissing(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/config": &vfst.Dir{Perm: 0755},
	})

	handler := NewHandler(fs, model.ConfigFormatRaw)
	entries := []model.ConfigEntry{
		{Value: "new content", Unmanaged: true},
	}

	_, err := handler.Apply("/config/new.raw", entries, nil)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := fs.ReadFile("/config/new.raw")
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if string(data) != "new content" {
		t.Errorf("expected new content, got '%s'", string(data))
	}
}

func TestGetHandler(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

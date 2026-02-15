package configformat

import (
	"fmt"
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"
	"gopkg.in/yaml.v3"

	"github.com/fnune/kyaraben/internal/model"
)

type yamlHandler struct {
	fs vfs.FS
}

func (h *yamlHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)

	data, err := h.fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var nested map[string]interface{}
	if err := yaml.Unmarshal(data, &nested); err != nil {
		return nil, err
	}

	flattenNestedMap(nested, nil, result)
	return result, nil
}

func (h *yamlHandler) Apply(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := vfs.MkdirAll(h.fs, filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	existing := make(map[string]interface{})
	if data, err := h.fs.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &existing); err != nil {
			return ApplyResult{}, fmt.Errorf("parsing existing YAML: %w", err)
		}
	}

	for _, entry := range entries {
		if entry.Unmanaged && hasNestedValue(existing, entry.Path) {
			continue
		}
		setNestedValue(existing, entry.Path, entry.Value)
	}

	f, err := h.fs.Create(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("creating config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintln(f, "# Configuration managed by kyaraben")
	_, _ = fmt.Fprintln(f, "# Manual changes will be preserved on next apply")
	_, _ = fmt.Fprintln(f)

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(existing); err != nil {
		return ApplyResult{}, fmt.Errorf("encoding YAML: %w", err)
	}
	_ = encoder.Close()

	hash, err := hashFileWithFS(h.fs, path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

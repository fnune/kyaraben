package configformat

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/fnune/kyaraben/internal/model"
)

type tomlHandler struct{}

func (h *tomlHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var nested map[string]interface{}
	if _, err := toml.Decode(string(data), &nested); err != nil {
		return nil, err
	}

	flattenNestedMap(nested, nil, result)
	return result, nil
}

func (h *tomlHandler) Apply(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		if _, err := toml.Decode(string(data), &existing); err != nil {
			return ApplyResult{}, fmt.Errorf("parsing existing TOML: %w", err)
		}
	}

	for _, entry := range entries {
		if entry.Unmanaged && hasNestedValue(existing, entry.Path) {
			continue
		}
		setNestedValue(existing, entry.Path, entry.Value)
	}

	f, err := os.Create(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("creating config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintln(f, "# Configuration managed by kyaraben")
	_, _ = fmt.Fprintln(f, "# Manual changes will be preserved on next apply")
	_, _ = fmt.Fprintln(f)

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(existing); err != nil {
		return ApplyResult{}, fmt.Errorf("encoding TOML: %w", err)
	}

	hash, err := hashFile(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

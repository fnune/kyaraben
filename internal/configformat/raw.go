package configformat

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type rawHandler struct{}

func (h *rawHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)
	return result, nil
}

func (h *rawHandler) Apply(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	if len(entries) != 1 {
		return ApplyResult{}, fmt.Errorf("raw format requires exactly one entry with full content")
	}

	content := entries[0].Value
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ApplyResult{}, fmt.Errorf("writing raw file: %w", err)
	}

	hash, err := hashFile(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

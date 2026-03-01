package configformat

import (
	"fmt"
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
)

type rawHandler struct {
	fs vfs.FS
}

func (h *rawHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)
	return result, nil
}

func (h *rawHandler) Apply(path string, entries []model.ConfigEntry, _ []model.ManagedRegion) (ApplyResult, error) {
	if len(entries) != 1 {
		return ApplyResult{}, fmt.Errorf("raw format requires exactly one entry with full content")
	}

	entry := entries[0]
	writtenEntries := make(map[string]string)

	if entry.DefaultOnly {
		if _, err := h.fs.Stat(path); err == nil {
			return ApplyResult{Path: path, WrittenEntries: writtenEntries}, nil
		}
	}

	if err := vfs.MkdirAll(h.fs, filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	if err := h.fs.WriteFile(path, []byte(entry.Value), 0644); err != nil {
		return ApplyResult{}, fmt.Errorf("writing raw file: %w", err)
	}

	writtenEntries[entry.FullPath()] = entry.Value
	return ApplyResult{Path: path, WrittenEntries: writtenEntries}, nil
}

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

	if entry.DefaultOnly {
		if _, err := h.fs.Stat(path); err == nil {
			hash, err := hashFileWithFS(h.fs, path)
			if err != nil {
				return ApplyResult{}, fmt.Errorf("hashing existing file: %w", err)
			}
			return ApplyResult{Path: path, BaselineHash: hash, PatchHash: ComputePatchHash(entries)}, nil
		}
	}

	if err := vfs.MkdirAll(h.fs, filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	if err := h.fs.WriteFile(path, []byte(entry.Value), 0644); err != nil {
		return ApplyResult{}, fmt.Errorf("writing raw file: %w", err)
	}

	hash, err := hashFileWithFS(h.fs, path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash, PatchHash: ComputePatchHash(entries)}, nil
}

package configformat

import (
	"bufio"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
)

type cfgHandler struct {
	fs vfs.FS
}

func (h *cfgHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)

	f, err := h.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		result[""][key] = value
	}

	return result, scanner.Err()
}

func (h *cfgHandler) Apply(path string, entries []model.ConfigEntry, managedRegions []model.ManagedRegion) (ApplyResult, error) {
	if err := vfs.MkdirAll(h.fs, filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	existing := make(map[string]string)
	if data, err := h.fs.Open(path); err == nil {
		scanner := bufio.NewScanner(data)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if idx := strings.Index(line, "="); idx != -1 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])
				existing[key] = value
			}
		}
		_ = data.Close()
	}

	snapshot := make(map[string]string, len(existing))
	for k, v := range existing {
		snapshot[k] = v
	}

	for _, region := range managedRegions {
		sr, ok := region.(model.SectionRegion)
		if !ok {
			continue
		}
		if sr.Section != "" {
			continue
		}
		if sr.KeyPrefix == "" {
			for k := range existing {
				delete(existing, k)
			}
			continue
		}
		for k := range existing {
			if strings.HasPrefix(k, sr.KeyPrefix) {
				delete(existing, k)
			}
		}
	}

	for _, entry := range entries {
		key := entry.Key()
		if entry.DefaultOnly && snapshot[key] != "" {
			continue
		}
		existing[key] = `"` + entry.Value + `"`
	}

	f, err := h.fs.Create(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("creating config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintln(f, "# Configuration managed by kyaraben")
	if isFullyManaged(managedRegions) {
		_, _ = fmt.Fprintln(f, "# Manual changes will be overwritten on next apply")
	} else {
		_, _ = fmt.Fprintln(f, "# Manual changes will be preserved on next apply")
	}
	_, _ = fmt.Fprintln(f)

	keys := make([]string, 0, len(existing))
	for key := range existing {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		_, _ = fmt.Fprintf(f, "%s = %s\n", key, existing[key])
	}

	hash, err := hashFileWithFS(h.fs, path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

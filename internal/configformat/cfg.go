package configformat

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type cfgHandler struct{}

func (h *cfgHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)

	f, err := os.Open(path)
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

func (h *cfgHandler) Apply(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	existing := make(map[string]string)
	if data, err := os.Open(path); err == nil {
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

	for _, entry := range entries {
		key := entry.Key()
		if entry.Unmanaged && existing[key] != "" {
			continue
		}
		existing[key] = `"` + entry.Value + `"`
	}

	f, err := os.Create(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("creating config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintln(f, "# Configuration managed by kyaraben")
	_, _ = fmt.Fprintln(f, "# Manual changes will be preserved on next apply")
	_, _ = fmt.Fprintln(f)

	keys := make([]string, 0, len(existing))
	for key := range existing {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		_, _ = fmt.Fprintf(f, "%s = %s\n", key, existing[key])
	}

	hash, err := hashFile(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

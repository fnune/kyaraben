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

type iniHandler struct {
	fs vfs.FS
}

func (h *iniHandler) Read(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string)

	f, err := h.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			if result[currentSection] == nil {
				result[currentSection] = make(map[string]string)
			}
			continue
		}

		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		if result[currentSection] == nil {
			result[currentSection] = make(map[string]string)
		}
		result[currentSection][key] = value
	}

	return result, scanner.Err()
}

func (h *iniHandler) Apply(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := vfs.MkdirAll(h.fs, filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	sections := make(map[string]map[string]string)
	currentSection := ""

	if data, err := h.fs.Open(path); err == nil {
		scanner := bufio.NewScanner(data)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				currentSection = line[1 : len(line)-1]
				if sections[currentSection] == nil {
					sections[currentSection] = make(map[string]string)
				}
				continue
			}
			if idx := strings.Index(line, "="); idx != -1 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])
				if sections[currentSection] == nil {
					sections[currentSection] = make(map[string]string)
				}
				sections[currentSection][key] = value
			}
		}
		_ = data.Close()
	}

	for _, entry := range entries {
		section := SectionKey(entry.Parent())
		if sections[section] == nil {
			sections[section] = make(map[string]string)
		}
		key := entry.Key()
		if entry.Unmanaged && sections[section][key] != "" {
			continue
		}
		sections[section][key] = entry.Value
	}

	f, err := h.fs.Create(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("creating config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintln(f, "; Configuration managed by kyaraben")
	_, _ = fmt.Fprintln(f, "; Manual changes will be preserved on next apply")
	_, _ = fmt.Fprintln(f)

	sectionNames := make([]string, 0, len(sections))
	for section := range sections {
		sectionNames = append(sectionNames, section)
	}
	sort.Strings(sectionNames)

	for _, section := range sectionNames {
		values := sections[section]
		if section != "" {
			_, _ = fmt.Fprintf(f, "[%s]\n", section)
		}

		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			_, _ = fmt.Fprintf(f, "%s = %s\n", key, values[key])
		}
		_, _ = fmt.Fprintln(f)
	}

	hash, err := hashFileWithFS(h.fs, path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

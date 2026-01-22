package emulators

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type ConfigWriter struct{}

func NewConfigWriter() *ConfigWriter {
	return &ConfigWriter{}
}

// Existing values are preserved unless kyaraben sets them.
func (w *ConfigWriter) Apply(patch model.ConfigPatch) error {
	switch patch.Config.Format {
	case model.ConfigFormatCFG:
		return w.applyCFG(patch)
	case model.ConfigFormatINI:
		return w.applyINI(patch)
	default:
		return fmt.Errorf("unsupported config format: %s", patch.Config.Format)
	}
}

func (w *ConfigWriter) applyCFG(patch model.ConfigPatch) error {
	path := patch.Config.Path

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Read existing config if it exists
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

	// Apply patches
	for _, entry := range patch.Entries {
		existing[entry.Key] = entry.Value
	}

	// Write config
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
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

	return nil
}

func (w *ConfigWriter) applyINI(patch model.ConfigPatch) error {
	path := patch.Config.Path

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	sections := make(map[string]map[string]string)
	currentSection := ""

	if data, err := os.Open(path); err == nil {
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

	for _, entry := range patch.Entries {
		if sections[entry.Section] == nil {
			sections[entry.Section] = make(map[string]string)
		}
		sections[entry.Section][entry.Key] = entry.Value
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
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

	return nil
}

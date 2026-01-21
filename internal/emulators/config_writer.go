package emulators

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

// ConfigWriter applies config patches to emulator config files.
type ConfigWriter struct{}

// NewConfigWriter creates a new config writer.
func NewConfigWriter() *ConfigWriter {
	return &ConfigWriter{}
}

// Apply writes a config patch to its target file.
// For MVP, this does a simple merge: existing values are preserved unless
// kyaraben sets them.
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

// applyCFG handles RetroArch-style key = "value" configs.
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
		data.Close()
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
	defer f.Close()

	// Write kyaraben header
	fmt.Fprintln(f, "# Configuration managed by kyaraben")
	fmt.Fprintln(f, "# Manual changes will be preserved on next apply")
	fmt.Fprintln(f)

	for key, value := range existing {
		fmt.Fprintf(f, "%s = %s\n", key, value)
	}

	return nil
}

// applyINI handles INI-style configs with [sections].
func (w *ConfigWriter) applyINI(patch model.ConfigPatch) error {
	path := patch.Config.Path

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Read existing config if it exists
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
		data.Close()
	}

	// Apply patches
	for _, entry := range patch.Entries {
		if sections[entry.Section] == nil {
			sections[entry.Section] = make(map[string]string)
		}
		sections[entry.Section][entry.Key] = entry.Value
	}

	// Write config
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	// Write kyaraben header
	fmt.Fprintln(f, "; Configuration managed by kyaraben")
	fmt.Fprintln(f, "; Manual changes will be preserved on next apply")
	fmt.Fprintln(f)

	// Write sections
	for section, values := range sections {
		if section != "" {
			fmt.Fprintf(f, "[%s]\n", section)
		}
		for key, value := range values {
			fmt.Fprintf(f, "%s = %s\n", key, value)
		}
		fmt.Fprintln(f)
	}

	return nil
}

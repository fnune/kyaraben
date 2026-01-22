package emulators

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

// ConfigDiff represents the difference between current and proposed config.
type ConfigDiff struct {
	Path    string
	Changes []ConfigChange
}

// ConfigChange represents a single change in a config file.
type ConfigChange struct {
	Type     ChangeType
	Section  string // For INI files
	Key      string
	OldValue string
	NewValue string
}

// ChangeType indicates what kind of change this is.
type ChangeType int

const (
	ChangeAdd ChangeType = iota
	ChangeModify
	ChangeRemove
)

// ComputeDiff computes the difference between current config and proposed patch.
func ComputeDiff(patch model.ConfigPatch) (*ConfigDiff, error) {
	diff := &ConfigDiff{
		Path:    patch.Config.Path,
		Changes: make([]ConfigChange, 0),
	}

	// Read current config
	current := make(map[string]map[string]string) // section -> key -> value

	if _, err := os.Stat(patch.Config.Path); err == nil {
		var err error
		current, err = readConfig(patch.Config.Path, patch.Config.Format)
		if err != nil {
			return nil, fmt.Errorf("reading current config: %w", err)
		}
	}

	// Compare with proposed changes
	for _, entry := range patch.Entries {
		section := entry.Section
		key := entry.Key
		newValue := entry.Value

		sectionMap, sectionExists := current[section]
		if !sectionExists {
			// New section and key
			diff.Changes = append(diff.Changes, ConfigChange{
				Type:     ChangeAdd,
				Section:  section,
				Key:      key,
				NewValue: newValue,
			})
			continue
		}

		oldValue, keyExists := sectionMap[key]
		if !keyExists {
			// New key in existing section
			diff.Changes = append(diff.Changes, ConfigChange{
				Type:     ChangeAdd,
				Section:  section,
				Key:      key,
				NewValue: newValue,
			})
			continue
		}

		// Key exists - check if value changed
		if oldValue != newValue {
			diff.Changes = append(diff.Changes, ConfigChange{
				Type:     ChangeModify,
				Section:  section,
				Key:      key,
				OldValue: oldValue,
				NewValue: newValue,
			})
		}
	}

	return diff, nil
}

// HasChanges returns true if there are any changes.
func (d *ConfigDiff) HasChanges() bool {
	return len(d.Changes) > 0
}

// Format returns a human-readable string representation of the diff.
func (d *ConfigDiff) Format() string {
	if !d.HasChanges() {
		return fmt.Sprintf("  %s: no changes", d.Path)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  %s\n", d.Path))

	for _, change := range d.Changes {
		prefix := ""
		switch change.Type {
		case ChangeAdd:
			prefix = "+"
		case ChangeModify:
			prefix = "~"
		case ChangeRemove:
			prefix = "-"
		}

		key := change.Key
		if change.Section != "" {
			key = fmt.Sprintf("[%s] %s", change.Section, change.Key)
		}

		switch change.Type {
		case ChangeAdd:
			sb.WriteString(fmt.Sprintf("    %s %s = %s\n", prefix, key, change.NewValue))
		case ChangeModify:
			sb.WriteString(fmt.Sprintf("    %s %s: %s -> %s\n", prefix, key, change.OldValue, change.NewValue))
		case ChangeRemove:
			sb.WriteString(fmt.Sprintf("    %s %s = %s\n", prefix, key, change.OldValue))
		}
	}

	return sb.String()
}

func readConfig(path string, format model.ConfigFormat) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	result[""] = make(map[string]string) // Default section for flat configs

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section header (INI format)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			if result[currentSection] == nil {
				result[currentSection] = make(map[string]string)
			}
			continue
		}

		// Parse key=value
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

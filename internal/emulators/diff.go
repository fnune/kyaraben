package emulators

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

type ConfigDiff struct {
	Path      string
	IsNewFile bool
	Changes   []ConfigChange
}

type ConfigChange struct {
	Type     ChangeType
	Section  string
	Key      string
	OldValue string
	NewValue string
}

type ChangeType int

const (
	ChangeAdd ChangeType = iota
	ChangeModify
	ChangeRemove
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

func ComputeDiff(patch model.ConfigPatch) (*ConfigDiff, error) {
	diff := &ConfigDiff{
		Path:      patch.Config.Path,
		IsNewFile: true,
		Changes:   make([]ConfigChange, 0),
	}

	// Read current config
	current := make(map[string]map[string]string) // section -> key -> value

	if _, err := os.Stat(patch.Config.Path); err == nil {
		diff.IsNewFile = false
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

func (d *ConfigDiff) HasChanges() bool {
	return len(d.Changes) > 0
}

func (d *ConfigDiff) Stats() (adds, modifies, removes int) {
	for _, c := range d.Changes {
		switch c.Type {
		case ChangeAdd:
			adds++
		case ChangeModify:
			modifies++
		case ChangeRemove:
			removes++
		}
	}
	return
}

func (d *ConfigDiff) Format() string {
	return d.FormatWithColor(true)
}

func (d *ConfigDiff) FormatWithColor(useColor bool) string {
	// Color helpers
	green := func(s string) string {
		if useColor {
			return colorGreen + s + colorReset
		}
		return s
	}
	yellow := func(s string) string {
		if useColor {
			return colorYellow + s + colorReset
		}
		return s
	}
	red := func(s string) string {
		if useColor {
			return colorRed + s + colorReset
		}
		return s
	}
	cyan := func(s string) string {
		if useColor {
			return colorCyan + s + colorReset
		}
		return s
	}
	bold := func(s string) string {
		if useColor {
			return colorBold + s + colorReset
		}
		return s
	}
	dim := func(s string) string {
		if useColor {
			return colorDim + s + colorReset
		}
		return s
	}

	var sb strings.Builder

	// File header with status
	if d.IsNewFile {
		sb.WriteString(fmt.Sprintf("  %s %s\n", green("CREATE"), bold(d.Path)))
	} else if !d.HasChanges() {
		sb.WriteString(fmt.Sprintf("  %s %s\n", dim("UNCHANGED"), d.Path))
		return sb.String()
	} else {
		sb.WriteString(fmt.Sprintf("  %s %s\n", yellow("MODIFY"), bold(d.Path)))
	}

	// Group changes by section for better readability
	sectionChanges := make(map[string][]ConfigChange)
	sectionOrder := make([]string, 0)
	for _, change := range d.Changes {
		if _, exists := sectionChanges[change.Section]; !exists {
			sectionOrder = append(sectionOrder, change.Section)
		}
		sectionChanges[change.Section] = append(sectionChanges[change.Section], change)
	}

	for _, section := range sectionOrder {
		changes := sectionChanges[section]

		// Section header (for INI files)
		if section != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", cyan(fmt.Sprintf("[%s]", section))))
		}

		for _, change := range changes {
			indent := "    "
			if section != "" {
				indent = "      "
			}

			switch change.Type {
			case ChangeAdd:
				sb.WriteString(fmt.Sprintf("%s%s %s = %s\n",
					indent, green("+"), change.Key, green(change.NewValue)))
			case ChangeModify:
				sb.WriteString(fmt.Sprintf("%s%s %s\n",
					indent, yellow("~"), change.Key))
				sb.WriteString(fmt.Sprintf("%s    %s %s\n",
					indent, red("-"), dim(change.OldValue)))
				sb.WriteString(fmt.Sprintf("%s    %s %s\n",
					indent, green("+"), change.NewValue))
			case ChangeRemove:
				sb.WriteString(fmt.Sprintf("%s%s %s = %s\n",
					indent, red("-"), change.Key, red(change.OldValue)))
			}
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

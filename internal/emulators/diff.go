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
	path, err := patch.Target.Resolve()
	if err != nil {
		return nil, fmt.Errorf("resolving config path: %w", err)
	}

	diff := &ConfigDiff{
		Path:      path,
		IsNewFile: true,
		Changes:   make([]ConfigChange, 0),
	}

	current := make(map[string]map[string]string)

	if _, err := os.Stat(path); err == nil {
		diff.IsNewFile = false
		current, err = readConfig(path, patch.Target.Format)
		if err != nil {
			return nil, fmt.Errorf("reading current config: %w", err)
		}
	}

	for _, entry := range patch.Entries {
		section := entry.Section
		key := entry.Key
		newValue := entry.Value

		sectionMap, sectionExists := current[section]
		if !sectionExists {
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
			diff.Changes = append(diff.Changes, ConfigChange{
				Type:     ChangeAdd,
				Section:  section,
				Key:      key,
				NewValue: newValue,
			})
			continue
		}

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

	if d.IsNewFile {
		sb.WriteString(fmt.Sprintf("  %s %s\n", green("CREATE"), bold(d.Path)))
	} else if !d.HasChanges() {
		sb.WriteString(fmt.Sprintf("  %s %s\n", dim("UNCHANGED"), d.Path))
		return sb.String()
	} else {
		sb.WriteString(fmt.Sprintf("  %s %s\n", yellow("MODIFY"), bold(d.Path)))
	}

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
	result[""] = make(map[string]string)

	f, err := os.Open(path)
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

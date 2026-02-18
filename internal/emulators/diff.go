package emulators

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/configformat"
	"github.com/fnune/kyaraben/internal/model"
)

type DiffComputer struct {
	fs       vfs.FS
	resolver model.BaseDirResolver
}

func NewDiffComputer(fs vfs.FS, resolver model.BaseDirResolver) *DiffComputer {
	return &DiffComputer{fs: fs, resolver: resolver}
}

func NewDefaultDiffComputer() *DiffComputer {
	return NewDiffComputer(vfs.OSFS, &model.OSBaseDirResolver{})
}

type ConfigDiff struct {
	Path         string
	IsNewFile    bool
	UserModified bool
	Changes      []ConfigChange
	UserChanges  []UserChange
}

type UserChange struct {
	Path          []string
	BaselineValue string
	CurrentValue  string
}

type ConfigChange struct {
	Type     ChangeType
	Path     []string
	OldValue string
	NewValue string
}

func (c ConfigChange) Key() string {
	if len(c.Path) == 0 {
		return ""
	}
	return c.Path[len(c.Path)-1]
}

func (c ConfigChange) Parent() []string {
	if len(c.Path) <= 1 {
		return nil
	}
	return c.Path[:len(c.Path)-1]
}

func (c ConfigChange) Section() string {
	parent := c.Parent()
	if len(parent) == 0 {
		return ""
	}
	return strings.Join(parent, ".")
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

func ComputeDiffWithBaseline(patch model.ConfigPatch, baseline *model.ManagedConfig) (*ConfigDiff, error) {
	return NewDefaultDiffComputer().ComputeDiffWithBaseline(patch, baseline)
}

func (d *DiffComputer) ComputeDiffWithBaseline(patch model.ConfigPatch, baseline *model.ManagedConfig) (*ConfigDiff, error) {
	diff, err := d.ComputeDiff(patch)
	if err != nil {
		return nil, err
	}

	if baseline == nil || diff.IsNewFile {
		return diff, nil
	}

	currentHash, err := d.hashConfigFile(diff.Path)
	if err != nil {
		return diff, nil
	}

	if currentHash != baseline.BaselineHash {
		diff.UserModified = true

		current, err := d.readConfig(diff.Path, patch.Target.Format)
		if err != nil {
			return diff, nil
		}

		for _, entry := range patch.Entries {
			if entry.DefaultOnly {
				continue
			}
			section := configformat.SectionKey(entry.Parent())
			key := entry.Key()

			if sectionMap, ok := current[section]; ok {
				if currentVal, ok := sectionMap[key]; ok && configformat.NormalizePath(currentVal, d.homeDir()) != configformat.NormalizePath(entry.Value, d.homeDir()) {
					diff.UserChanges = append(diff.UserChanges, UserChange{
						Path:          entry.Path,
						BaselineValue: entry.Value,
						CurrentValue:  currentVal,
					})
				}
			}
		}
	}

	return diff, nil
}

func (d *DiffComputer) hashConfigFile(path string) (string, error) {
	data, err := d.fs.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (d *DiffComputer) homeDir() string {
	home, _ := d.resolver.UserHomeDir()
	return home
}

func ComputeDiff(patch model.ConfigPatch) (*ConfigDiff, error) {
	return NewDefaultDiffComputer().ComputeDiff(patch)
}

func (d *DiffComputer) ComputeDiff(patch model.ConfigPatch) (*ConfigDiff, error) {
	path, err := patch.Target.ResolveWith(d.resolver)
	if err != nil {
		return nil, fmt.Errorf("resolving config path: %w", err)
	}

	diff := &ConfigDiff{
		Path:      path,
		IsNewFile: true,
		Changes:   make([]ConfigChange, 0),
	}

	current := make(map[string]map[string]string)

	if _, err := d.fs.Stat(path); err == nil {
		diff.IsNewFile = false
		current, err = d.readConfig(path, patch.Target.Format)
		if err != nil {
			return nil, fmt.Errorf("reading current config: %w", err)
		}
	}

	for _, entry := range patch.Entries {
		section := configformat.SectionKey(entry.Parent())
		key := entry.Key()

		newValue := entry.Value
		if patch.Target.Format == model.ConfigFormatCFG {
			newValue = `"` + entry.Value + `"`
		}

		sectionMap, sectionExists := current[section]
		if !sectionExists {
			diff.Changes = append(diff.Changes, ConfigChange{
				Type:     ChangeAdd,
				Path:     entry.Path,
				NewValue: newValue,
			})
			continue
		}

		oldValue, keyExists := sectionMap[key]
		if !keyExists {
			diff.Changes = append(diff.Changes, ConfigChange{
				Type:     ChangeAdd,
				Path:     entry.Path,
				NewValue: newValue,
			})
			continue
		}

		if configformat.NormalizePath(oldValue, d.homeDir()) != configformat.NormalizePath(newValue, d.homeDir()) {
			diff.Changes = append(diff.Changes, ConfigChange{
				Type:     ChangeModify,
				Path:     entry.Path,
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
		if d.UserModified && len(d.UserChanges) > 0 {
			sb.WriteString(fmt.Sprintf("    %s\n", yellow("⚠ You modified settings managed by kyaraben:")))
			for _, uc := range d.UserChanges {
				key := uc.Path[len(uc.Path)-1]
				sb.WriteString(fmt.Sprintf("      %s: %s → %s\n", key, dim(uc.BaselineValue), yellow(uc.CurrentValue)))
			}
		}
		return sb.String()
	} else {
		sb.WriteString(fmt.Sprintf("  %s %s\n", yellow("MODIFY"), bold(d.Path)))
	}

	if d.UserModified && len(d.UserChanges) > 0 {
		sb.WriteString(fmt.Sprintf("    %s\n", yellow("⚠ You modified settings managed by kyaraben (will be overwritten):")))
		for _, uc := range d.UserChanges {
			key := uc.Path[len(uc.Path)-1]
			sb.WriteString(fmt.Sprintf("      %s: %s → %s\n", key, dim(uc.BaselineValue), yellow(uc.CurrentValue)))
		}
		sb.WriteString("\n")
	}

	sectionChanges := make(map[string][]ConfigChange)
	sectionOrder := make([]string, 0)
	for _, change := range d.Changes {
		section := change.Section()
		if _, exists := sectionChanges[section]; !exists {
			sectionOrder = append(sectionOrder, section)
		}
		sectionChanges[section] = append(sectionChanges[section], change)
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
					indent, green("+"), change.Key(), green(change.NewValue)))
			case ChangeModify:
				sb.WriteString(fmt.Sprintf("%s%s %s\n",
					indent, yellow("~"), change.Key()))
				sb.WriteString(fmt.Sprintf("%s    %s %s\n",
					indent, red("-"), dim(change.OldValue)))
				sb.WriteString(fmt.Sprintf("%s    %s %s\n",
					indent, green("+"), change.NewValue))
			case ChangeRemove:
				sb.WriteString(fmt.Sprintf("%s%s %s = %s\n",
					indent, red("-"), change.Key(), red(change.OldValue)))
			}
		}
	}

	return sb.String()
}

func (d *DiffComputer) readConfig(path string, format model.ConfigFormat) (map[string]map[string]string, error) {
	handler := configformat.NewHandler(d.fs, format)
	return handler.Read(path)
}

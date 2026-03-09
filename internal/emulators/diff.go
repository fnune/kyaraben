package emulators

import (
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
	return NewDiffComputer(vfs.OSFS, model.NewDefaultResolver())
}

type ConfigDiff struct {
	Path            string
	IsNewFile       bool
	UserModified    bool
	KyarabenChanged bool
	Changes         []ConfigChange
	UserChanges     []UserChange
	VersionUpgrades []VersionUpgrade
}

type UserChange struct {
	Key          string
	WrittenValue string
	CurrentValue string
}

type VersionUpgrade struct {
	Key      string
	OldValue string
	NewValue string
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

type DiffContext struct {
	CurrentConfigInputs map[string]string
}

func ComputeDiffWithBaseline(patch model.ConfigPatch, baseline *model.ManagedConfig) (*ConfigDiff, error) {
	return NewDefaultDiffComputer().ComputeDiffWithBaseline(patch, baseline, nil)
}

func (d *DiffComputer) ComputeDiffWithBaseline(patch model.ConfigPatch, baseline *model.ManagedConfig, ctx *DiffContext) (*ConfigDiff, error) {
	diff, err := d.ComputeDiff(patch)
	if err != nil {
		return nil, err
	}

	if baseline == nil || diff.IsNewFile {
		return diff, nil
	}

	if len(baseline.WrittenEntries) == 0 {
		return diff, nil
	}

	current, err := d.readConfig(diff.Path, patch.Target.Format)
	if err != nil {
		return diff, nil
	}

	entryByPath := make(map[string]model.ConfigEntry)
	for _, entry := range patch.Entries {
		entryByPath[entry.FullPath()] = entry
	}

	for _, entry := range patch.Entries {
		if entry.DefaultOnly {
			continue
		}

		entryKey := entry.FullPath()
		writtenValue, wasWritten := baseline.WrittenEntries[entryKey]
		if !wasWritten {
			continue
		}

		section := configformat.SectionKey(entry.Parent())
		key := entry.Key()

		var currentValue string
		if sectionMap, ok := current[section]; ok {
			currentValue = sectionMap[key]
		}

		newValue := entry.Value

		userModifiedThisKey := !valuesEqualNormalized(writtenValue, currentValue, entry.EqualityFunc, d.homeDir())
		if userModifiedThisKey {
			diff.UserModified = true
			diff.UserChanges = append(diff.UserChanges, UserChange{
				Key:          key,
				WrittenValue: writtenValue,
				CurrentValue: currentValue,
			})
		}

		if writtenValue != newValue && !userModifiedThisKey {
			if isUIDrivenChange(entry, baseline.ConfigInputsWhenWritten, ctx) {
				continue
			}
			diff.KyarabenChanged = true
			diff.VersionUpgrades = append(diff.VersionUpgrades, VersionUpgrade{
				Key:      key,
				OldValue: writtenValue,
				NewValue: newValue,
			})
		}
	}

	return diff, nil
}

func isUIDrivenChange(entry model.ConfigEntry, configInputsWhenWritten map[string]string, ctx *DiffContext) bool {
	if len(entry.DependsOn) == 0 {
		return false
	}
	if ctx == nil || ctx.CurrentConfigInputs == nil {
		return false
	}
	if configInputsWhenWritten == nil {
		return false
	}

	for _, dep := range entry.DependsOn {
		key := string(dep)
		oldValue, hadOld := configInputsWhenWritten[key]
		newValue, hasNew := ctx.CurrentConfigInputs[key]

		if hadOld != hasNew {
			return true
		}
		if hadOld && oldValue != newValue {
			return true
		}
	}
	return false
}

func valuesEqualNormalized(a, b string, equalityFunc model.ValueEqualityFunc, homeDir string) bool {
	if equalityFunc != nil {
		return equalityFunc(a, b)
	}
	return configformat.NormalizePath(a, homeDir) == configformat.NormalizePath(b, homeDir)
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

		if !valuesEqual(entry, oldValue, d.homeDir()) {
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
			sb.WriteString(fmt.Sprintf("    %s\n", yellow("Your changes will be overwritten:")))
			for _, uc := range d.UserChanges {
				currentDisplay := uc.CurrentValue
				if currentDisplay == "" {
					currentDisplay = "(deleted)"
				}
				sb.WriteString(fmt.Sprintf("      %s: yours: %s -> kyaraben: %s\n", uc.Key, yellow(currentDisplay), dim(uc.WrittenValue)))
			}
		}
		return sb.String()
	} else {
		sb.WriteString(fmt.Sprintf("  %s %s\n", yellow("MODIFY"), bold(d.Path)))
	}

	if d.UserModified && len(d.UserChanges) > 0 {
		sb.WriteString(fmt.Sprintf("    %s\n", yellow("Your changes will be overwritten:")))
		for _, uc := range d.UserChanges {
			currentDisplay := uc.CurrentValue
			if currentDisplay == "" {
				currentDisplay = "(deleted)"
			}
			sb.WriteString(fmt.Sprintf("      %s: yours: %s -> kyaraben: %s\n", uc.Key, yellow(currentDisplay), dim(uc.WrittenValue)))
		}
		sb.WriteString("\n")
	}

	if d.KyarabenChanged && len(d.VersionUpgrades) > 0 {
		sb.WriteString(fmt.Sprintf("    %s\n", green("Kyaraben has new defaults:")))
		for _, vu := range d.VersionUpgrades {
			newDisplay := vu.NewValue
			if newDisplay == "" {
				newDisplay = "(removed)"
			}
			sb.WriteString(fmt.Sprintf("      %s: was: %s -> becomes: %s\n", vu.Key, dim(vu.OldValue), green(newDisplay)))
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

func valuesEqual(entry model.ConfigEntry, existingValue, homeDir string) bool {
	if entry.EqualityFunc != nil {
		return entry.EqualityFunc(entry.Value, existingValue)
	}
	return configformat.NormalizePath(entry.Value, homeDir) == configformat.NormalizePath(existingValue, homeDir)
}

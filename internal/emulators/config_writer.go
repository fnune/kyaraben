package emulators

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/model"
	"gopkg.in/yaml.v3"

	"github.com/beevik/etree"
)

type ConfigWriter struct{}

func NewConfigWriter() *ConfigWriter {
	return &ConfigWriter{}
}

type ApplyResult struct {
	Path         string
	BaselineHash string
	BackupPath   string
}

type ApplyOptions struct {
	CreateBackup bool
}

func (w *ConfigWriter) NeedsBackup(patch model.ConfigPatch) (string, bool, error) {
	path, err := patch.Target.Resolve()
	if err != nil {
		return "", false, fmt.Errorf("resolving config path: %w", err)
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return path, false, nil
	}
	if err != nil {
		return path, false, err
	}
	return path, true, nil
}

func (w *ConfigWriter) createBackup(path string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s.bak", path, timestamp)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file for backup: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("writing backup file: %w", err)
	}

	return backupPath, nil
}

func (w *ConfigWriter) Apply(patch model.ConfigPatch) (ApplyResult, error) {
	return w.ApplyWithOptions(patch, ApplyOptions{})
}

func (w *ConfigWriter) ApplyWithOptions(patch model.ConfigPatch, opts ApplyOptions) (ApplyResult, error) {
	path, err := patch.Target.Resolve()
	if err != nil {
		return ApplyResult{}, fmt.Errorf("resolving config path: %w", err)
	}

	var backupPath string
	if opts.CreateBackup {
		if _, err := os.Stat(path); err == nil {
			backupPath, err = w.createBackup(path)
			if err != nil {
				return ApplyResult{}, fmt.Errorf("creating backup: %w", err)
			}
		}
	}

	var result ApplyResult
	switch patch.Target.Format {
	case model.ConfigFormatCFG:
		result, err = w.applyCFG(path, patch.Entries)
	case model.ConfigFormatINI:
		result, err = w.applyINI(path, patch.Entries)
	case model.ConfigFormatYAML:
		result, err = w.applyYAML(path, patch.Entries)
	case model.ConfigFormatXML:
		result, err = w.applyXML(path, patch.Entries)
	default:
		return ApplyResult{}, fmt.Errorf("unsupported config format: %s", patch.Target.Format)
	}

	if err != nil {
		return result, err
	}

	result.BackupPath = backupPath
	return result, nil
}

func iniSection(path []string) string {
	if len(path) == 0 {
		return ""
	}
	return strings.Join(path, ".")
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (w *ConfigWriter) applyCFG(path string, entries []model.ConfigEntry) (ApplyResult, error) {
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
		// CFG format requires quoted values (RetroArch style)
		existing[entry.Key()] = `"` + entry.Value + `"`
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

func (w *ConfigWriter) applyINI(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
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

	for _, entry := range entries {
		section := iniSection(entry.Parent())
		if sections[section] == nil {
			sections[section] = make(map[string]string)
		}
		sections[section][entry.Key()] = entry.Value
	}

	f, err := os.Create(path)
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

	hash, err := hashFile(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

func (w *ConfigWriter) applyYAML(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, &existing); err != nil {
			return ApplyResult{}, fmt.Errorf("parsing existing YAML: %w", err)
		}
	}

	for _, entry := range entries {
		setNestedValue(existing, entry.Path, entry.Value)
	}

	f, err := os.Create(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("creating config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintln(f, "# Configuration managed by kyaraben")
	_, _ = fmt.Fprintln(f, "# Manual changes will be preserved on next apply")
	_, _ = fmt.Fprintln(f)

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(existing); err != nil {
		return ApplyResult{}, fmt.Errorf("encoding YAML: %w", err)
	}
	_ = encoder.Close()

	hash, err := hashFile(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

func setNestedValue(m map[string]interface{}, path []string, value string) {
	if len(path) == 0 {
		return
	}

	if len(path) == 1 {
		m[path[0]] = value
		return
	}

	key := path[0]
	if m[key] == nil {
		m[key] = make(map[string]interface{})
	}

	if nested, ok := m[key].(map[string]interface{}); ok {
		setNestedValue(nested, path[1:], value)
	} else {
		nested := make(map[string]interface{})
		m[key] = nested
		setNestedValue(nested, path[1:], value)
	}
}

func (w *ConfigWriter) applyXML(path string, entries []model.ConfigEntry) (ApplyResult, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ApplyResult{}, fmt.Errorf("creating config directory: %w", err)
	}

	doc := etree.NewDocument()
	if data, err := os.ReadFile(path); err == nil {
		if err := doc.ReadFromBytes(data); err != nil {
			doc = etree.NewDocument()
		}
	}

	for _, entry := range entries {
		setXMLValue(doc, entry.Path, entry.Value)
	}

	doc.Indent(2)
	if err := doc.WriteToFile(path); err != nil {
		return ApplyResult{}, fmt.Errorf("writing XML file: %w", err)
	}

	hash, err := hashFile(path)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("hashing config file: %w", err)
	}

	return ApplyResult{Path: path, BaselineHash: hash}, nil
}

func setXMLValue(doc *etree.Document, path []string, value string) {
	if len(path) == 0 {
		return
	}

	root := doc.Root()
	if root == nil {
		root = doc.CreateElement(path[0])
		path = path[1:]
	} else if root.Tag != path[0] {
		root = doc.CreateElement(path[0])
		path = path[1:]
	} else {
		path = path[1:]
	}

	elem := root
	for i, key := range path {
		isLast := i == len(path)-1
		child := elem.SelectElement(key)
		if child == nil {
			child = elem.CreateElement(key)
		}
		if isLast {
			child.SetText(value)
		}
		elem = child
	}
}

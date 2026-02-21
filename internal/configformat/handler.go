package configformat

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"slices"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
)

type ApplyResult struct {
	Path         string
	BaselineHash string
	PatchHash    string
}

type Handler interface {
	Read(path string) (map[string]map[string]string, error)
	Apply(path string, entries []model.ConfigEntry, managedRegions []model.ManagedRegion) (ApplyResult, error)
}

func GetHandler(format model.ConfigFormat) Handler {
	return NewHandler(vfs.OSFS, format)
}

func NewHandler(fs vfs.FS, format model.ConfigFormat) Handler {
	switch format {
	case model.ConfigFormatINI:
		return &iniHandler{fs: fs}
	case model.ConfigFormatCFG:
		return &cfgHandler{fs: fs}
	case model.ConfigFormatTOML:
		return &tomlHandler{fs: fs}
	case model.ConfigFormatYAML:
		return &yamlHandler{fs: fs}
	case model.ConfigFormatXML:
		return &xmlHandler{fs: fs}
	case model.ConfigFormatXMLAttr:
		return &xmlAttrHandler{fs: fs}
	case model.ConfigFormatRaw:
		return &rawHandler{fs: fs}
	default:
		return &iniHandler{fs: fs}
	}
}

func snapshotSections(sections map[string]map[string]string) map[string]map[string]string {
	snap := make(map[string]map[string]string, len(sections))
	for section, keys := range sections {
		snapKeys := make(map[string]string, len(keys))
		for k, v := range keys {
			snapKeys[k] = v
		}
		snap[section] = snapKeys
	}
	return snap
}

func deleteManagedKeys(sections map[string]map[string]string, regions []model.ManagedRegion) {
	for _, region := range regions {
		sr, ok := region.(model.SectionRegion)
		if !ok {
			continue
		}
		sectionMap, exists := sections[sr.Section]
		if !exists {
			continue
		}
		if sr.KeyPrefix == "" {
			for k := range sectionMap {
				delete(sectionMap, k)
			}
			continue
		}
		for k := range sectionMap {
			if strings.HasPrefix(k, sr.KeyPrefix) {
				delete(sectionMap, k)
			}
		}
	}
}

func SectionKey(path []string) string {
	if len(path) == 0 {
		return ""
	}
	return strings.Join(path, ".")
}

func Unquote(v string) string {
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		return v[1 : len(v)-1]
	}
	return v
}

func NormalizePath(v, homeDir string) string {
	v = Unquote(v)
	if strings.HasPrefix(v, "~/") && homeDir != "" {
		return filepath.Join(homeDir, v[2:])
	}
	return v
}

// BindingValuesEqual compares two binding strings (key:value,key:value,...)
// for semantic equality, ignoring key ordering. Returns true if both strings
// parse to the same key-value pairs. Use this as a ConfigEntry.EqualityFunc
// for emulator binding values where key ordering is nondeterministic.
func BindingValuesEqual(a, b string) bool {
	mapA, okA := parseBindingString(a)
	mapB, okB := parseBindingString(b)

	if !okA || !okB {
		return Unquote(a) == Unquote(b)
	}

	return bindingMapsEqual(mapA, mapB)
}

// parseBindingString parses a string like "engine:sdl,port:0,guid:X,button:1"
// into a map. Returns ok=false if the string doesn't look like a binding.
func parseBindingString(s string) (map[string]string, bool) {
	s = Unquote(s)
	if s == "" || !strings.Contains(s, ":") {
		return nil, false
	}

	result := make(map[string]string)
	parts := strings.Split(s, ",")

	for _, part := range parts {
		idx := strings.Index(part, ":")
		if idx <= 0 {
			return nil, false
		}
		key := part[:idx]
		value := part[idx+1:]
		result[key] = value
	}

	return result, len(result) > 0
}

func bindingMapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || va != vb {
			return false
		}
	}
	return true
}

func hashFileWithFS(fs vfs.FS, path string) (string, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func flattenNestedMap(m map[string]interface{}, prefix []string, result map[string]map[string]string) {
	for k, v := range m {
		path := append(prefix, k)
		switch val := v.(type) {
		case map[string]interface{}:
			flattenNestedMap(val, path, result)
		case string:
			section := SectionKey(path[:len(path)-1])
			if result[section] == nil {
				result[section] = make(map[string]string)
			}
			result[section][k] = val
		default:
			section := SectionKey(path[:len(path)-1])
			if result[section] == nil {
				result[section] = make(map[string]string)
			}
			result[section][k] = toString(val)
		}
	}
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return intToString(val)
	case int64:
		return int64ToString(val)
	case float64:
		return float64ToString(val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func int64ToString(n int64) string {
	return intToString(int(n))
}

func float64ToString(f float64) string {
	if f == float64(int64(f)) {
		return int64ToString(int64(f))
	}
	return strings.TrimRight(strings.TrimRight(floatFormat(f), "0"), ".")
}

func floatFormat(f float64) string {
	neg := f < 0
	if neg {
		f = -f
	}
	intPart := int64(f)
	fracPart := f - float64(intPart)

	result := int64ToString(intPart)
	if fracPart > 0 {
		result += "."
		for i := 0; i < 6 && fracPart > 0; i++ {
			fracPart *= 10
			digit := int(fracPart)
			result += string(byte('0' + digit))
			fracPart -= float64(digit)
		}
	}
	if neg {
		result = "-" + result
	}
	return result
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

func hasNestedValue(m map[string]interface{}, path []string) bool {
	if len(path) == 0 {
		return false
	}

	if len(path) == 1 {
		_, exists := m[path[0]]
		return exists
	}

	key := path[0]
	if nested, ok := m[key].(map[string]interface{}); ok {
		return hasNestedValue(nested, path[1:])
	}
	return false
}

func isFullyManaged(regions []model.ManagedRegion) bool {
	for _, r := range regions {
		if _, ok := r.(model.FileRegion); ok {
			return true
		}
	}
	return false
}

func ComputePatchHash(entries []model.ConfigEntry) string {
	sorted := make([]model.ConfigEntry, len(entries))
	copy(sorted, entries)
	slices.SortFunc(sorted, func(a, b model.ConfigEntry) int {
		pathA := strings.Join(a.Path, ".")
		pathB := strings.Join(b.Path, ".")
		if pathA < pathB {
			return -1
		}
		if pathA > pathB {
			return 1
		}
		return 0
	})

	h := sha256.New()
	for _, e := range sorted {
		h.Write([]byte(strings.Join(e.Path, ".")))
		h.Write([]byte{0})
		h.Write([]byte(e.Value))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ConfigFormat string

const (
	ConfigFormatINI     ConfigFormat = "ini"
	ConfigFormatTOML    ConfigFormat = "toml"
	ConfigFormatCFG     ConfigFormat = "cfg"
	ConfigFormatXML     ConfigFormat = "xml"
	ConfigFormatXMLAttr ConfigFormat = "xml_attr"
	ConfigFormatJSON    ConfigFormat = "json"
	ConfigFormatYAML    ConfigFormat = "yaml"
	ConfigFormatRaw     ConfigFormat = "raw"
)

type ConfigBaseDir string

const (
	ConfigBaseDirUserConfig ConfigBaseDir = "user_config"
	ConfigBaseDirUserData   ConfigBaseDir = "user_data"
	ConfigBaseDirHome       ConfigBaseDir = "home"
)

type ConfigTarget struct {
	RelPath string
	Format  ConfigFormat
	BaseDir ConfigBaseDir
}

type BaseDirResolver interface {
	UserConfigDir() (string, error)
	UserHomeDir() (string, error)
	UserDataDir() (string, error)
}

type osBaseDirResolver struct{}

func (osBaseDirResolver) UserConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	return os.UserConfigDir()
}

func (osBaseDirResolver) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (osBaseDirResolver) UserDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

func NewDefaultResolver() BaseDirResolver {
	return osBaseDirResolver{}
}

func (ct ConfigTarget) ResolveWith(resolver BaseDirResolver) (string, error) {
	var baseDir string

	switch ct.BaseDir {
	case ConfigBaseDirUserConfig:
		dir, err := resolver.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("getting user config dir: %w", err)
		}
		baseDir = dir

	case ConfigBaseDirUserData:
		dir, err := resolver.UserDataDir()
		if err != nil {
			return "", fmt.Errorf("getting user data dir: %w", err)
		}
		baseDir = dir

	case ConfigBaseDirHome:
		home, err := resolver.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home dir: %w", err)
		}
		baseDir = home

	default:
		return "", fmt.Errorf("unknown config base dir: %s", ct.BaseDir)
	}

	return filepath.Join(baseDir, ct.RelPath), nil
}

func (ct ConfigTarget) ResolveDirWith(resolver BaseDirResolver) (string, error) {
	if ct.RelPath == "" {
		return "", fmt.Errorf("empty RelPath")
	}

	if !strings.Contains(ct.RelPath, string(filepath.Separator)) {
		return "", fmt.Errorf("RelPath %q has no subdirectory, refusing to resolve", ct.RelPath)
	}

	parts := strings.SplitN(ct.RelPath, string(filepath.Separator), 2)
	topDir := parts[0]
	if topDir == "" || topDir == "." || topDir == ".." {
		return "", fmt.Errorf("invalid top-level directory in RelPath: %q", ct.RelPath)
	}

	var baseDir string
	switch ct.BaseDir {
	case ConfigBaseDirUserConfig:
		dir, err := resolver.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("getting user config dir: %w", err)
		}
		baseDir = dir
	case ConfigBaseDirUserData:
		dir, err := resolver.UserDataDir()
		if err != nil {
			return "", fmt.Errorf("getting user data dir: %w", err)
		}
		baseDir = dir
	case ConfigBaseDirHome:
		dir, err := resolver.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home dir: %w", err)
		}
		baseDir = dir
	default:
		return "", fmt.Errorf("unknown config base dir: %s", ct.BaseDir)
	}

	result := filepath.Join(baseDir, topDir)
	return ct.rejectDangerousPaths(resolver, result)
}

func (ct ConfigTarget) rejectDangerousPaths(resolver BaseDirResolver, path string) (string, error) {
	cleanPath := filepath.Clean(path)

	if cleanPath == "/" || cleanPath == "." {
		return "", fmt.Errorf("refusing to resolve to root or current directory: %q", path)
	}

	home, err := resolver.UserHomeDir()
	if err == nil {
		if cleanPath == filepath.Clean(home) {
			return "", fmt.Errorf("refusing to resolve to home directory: %q", path)
		}
	}

	configDir, err := resolver.UserConfigDir()
	if err == nil {
		if cleanPath == filepath.Clean(configDir) {
			return "", fmt.Errorf("refusing to resolve to config directory: %q", path)
		}
	}

	dataDir, err := resolver.UserDataDir()
	if err == nil {
		if cleanPath == filepath.Clean(dataDir) {
			return "", fmt.Errorf("refusing to resolve to data directory: %q", path)
		}
	}

	return cleanPath, nil
}

// ValueEqualityFunc compares two config values for equality.
// Used for semantic comparison of values like binding strings where
// key ordering may vary but the values are functionally equivalent.
type ValueEqualityFunc func(a, b string) bool

// ConfigInput identifies a user-configurable input that affects generated config values.
// Used to distinguish UI-driven changes from version upgrades during diff detection.
// Constants are defined near their corresponding schema fields (e.g., in config.go).
type ConfigInput string

// ConfigInputNone indicates an entry has no dynamic dependencies.
// Use this for static values that don't change based on user settings.
const ConfigInputNone ConfigInput = "none"

// Shorthand dependency lists for common cases.
var (
	None     = []ConfigInput{ConfigInputNone}
	Store    = []ConfigInput{ConfigInputUserStore}
	Nintendo = []ConfigInput{ConfigInputNintendoConfirm}
)

// Path builds a config path from variadic string arguments.
func Path(parts ...string) []string { return parts }

// Entry creates a ConfigEntry with the given dependencies, path, and value.
// Dependencies must be specified explicitly to ensure correct diff detection.
func Entry(deps []ConfigInput, path []string, value string) ConfigEntry {
	return ConfigEntry{Path: path, Value: value, DependsOn: deps}
}

// Default creates a DefaultOnly ConfigEntry that only sets the value if the key doesn't exist.
func Default(deps []ConfigInput, path []string, value string) ConfigEntry {
	return ConfigEntry{Path: path, Value: value, DependsOn: deps, DefaultOnly: true}
}

type ConfigEntry struct {
	Path         []string
	Value        string
	DefaultOnly  bool              // Only set if key doesn't exist; user changes are preserved
	EqualityFunc ValueEqualityFunc // Optional custom equality check; nil uses string comparison
	DependsOn    []ConfigInput     // Config inputs this entry depends on; use None for static values
}

// ManagedRegion describes a portion of a config file that kyaraben manages.
// On apply, existing content within the region is cleared before writing entries.
//
// Implementations:
//   - FileRegion: the entire file is managed
//   - SectionRegion: a section (with optional key prefix) is managed
type ManagedRegion interface {
	managedRegion()
}

// FileRegion indicates that kyaraben manages the entire config file.
// The file is deleted and rewritten from scratch on each apply.
type FileRegion struct{}

func (FileRegion) managedRegion() {}

// SectionRegion indicates that kyaraben manages a section of a config file.
// On apply, keys matching the prefix within the section are cleared before writing.
type SectionRegion struct {
	Section   string // INI section name. Empty for flat formats (CFG, TOML root keys).
	KeyPrefix string // Keys starting with this prefix are managed. Empty means the entire section.
}

func (e ConfigEntry) Key() string {
	if len(e.Path) == 0 {
		return ""
	}
	return e.Path[len(e.Path)-1]
}

func (e ConfigEntry) Parent() []string {
	if len(e.Path) <= 1 {
		return nil
	}
	return e.Path[:len(e.Path)-1]
}

func (e ConfigEntry) FullPath() string {
	return strings.Join(e.Path, ".")
}

func (SectionRegion) managedRegion() {}

type ConfigPatch struct {
	Target         ConfigTarget
	Entries        []ConfigEntry
	ManagedRegions []ManagedRegion
	Delete         bool
}

// ManagesWholeFile reports whether any managed region covers the entire file.
func (p ConfigPatch) ManagesWholeFile() bool {
	for _, r := range p.ManagedRegions {
		if _, ok := r.(FileRegion); ok {
			return true
		}
	}
	return false
}

// WithDependsOn returns copies of the entries with DependsOn set to the given dependencies.
func WithDependsOn(entries []ConfigEntry, deps ...ConfigInput) []ConfigEntry {
	result := make([]ConfigEntry, len(entries))
	for i, e := range entries {
		result[i] = e
		result[i].DependsOn = deps
	}
	return result
}

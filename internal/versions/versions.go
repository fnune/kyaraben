package versions

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

//go:embed versions.toml
var versionsData string

// Versions holds all version information parsed from versions.toml.
type Versions struct {
	Nixpkgs   NixpkgsVersion          `toml:"nixpkgs"`
	Emulators map[string]EmulatorSpec // Populated after parsing
}

type NixpkgsVersion struct {
	Commit string `toml:"commit"`
}

// EmulatorSpec describes all available versions of an emulator.
type EmulatorSpec struct {
	URLTemplate string                  // URL template with {version}, {target}, {release_tag}, {arch} placeholders
	BinaryPath  string                  // Default binary path for archives
	Default     string                  // Default version string
	IconURL     string                  // URL to download icon from
	IconSHA256  string                  // SHA256 hash of icon
	Versions    map[string]VersionEntry // Map of version string to entry
}

// GetVersion returns a specific version entry, or nil if not found.
func (e *EmulatorSpec) GetVersion(version string) *VersionEntry {
	if v, ok := e.Versions[version]; ok {
		return &v
	}
	return nil
}

// GetDefault returns the default version entry.
func (e *EmulatorSpec) GetDefault() *VersionEntry {
	return e.GetVersion(e.Default)
}

// AvailableVersions returns all version strings.
func (e *EmulatorSpec) AvailableVersions() []string {
	versions := make([]string, 0, len(e.Versions))
	for v := range e.Versions {
		versions = append(versions, v)
	}
	return versions
}

// VersionEntry describes a specific version of an emulator.
type VersionEntry struct {
	Version    string                 // The version string (copied from map key)
	ReleaseTag string                 // For repos where tag != version
	BinaryPath string                 // Version-specific binary path override
	Targets    map[string]TargetBuild // Map of target name to build info
}

// EffectiveReleaseTag returns the release tag to use in URLs.
// Falls back to Version if ReleaseTag is not set.
func (v *VersionEntry) EffectiveReleaseTag() string {
	if v.ReleaseTag != "" {
		return v.ReleaseTag
	}
	return v.Version
}

// URL returns the download URL for a given target.
func (v *VersionEntry) URL(target string, spec *EmulatorSpec) string {
	url := strings.ReplaceAll(spec.URLTemplate, "{version}", v.Version)
	url = strings.ReplaceAll(url, "{target}", target)
	url = strings.ReplaceAll(url, "{release_tag}", v.EffectiveReleaseTag())
	if t := v.Target(target); t != nil {
		url = strings.ReplaceAll(url, "{arch}", t.Arch)
	}
	return url
}

// Target returns the target build info by name, or nil if not found.
func (v *VersionEntry) Target(name string) *TargetBuild {
	if t, ok := v.Targets[name]; ok {
		return &t
	}
	return nil
}

// TargetsForArch returns all target names available for a given architecture.
func (v *VersionEntry) TargetsForArch(arch string) []string {
	var targets []string
	for name, t := range v.Targets {
		if t.Arch == arch {
			targets = append(targets, name)
		}
	}
	return targets
}

// DefaultTargetForArch returns the first target matching the given architecture.
func (v *VersionEntry) DefaultTargetForArch(arch string) string {
	for name, t := range v.Targets {
		if t.Arch == arch {
			return name
		}
	}
	return ""
}

// BinaryPathForTarget returns the path to the binary inside an archive.
// Priority: target-specific > version-specific > spec-level > empty.
func (v *VersionEntry) BinaryPathForTarget(target string, spec *EmulatorSpec) string {
	if t := v.Target(target); t != nil && t.BinaryPath != "" {
		return t.BinaryPath
	}
	if v.BinaryPath != "" {
		return v.BinaryPath
	}
	return spec.BinaryPath
}

// ArchiveType returns the archive type based on URL extension.
func (v *VersionEntry) ArchiveType(target string, spec *EmulatorSpec) string {
	url := v.URL(target, spec)
	switch {
	case strings.HasSuffix(url, ".7z"):
		return "7z"
	case strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(url, ".tar.xz"):
		return "tar.xz"
	case strings.HasSuffix(url, ".zip"):
		return "zip"
	default:
		return ""
	}
}

// TargetBuild describes a build for a specific hardware target.
type TargetBuild struct {
	Arch       string `toml:"arch"`
	SHA256     string `toml:"sha256"`
	BinaryPath string `toml:"binary_path"`
}

// Known emulator names for parsing
var emulatorNames = []string{
	"nix-portable",
	"eden",
	"duckstation",
	"pcsx2",
	"ppsspp",
	"mgba",
	"cemu",
	"azahar",
	"dolphin",
	"melonds",
	"vita3k",
	"rpcs3",
	"flycast",
	"retroarch",
}

// GetEmulator returns the EmulatorSpec for a given emulator name.
func (v *Versions) GetEmulator(name string) (*EmulatorSpec, bool) {
	spec, ok := v.Emulators[name]
	if !ok {
		return nil, false
	}
	return &spec, true
}

var parsed *Versions

func Get() (*Versions, error) {
	if parsed != nil {
		return parsed, nil
	}

	v, err := parse(versionsData)
	if err != nil {
		return nil, err
	}

	parsed = v
	return parsed, nil
}

func MustGet() *Versions {
	v, err := Get()
	if err != nil {
		panic(err)
	}
	return v
}

// ResetCache clears the cached versions (for testing).
func ResetCache() {
	parsed = nil
}

// parse parses the versions.toml content into Versions struct.
func parse(data string) (*Versions, error) {
	// First pass: decode into generic map
	var raw map[string]interface{}
	if _, err := toml.Decode(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing versions.toml: %w", err)
	}

	v := &Versions{
		Emulators: make(map[string]EmulatorSpec),
	}

	// Parse nixpkgs
	if nixpkgsRaw, ok := raw["nixpkgs"].(map[string]interface{}); ok {
		if commit, ok := nixpkgsRaw["commit"].(string); ok {
			v.Nixpkgs.Commit = commit
		}
	}

	// Parse each emulator
	for _, name := range emulatorNames {
		emuRaw, ok := raw[name].(map[string]interface{})
		if !ok {
			continue
		}

		spec, err := parseEmulatorSpec(emuRaw)
		if err != nil {
			return nil, fmt.Errorf("parsing emulator %s: %w", name, err)
		}
		v.Emulators[name] = spec
	}

	return v, nil
}

// parseEmulatorSpec parses a single emulator's specification from raw TOML data.
func parseEmulatorSpec(raw map[string]interface{}) (EmulatorSpec, error) {
	spec := EmulatorSpec{
		Versions: make(map[string]VersionEntry),
	}

	// Extract known fields
	if v, ok := raw["url_template"].(string); ok {
		spec.URLTemplate = v
	}
	if v, ok := raw["binary_path"].(string); ok {
		spec.BinaryPath = v
	}
	if v, ok := raw["default"].(string); ok {
		spec.Default = v
	}
	if v, ok := raw["icon_url"].(string); ok {
		spec.IconURL = v
	}
	if v, ok := raw["icon_sha256"].(string); ok {
		spec.IconSHA256 = v
	}

	// Everything else is a version entry
	knownKeys := map[string]bool{
		"url_template": true,
		"binary_path":  true,
		"default":      true,
		"icon_url":     true,
		"icon_sha256":  true,
	}

	for key, value := range raw {
		if knownKeys[key] {
			continue
		}

		// This should be a version entry
		versionRaw, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		entry, err := parseVersionEntry(key, versionRaw)
		if err != nil {
			return spec, fmt.Errorf("parsing version %s: %w", key, err)
		}
		spec.Versions[key] = entry
	}

	// If no explicit default and we have versions, use the first one found
	// (though this shouldn't happen with proper TOML)
	if spec.Default == "" && len(spec.Versions) > 0 {
		for v := range spec.Versions {
			spec.Default = v
			break
		}
	}

	return spec, nil
}

// parseVersionEntry parses a single version entry from raw TOML data.
func parseVersionEntry(version string, raw map[string]interface{}) (VersionEntry, error) {
	entry := VersionEntry{
		Version: version,
		Targets: make(map[string]TargetBuild),
	}

	if v, ok := raw["release_tag"].(string); ok {
		entry.ReleaseTag = v
	}
	if v, ok := raw["binary_path"].(string); ok {
		entry.BinaryPath = v
	}

	// Parse targets
	if targetsRaw, ok := raw["targets"].(map[string]interface{}); ok {
		for targetName, targetValue := range targetsRaw {
			targetRaw, ok := targetValue.(map[string]interface{})
			if !ok {
				continue
			}

			target := TargetBuild{}
			if v, ok := targetRaw["arch"].(string); ok {
				target.Arch = v
			}
			if v, ok := targetRaw["sha256"].(string); ok {
				target.SHA256 = v
			}
			if v, ok := targetRaw["binary_path"].(string); ok {
				target.BinaryPath = v
			}

			entry.Targets[targetName] = target
		}
	}

	return entry, nil
}

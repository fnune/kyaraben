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
	Packages map[string]PackageSpec // Populated after parsing
}

// PackageSpec describes all available versions of a package.
type PackageSpec struct {
	ReleasesURL     string                  // URL for checking releases (e.g., "github:owner/repo" or "gitlab:namespace/project")
	URLTemplate     string                  // URL template with {version}, {release_tag}, {variant} placeholders
	BinaryPath      string                  // Default binary path for archives
	Default         string                  // Default version string
	IconURL         string                  // URL to download icon from
	IconSHA256      string                  // SHA256 hash of icon
	InstallType     string                  // Installation type: "" for normal, "retroarch-core" for cores
	BundleSize      int64                   // Download size for shared bundles (retroarch-cores)
	Versions        map[string]VersionEntry // Map of version string to entry
	VersionsInOrder []string                // TOML definition order (newest first)
}

// IsRetroArchCore returns true if this package is a RetroArch core.
func (s *PackageSpec) IsRetroArchCore() bool {
	return s.InstallType == "retroarch-core"
}

// GetVersion returns a specific version entry, or nil if not found.
func (e *PackageSpec) GetVersion(version string) *VersionEntry {
	if v, ok := e.Versions[version]; ok {
		return &v
	}
	return nil
}

// GetDefault returns the default version entry.
func (e *PackageSpec) GetDefault() *VersionEntry {
	return e.GetVersion(e.Default)
}

// AvailableVersions returns all version strings in TOML definition order (newest first).
func (e *PackageSpec) AvailableVersions() []string {
	return e.VersionsInOrder
}

// VersionEntry describes a specific version of a package.
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
func (v *VersionEntry) URL(target string, spec *PackageSpec) string {
	if t := v.Target(target); t != nil && t.URL != "" {
		return t.URL
	}
	url := strings.ReplaceAll(spec.URLTemplate, "{version}", v.Version)
	url = strings.ReplaceAll(url, "{release_tag}", v.EffectiveReleaseTag())
	if t := v.Target(target); t != nil {
		url = strings.ReplaceAll(url, "{variant}", t.Variant)
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

// archFallback maps architectures to their canonical target names.
var archFallback = map[string]TargetName{
	"x86_64":  TargetX64,
	"aarch64": TargetAarch64,
}

// SelectTarget returns the best matching target for the given detected hardware.
// It tries exact match first, then falls back to the canonical target for the architecture.
func (v *VersionEntry) SelectTarget(name, arch string) string {
	if v.Target(name) != nil {
		return name
	}
	if fallback, ok := archFallback[arch]; ok {
		if v.Target(fallback.String()) != nil {
			return fallback.String()
		}
	}
	return ""
}

// BinaryPathForTarget returns the path to the binary inside an archive.
// Priority: target-specific > version-specific > spec-level > empty.
func (v *VersionEntry) BinaryPathForTarget(target string, spec *PackageSpec) string {
	if t := v.Target(target); t != nil && t.BinaryPath != "" {
		return t.BinaryPath
	}
	if v.BinaryPath != "" {
		return v.BinaryPath
	}
	return spec.BinaryPath
}

// ArchiveType returns the archive type based on URL extension.
func (v *VersionEntry) ArchiveType(target string, spec *PackageSpec) string {
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
	Variant    string `toml:"variant"` // String used in URL construction
	SHA256     string `toml:"sha256"`
	BinaryPath string `toml:"binary_path"`
	Size       int64  `toml:"size"`
	URL        string `toml:"url"`
}

// TargetName represents a valid target identifier. The unexported field ensures
// only predefined values can be used - code outside this package cannot construct
// arbitrary TargetName values.
type TargetName struct {
	name string
}

func (t TargetName) String() string { return t.name }

var (
	TargetX64       = TargetName{"x64"}
	TargetAarch64   = TargetName{"aarch64"}
	TargetSteamdeck = TargetName{"steamdeck"}
	TargetRogAlly   = TargetName{"rog-ally"}
)

var validTargets = map[string]TargetName{
	"x64":       TargetX64,
	"aarch64":   TargetAarch64,
	"steamdeck": TargetSteamdeck,
	"rog-ally":  TargetRogAlly,
}

// ParseTargetName validates and returns a TargetName from a string.
// Returns false if the string is not a valid target name.
func ParseTargetName(s string) (TargetName, bool) {
	t, ok := validTargets[s]
	return t, ok
}

// Known package names for parsing
var packageNames = []string{
	// RetroArch cores (from bundle)
	"bsnes",
	"snes9x",
	"mesen",
	"genesis_plus_gx",
	"mupen64plus_next",
	"mednafen_saturn",
	"mednafen_pce_fast",
	"mednafen_ngp",
	"mgba",
	"citra",
	"fbneo",
	"stella",
	"vice_x64sc",
	"melondsds",
	"azahar",
	// Standalone emulators
	"eden",
	"duckstation",
	"pcsx2",
	"ppsspp",
	"cemu",
	"dolphin",
	"vita3k",
	"rpcs3",
	"flycast",
	"xemu",
	"xenia-edge",
	"retroarch",
	// Frontends
	"esde",
	// Utilities
	"syncthing",
}

// GetPackage returns the PackageSpec for a given package name.
func (v *Versions) GetPackage(name string) (*PackageSpec, bool) {
	spec, ok := v.Packages[name]
	if !ok {
		return nil, false
	}
	return &spec, true
}

var parsed *Versions

// Init parses and caches version data. Call from main.go at startup.
func Init() error {
	v, err := parse(versionsData)
	if err != nil {
		return err
	}
	parsed = v
	return nil
}

func Get() (*Versions, error) {
	if parsed == nil {
		return nil, fmt.Errorf("versions not initialized: call versions.Init() at startup")
	}
	return parsed, nil
}

func MustGet() *Versions {
	if parsed == nil {
		panic("versions not initialized: call versions.Init() at startup")
	}
	return parsed
}

// ResetCache clears the cached versions (for testing).
func ResetCache() {
	parsed = nil
}

// parse parses the versions.toml content into Versions struct.
func parse(data string) (*Versions, error) {
	var raw map[string]interface{}
	meta, err := toml.Decode(data, &raw)
	if err != nil {
		return nil, fmt.Errorf("parsing versions.toml: %w", err)
	}

	v := &Versions{
		Packages: make(map[string]PackageSpec),
	}

	for _, name := range packageNames {
		pkgRaw, ok := raw[name].(map[string]interface{})
		if !ok {
			continue
		}

		spec, err := parsePackageSpec(pkgRaw, name, meta)
		if err != nil {
			return nil, fmt.Errorf("parsing package %s: %w", name, err)
		}
		v.Packages[name] = spec
	}

	return v, nil
}

// parsePackageSpec parses a single package specification from raw TOML data.
func parsePackageSpec(raw map[string]interface{}, pkgName string, meta toml.MetaData) (PackageSpec, error) {
	spec := PackageSpec{
		Versions: make(map[string]VersionEntry),
	}

	if v, ok := raw["releases_url"].(string); ok {
		spec.ReleasesURL = v
	}
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
	if v, ok := raw["install_type"].(string); ok {
		spec.InstallType = v
	}
	if v, ok := raw["bundle_size"].(int64); ok {
		spec.BundleSize = v
	}

	knownKeys := map[string]bool{
		"releases_url": true,
		"url_template": true,
		"binary_path":  true,
		"default":      true,
		"icon_url":     true,
		"icon_sha256":  true,
		"install_type": true,
		"bundle_size":  true,
	}

	for key, value := range raw {
		if knownKeys[key] {
			continue
		}

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

	// Extract version keys in TOML definition order
	seen := make(map[string]bool)
	for _, key := range meta.Keys() {
		if len(key) >= 2 && key[0] == pkgName {
			version := key[1]
			if _, isVersion := spec.Versions[version]; isVersion && !seen[version] {
				spec.VersionsInOrder = append(spec.VersionsInOrder, version)
				seen[version] = true
			}
		}
	}

	if spec.Default == "" && len(spec.VersionsInOrder) > 0 {
		spec.Default = spec.VersionsInOrder[0]
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

	if targetsRaw, ok := raw["targets"].(map[string]interface{}); ok {
		for targetName, targetValue := range targetsRaw {
			if _, ok := ParseTargetName(targetName); !ok {
				return entry, fmt.Errorf("invalid target name %q", targetName)
			}
			targetRaw, ok := targetValue.(map[string]interface{})
			if !ok {
				continue
			}

			target := TargetBuild{}
			if v, ok := targetRaw["variant"].(string); ok {
				target.Variant = v
			}
			if v, ok := targetRaw["sha256"].(string); ok {
				target.SHA256 = v
			}
			if v, ok := targetRaw["binary_path"].(string); ok {
				target.BinaryPath = v
			}
			if v, ok := targetRaw["size"].(int64); ok {
				target.Size = v
			}
			if v, ok := targetRaw["url"].(string); ok {
				target.URL = v
			}

			entry.Targets[targetName] = target
		}
	}

	return entry, nil
}

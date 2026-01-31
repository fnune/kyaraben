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
	Nixpkgs     NixpkgsVersion `toml:"nixpkgs"`
	NixPortable EmulatorSpec   `toml:"nix-portable"`
	Eden        EmulatorSpec   `toml:"eden"`
	DuckStation EmulatorSpec   `toml:"duckstation"`
	PCSX2       EmulatorSpec   `toml:"pcsx2"`
	PPSSPP      EmulatorSpec   `toml:"ppsspp"`
	MGBA        EmulatorSpec   `toml:"mgba"`
	Cemu        EmulatorSpec   `toml:"cemu"`
	Azahar      EmulatorSpec   `toml:"azahar"`
	Dolphin     EmulatorSpec   `toml:"dolphin"`
	MelonDS     EmulatorSpec   `toml:"melonds"`
	Vita3K      EmulatorSpec   `toml:"vita3k"`
	RPCS3       EmulatorSpec   `toml:"rpcs3"`
	Flycast     EmulatorSpec   `toml:"flycast"`
	RetroArch   EmulatorSpec   `toml:"retroarch"`
	TIC80       EmulatorSpec   `toml:"tic80"`
}

// GetEmulator returns the EmulatorSpec for a given emulator name.
func (v *Versions) GetEmulator(name string) (*EmulatorSpec, bool) {
	switch name {
	case "nix-portable":
		return &v.NixPortable, true
	case "eden":
		return &v.Eden, true
	case "duckstation":
		return &v.DuckStation, true
	case "pcsx2":
		return &v.PCSX2, true
	case "ppsspp":
		return &v.PPSSPP, true
	case "mgba":
		return &v.MGBA, true
	case "cemu":
		return &v.Cemu, true
	case "azahar":
		return &v.Azahar, true
	case "dolphin":
		return &v.Dolphin, true
	case "melonds":
		return &v.MelonDS, true
	case "vita3k":
		return &v.Vita3K, true
	case "rpcs3":
		return &v.RPCS3, true
	case "flycast":
		return &v.Flycast, true
	case "retroarch":
		return &v.RetroArch, true
	case "tic80":
		return &v.TIC80, true
	default:
		return nil, false
	}
}

type NixpkgsVersion struct {
	Commit string `toml:"commit"`
}

// EmulatorSpec describes all available versions of an emulator.
type EmulatorSpec struct {
	URLTemplate string         `toml:"url_template"`
	BinaryPath  string         `toml:"binary_path"` // default for archives
	Versions    []VersionEntry `toml:"versions"`
}

// GetVersion returns a specific version entry, or nil if not found.
func (e *EmulatorSpec) GetVersion(version string) *VersionEntry {
	for i := range e.Versions {
		if e.Versions[i].Version == version {
			return &e.Versions[i]
		}
	}
	return nil
}

// GetDefault returns the default version entry.
// If no version is marked as default, returns the first one.
func (e *EmulatorSpec) GetDefault() *VersionEntry {
	for i := range e.Versions {
		if e.Versions[i].Default {
			return &e.Versions[i]
		}
	}
	if len(e.Versions) > 0 {
		return &e.Versions[0]
	}
	return nil
}

// AvailableVersions returns all version strings in order.
func (e *EmulatorSpec) AvailableVersions() []string {
	versions := make([]string, len(e.Versions))
	for i, v := range e.Versions {
		versions[i] = v.Version
	}
	return versions
}

// VersionEntry describes a specific version of an emulator.
type VersionEntry struct {
	Version    string        `toml:"version"`
	ReleaseTag string        `toml:"release_tag"` // for repos where tag != version
	Default    bool          `toml:"default"`
	BinaryPath string        `toml:"binary_path"` // version-specific override
	Targets    []TargetBuild `toml:"targets"`
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
	// Also support {arch} placeholder
	if t := v.Target(target); t != nil {
		url = strings.ReplaceAll(url, "{arch}", t.Arch)
	}
	return url
}

// Target returns the target build info by name, or nil if not found.
func (v *VersionEntry) Target(name string) *TargetBuild {
	for i := range v.Targets {
		if v.Targets[i].Name == name {
			return &v.Targets[i]
		}
	}
	return nil
}

// TargetsForArch returns all target names available for a given architecture.
func (v *VersionEntry) TargetsForArch(arch string) []string {
	var targets []string
	for _, t := range v.Targets {
		if t.Arch == arch {
			targets = append(targets, t.Name)
		}
	}
	return targets
}

// DefaultTargetForArch returns the first target matching the given architecture.
func (v *VersionEntry) DefaultTargetForArch(arch string) string {
	for _, t := range v.Targets {
		if t.Arch == arch {
			return t.Name
		}
	}
	return ""
}

// BinaryPathForTarget returns the path to the binary inside an archive.
// Returns empty string for direct AppImage downloads (no extraction needed).
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
// Returns empty string for direct binaries (AppImage, etc).
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
	Name       string `toml:"name"`       // e.g., "amd64", "steamdeck"
	Arch       string `toml:"arch"`       // e.g., "x86_64", "aarch64"
	SHA256     string `toml:"sha256"`
	BinaryPath string `toml:"binary_path"` // target-specific override
}

var parsed *Versions

func Get() (*Versions, error) {
	if parsed != nil {
		return parsed, nil
	}

	var v Versions
	if _, err := toml.Decode(versionsData, &v); err != nil {
		return nil, fmt.Errorf("parsing versions.toml: %w", err)
	}

	parsed = &v
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

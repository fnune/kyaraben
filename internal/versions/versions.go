package versions

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

//go:embed versions.toml
var versionsData string

type Versions struct {
	Nixpkgs     NixpkgsVersion  `toml:"nixpkgs"`
	Eden        AppImageVersion `toml:"eden"`
	DuckStation AppImageVersion `toml:"duckstation"`
	// New standalone emulators replacing nix-portable GUI apps
	PCSX2     AppImageVersion `toml:"pcsx2"`
	PPSSPP    AppImageVersion `toml:"ppsspp"`
	MGBA      AppImageVersion `toml:"mgba"`
	Cemu      AppImageVersion `toml:"cemu"`
	Azahar    AppImageVersion `toml:"azahar"`
	Dolphin   AppImageVersion `toml:"dolphin"`
	RetroArch AppImageVersion `toml:"retroarch"`
	TIC80     AppImageVersion `toml:"tic80"`
}

type NixpkgsVersion struct {
	Commit string `toml:"commit"`
}

type AppImageVersion struct {
	Version     string                 `toml:"version"`
	URLTemplate string                 `toml:"url_template"`
	Targets     map[string]TargetBuild `toml:"targets"`
}

type TargetBuild struct {
	Arch   string `toml:"arch"`
	SHA256 string `toml:"sha256"`
}

func (a *AppImageVersion) DefaultTargetForArch(arch string) string {
	for name, t := range a.Targets {
		if t.Arch == arch {
			return name
		}
	}
	return ""
}

// URL returns the download URL for a given target.
func (a *AppImageVersion) URL(target string) string {
	url := strings.ReplaceAll(a.URLTemplate, "{version}", a.Version)
	url = strings.ReplaceAll(url, "{target}", target)
	return url
}

// Target returns the target build info, or nil if not found.
func (a *AppImageVersion) Target(name string) *TargetBuild {
	t, ok := a.Targets[name]
	if !ok {
		return nil
	}
	return &t
}

// TargetsForArch returns all target names available for a given architecture.
func (a *AppImageVersion) TargetsForArch(arch string) []string {
	var targets []string
	for name, t := range a.Targets {
		if t.Arch == arch {
			targets = append(targets, name)
		}
	}
	return targets
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

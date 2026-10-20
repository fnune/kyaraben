// Package collections launches curated sets of games with per-game emulator
// settings. A collection manifest is published by a third party (such as a
// personal games site) and describes which ROM sets and settings each game
// needs; kyaraben resolves them against the local collection and writes
// RetroArch per-game overrides it owns.
package collections

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

const SchemaVersion = 1

type Manifest struct {
	Schema int    `json:"schema"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Games  []Game `json:"games"`
}

type Orientation string

const (
	OrientationHorizontal Orientation = "horizontal"
	OrientationVertical   Orientation = "vertical"
)

type Game struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	System      model.SystemID `json:"system"`
	Orientation Orientation    `json:"orientation,omitempty"`
	// Sets are tried in order; the first one whose emulator is enabled and
	// whose files are present wins.
	Sets     []Set    `json:"sets"`
	Settings Settings `json:"settings,omitempty"`
}

type Set struct {
	Emulator model.EmulatorID `json:"emulator"`
	Name     string           `json:"name"`
	// Requires lists parent and BIOS set names that must sit next to the ROM
	// or wherever the emulator looks for them.
	Requires    []string          `json:"requires,omitempty"`
	CoreOptions map[string]string `json:"core_options,omitempty"`
}

type Settings struct {
	RunAheadFrames int    `json:"run_ahead_frames,omitempty"`
	FrameDelay     string `json:"frame_delay,omitempty"`
	IntegerScale   *bool  `json:"integer_scale,omitempty"`
	// RetroArch holds raw retroarch.cfg keys for anything the typed settings
	// do not cover. Only display, latency and audio keys are accepted.
	RetroArch map[string]string `json:"retroarch,omitempty"`
}

func (m *Manifest) Game(id string) (Game, bool) {
	for _, g := range m.Games {
		if g.ID == id {
			return g, true
		}
	}
	return Game{}, false
}

func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing collection manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

var (
	idPattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)
	setNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	optionPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

// Manifests come from the network, so every value that ends up in a file name
// or a RetroArch config is constrained here rather than trusted downstream.
func (m *Manifest) Validate() error {
	if m.Schema != SchemaVersion {
		return fmt.Errorf("unsupported collection schema %d (want %d)", m.Schema, SchemaVersion)
	}
	if !idPattern.MatchString(m.ID) {
		return fmt.Errorf("invalid collection id %q", m.ID)
	}
	seen := make(map[string]bool, len(m.Games))
	for _, g := range m.Games {
		if err := g.validate(); err != nil {
			return fmt.Errorf("game %q: %w", g.ID, err)
		}
		if seen[g.ID] {
			return fmt.Errorf("duplicate game id %q", g.ID)
		}
		seen[g.ID] = true
	}
	return nil
}

func (g Game) validate() error {
	if !idPattern.MatchString(g.ID) {
		return fmt.Errorf("invalid id")
	}
	if g.System == "" {
		return fmt.Errorf("missing system")
	}
	switch g.Orientation {
	case "", OrientationHorizontal, OrientationVertical:
	default:
		return fmt.Errorf("invalid orientation %q", g.Orientation)
	}
	for _, s := range g.Sets {
		if !setNamePattern.MatchString(s.Name) {
			return fmt.Errorf("invalid set name %q", s.Name)
		}
		for _, r := range s.Requires {
			if !setNamePattern.MatchString(r) {
				return fmt.Errorf("set %s: invalid required set %q", s.Name, r)
			}
		}
		for k, v := range s.CoreOptions {
			if !optionPattern.MatchString(k) || strings.ContainsAny(v, "\"\n\r") {
				return fmt.Errorf("set %s: invalid core option %q", s.Name, k)
			}
		}
	}
	return g.Settings.validate()
}

func (s Settings) validate() error {
	if s.RunAheadFrames < 0 || s.RunAheadFrames > 6 {
		return fmt.Errorf("run_ahead_frames must be between 0 and 6")
	}
	if s.FrameDelay != "" && s.FrameDelay != "auto" {
		n, err := strconv.Atoi(s.FrameDelay)
		if err != nil || n < 0 || n > 15 {
			return fmt.Errorf("frame_delay must be \"auto\" or 0-15")
		}
	}
	for k, v := range s.RetroArch {
		if !allowedRetroArchKey(k) {
			return fmt.Errorf("retroarch key %q is not allowed", k)
		}
		if strings.ContainsAny(v, "\"\n\r") {
			return fmt.Errorf("retroarch key %q has an invalid value", k)
		}
	}
	return nil
}

var allowedRetroArchPrefixes = []string{
	"video_", "run_ahead_", "preemptive_frames_", "vrr_", "audio_latency",
	"aspect_ratio_index", "custom_viewport_", "input_poll_type_behavior",
}

// Paths, cores and save locations are what keep sync working, so a manifest
// may never redirect them.
var deniedRetroArchKeys = []string{
	"video_shader", "video_filter", "video_font_path", "video_layout_path",
	"video_shader_dir", "video_filter_dir", "video_layout_directory",
}

func allowedRetroArchKey(key string) bool {
	if !optionPattern.MatchString(key) {
		return false
	}
	if strings.HasSuffix(key, "_directory") || strings.HasSuffix(key, "_path") || strings.HasSuffix(key, "_dir") {
		return false
	}
	for _, denied := range deniedRetroArchKeys {
		if key == denied {
			return false
		}
	}
	for _, prefix := range allowedRetroArchPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

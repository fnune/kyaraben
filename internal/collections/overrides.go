package collections

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

// RetroArch loads <config>/<core library name>/<content name>.cfg and .opt
// on top of everything kyaraben writes globally, and never saves them back
// into retroarch.cfg. Kyaraben does not manage these per-game files
// otherwise, so collections own them outright.
type OverrideFile struct {
	Path    string
	Content []byte
}

type DeviceSettings struct {
	Vertical string
	VRR      bool
}

func GameOverrides(configDir string, game Game, res Resolution, device DeviceSettings, fs vfs.FS) ([]OverrideFile, error) {
	if !res.Playable || res.set == nil {
		return nil, nil
	}
	info := retroarch.MustGetCoreInfo(res.Emulator)
	coreDir := filepath.Join(configDir, "retroarch", "config", info.LibraryName)
	base := strings.TrimSuffix(filepath.Base(res.RomPath), filepath.Ext(res.RomPath))

	files := []OverrideFile{{
		Path:    filepath.Join(coreDir, base+".cfg"),
		Content: renderCFG(retroArchSettings(game, device)),
	}}

	if len(res.set.CoreOptions) > 0 {
		options, err := inheritedCoreOptions(fs, coreDir, info.LibraryName, game.System)
		if err != nil {
			return nil, err
		}
		for k, v := range res.set.CoreOptions {
			options[k] = v
		}
		files = append(files, OverrideFile{
			Path:    filepath.Join(coreDir, base+".opt"),
			Content: renderCFG(options),
		})
	}
	return files, nil
}

func retroArchSettings(game Game, device DeviceSettings) map[string]string {
	settings := map[string]string{}
	if game.Orientation == OrientationVertical {
		settings["video_allow_rotate"] = strconv.FormatBool(device.Vertical != model.VerticalNative)
	}
	s := game.Settings
	if s.RunAheadFrames > 0 {
		settings["run_ahead_enabled"] = "true"
		settings["run_ahead_frames"] = strconv.Itoa(s.RunAheadFrames)
	}
	switch s.FrameDelay {
	case "":
	case "auto":
		settings["video_frame_delay_auto"] = "true"
	default:
		settings["video_frame_delay_auto"] = "false"
		settings["video_frame_delay"] = s.FrameDelay
	}
	if s.IntegerScale != nil {
		settings["video_scale_integer"] = strconv.FormatBool(*s.IntegerScale)
	}
	if device.VRR {
		settings["vrr_runloop_enable"] = "true"
	}
	for k, v := range s.RetroArch {
		settings[k] = v
	}
	return settings
}

// A per-game .opt replaces the core's options wholesale rather than layering
// on them, so it starts from the options the game would otherwise get.
func inheritedCoreOptions(fs vfs.FS, coreDir, libraryName string, sys model.SystemID) (map[string]string, error) {
	for _, candidate := range []string{string(sys) + ".opt", libraryName + ".opt"} {
		data, err := fs.ReadFile(filepath.Join(coreDir, candidate))
		if isNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return parseCFG(data), nil
	}
	return map[string]string{}, nil
}

func parseCFG(data []byte) map[string]string {
	values := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"`)
	}
	return values
}

func renderCFG(values map[string]string) []byte {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b bytes.Buffer
	b.WriteString("# Per-game settings managed by kyaraben collections\n")
	b.WriteString("# Edit this file to keep your own settings; kyaraben stops updating it\n\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "%s = \"%s\"\n", k, values[k])
	}
	return b.Bytes()
}

type ownedFile struct {
	Collection string `json:"collection"`
	SHA256     string `json:"sha256"`
}

type ownedState struct {
	Files map[string]ownedFile `json:"files"`
}

type FileOutcome string

const (
	FileWritten   FileOutcome = "written"
	FileUnchanged FileOutcome = "unchanged"
	FileRemoved   FileOutcome = "removed"
	// FileForeign means the file existed before kyaraben and is left alone.
	FileForeign FileOutcome = "foreign"
	// FileModified means someone edited a file kyaraben wrote, so it is kept.
	FileModified FileOutcome = "modified"
)

type FileResult struct {
	Path    string      `json:"path"`
	Outcome FileOutcome `json:"outcome"`
}

// OverrideStore writes per-game files and remembers which ones it wrote, so
// it never overwrites a user's own overrides and can clean up after itself.
type OverrideStore struct {
	fs        vfs.FS
	statePath string
}

func NewOverrideStore(fs vfs.FS, statePath string) *OverrideStore {
	return &OverrideStore{fs: fs, statePath: statePath}
}

// Write writes files for one collection. With prune set, files the
// collection wrote earlier but no longer wants are removed. Force overwrites
// files the user edited, never files kyaraben did not create.
func (s *OverrideStore) Write(collectionID string, files []OverrideFile, prune, force bool) ([]FileResult, error) {
	state, err := s.load()
	if err != nil {
		return nil, err
	}

	var results []FileResult
	wanted := make(map[string]bool, len(files))
	for _, f := range files {
		wanted[f.Path] = true
		outcome, err := s.writeOne(state, collectionID, f, force)
		if err != nil {
			return results, err
		}
		results = append(results, FileResult{Path: f.Path, Outcome: outcome})
	}

	if prune {
		for path, owned := range state.Files {
			if owned.Collection != collectionID || wanted[path] {
				continue
			}
			outcome, err := s.removeOne(state, path, force)
			if err != nil {
				return results, err
			}
			results = append(results, FileResult{Path: path, Outcome: outcome})
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Path < results[j].Path })
	return results, s.save(state)
}

// RemoveAll deletes every unmodified file kyaraben wrote, for uninstall.
func (s *OverrideStore) RemoveAll() ([]FileResult, error) {
	state, err := s.load()
	if err != nil {
		return nil, err
	}
	var results []FileResult
	for path := range state.Files {
		outcome, err := s.removeOne(state, path, false)
		if err != nil {
			return results, err
		}
		results = append(results, FileResult{Path: path, Outcome: outcome})
	}
	return results, s.save(state)
}

func (s *OverrideStore) writeOne(state *ownedState, collectionID string, f OverrideFile, force bool) (FileOutcome, error) {
	current, err := s.fs.ReadFile(f.Path)
	exists := err == nil
	if err != nil && !isNotExist(err) {
		return "", err
	}

	owned, isOwned := state.Files[f.Path]
	if exists && (!isOwned || owned.Collection != collectionID) {
		return FileForeign, nil
	}
	if exists && hash(current) != owned.SHA256 && !force {
		return FileModified, nil
	}
	if exists && bytes.Equal(current, f.Content) {
		return FileUnchanged, nil
	}

	if err := vfs.MkdirAll(s.fs, filepath.Dir(f.Path), 0o755); err != nil {
		return "", err
	}
	if err := s.fs.WriteFile(f.Path, f.Content, 0o644); err != nil {
		return "", fmt.Errorf("writing %s: %w", f.Path, err)
	}
	state.Files[f.Path] = ownedFile{Collection: collectionID, SHA256: hash(f.Content)}
	return FileWritten, nil
}

func (s *OverrideStore) removeOne(state *ownedState, path string, force bool) (FileOutcome, error) {
	current, err := s.fs.ReadFile(path)
	if isNotExist(err) {
		delete(state.Files, path)
		return FileRemoved, nil
	}
	if err != nil {
		return "", err
	}
	if hash(current) != state.Files[path].SHA256 && !force {
		delete(state.Files, path)
		return FileModified, nil
	}
	if err := s.fs.Remove(path); err != nil {
		return "", err
	}
	delete(state.Files, path)
	return FileRemoved, nil
}

func (s *OverrideStore) load() (*ownedState, error) {
	state := &ownedState{Files: map[string]ownedFile{}}
	data, err := s.fs.ReadFile(s.statePath)
	if isNotExist(err) {
		return state, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("reading %s: %w", s.statePath, err)
	}
	if state.Files == nil {
		state.Files = map[string]ownedFile{}
	}
	return state, nil
}

func (s *OverrideStore) save(state *ownedState) error {
	if err := vfs.MkdirAll(s.fs, filepath.Dir(s.statePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return s.fs.WriteFile(s.statePath, data, 0o644)
}

func hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

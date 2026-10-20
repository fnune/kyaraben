package collections

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/emulators/retroarch"
	"github.com/fnune/kyaraben/internal/model"
)

var romExtensions = []string{".zip", ".7z"}

type Resolution struct {
	GameID   string           `json:"-"`
	Playable bool             `json:"playable"`
	Emulator model.EmulatorID `json:"emulator,omitempty"`
	Set      string           `json:"set,omitempty"`
	RomPath  string           `json:"romPath,omitempty"`
	Missing  []string         `json:"missing,omitempty"`
	Reason   string           `json:"reason,omitempty"`
	// Non-merged sets carry these inside the game's zip, so they don't block.
	MayNeed []string `json:"mayNeed,omitempty"`
	Note    string   `json:"note,omitempty"`

	set *Set
}

type Resolver struct {
	fs      vfs.FS
	store   model.StoreReader
	enabled map[model.SystemID][]model.EmulatorID
}

func NewResolver(fs vfs.FS, store model.StoreReader, enabled map[model.SystemID][]model.EmulatorID) *Resolver {
	return &Resolver{fs: fs, store: store, enabled: enabled}
}

func (r *Resolver) Resolve(game Game, local model.CollectionGameConfig) Resolution {
	res := Resolution{GameID: game.ID}

	var candidates []Set
	for _, s := range game.Sets {
		if local.Emulator != "" && s.Emulator != local.Emulator {
			continue
		}
		if !retroarch.IsRetroArchCore(s.Emulator) {
			continue
		}
		candidates = append(candidates, s)
	}
	if len(candidates) == 0 {
		if local.Emulator != "" {
			res.Reason = fmt.Sprintf("the manifest has no %s set for this game", local.Emulator)
		} else {
			res.Reason = "the manifest has no set for a supported emulator"
		}
		return res
	}

	var firstPlayable, firstMissing *Resolution
	for i := range candidates {
		s := candidates[i]
		if !slices.Contains(r.enabled[game.System], s.Emulator) {
			continue
		}
		attempt := r.check(game, s)
		switch {
		case attempt.Playable && len(attempt.MayNeed) == 0:
			return attempt
		case attempt.Playable && firstPlayable == nil:
			firstPlayable = &attempt
		case !attempt.Playable && firstMissing == nil:
			firstMissing = &attempt
		}
	}
	if firstPlayable != nil {
		return *firstPlayable
	}
	if firstMissing != nil {
		return *firstMissing
	}

	var emulators []string
	for _, s := range candidates {
		emulators = append(emulators, string(s.Emulator))
	}
	res.Reason = fmt.Sprintf("enable one of %v for %s", emulators, game.System)
	return res
}

func (r *Resolver) check(game Game, s Set) Resolution {
	res := Resolution{GameID: game.ID, Emulator: s.Emulator, Set: s.Name, set: &s}

	romsDir := r.store.SystemRomsDir(game.System)
	if rom, ok := r.find([]string{romsDir}, s.Name); ok {
		res.RomPath = rom
		res.Playable = true
	} else {
		res.Missing = []string{s.Name + ".zip"}
		res.Reason = fmt.Sprintf("%s.zip is not in %s", s.Name, romsDir)
	}

	searchDirs := r.requirementDirs(game.System, s.Emulator)
	for _, req := range s.Requires {
		if _, ok := r.find(searchDirs, req); !ok {
			res.MayNeed = append(res.MayNeed, req+".zip")
		}
	}
	if len(res.MayNeed) > 0 {
		res.Note = fmt.Sprintf("unless %s.zip is non-merged, %s needs %s in one of %s",
			s.Name, s.Emulator, strings.Join(res.MayNeed, ", "), strings.Join(searchDirs, ", "))
	}
	return res
}

func (r *Resolver) requirementDirs(sys model.SystemID, emu model.EmulatorID) []string {
	romsDir := r.store.SystemRomsDir(sys)
	biosDir := r.store.SystemBiosDir(sys)
	switch emu {
	case model.EmulatorIDRetroArchMAME: // its libretro rompath ignores the plain system directory
		return []string{romsDir, filepath.Join(biosDir, "mame", "bios"), filepath.Join(biosDir, "mame", "roms")}
	case model.EmulatorIDRetroArchFBNeo:
		return []string{romsDir, biosDir, filepath.Join(biosDir, "fbneo")}
	default:
		return []string{romsDir, biosDir}
	}
}

func (r *Resolver) find(dirs []string, set string) (string, bool) {
	for _, dir := range dirs {
		for _, ext := range romExtensions {
			p := filepath.Join(dir, set+ext)
			if info, err := r.fs.Stat(p); err == nil && !info.IsDir() {
				return p, true
			}
		}
	}
	return "", false
}

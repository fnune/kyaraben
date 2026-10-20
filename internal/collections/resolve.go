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

	var closest *Resolution
	for i := range candidates {
		s := candidates[i]
		if !slices.Contains(r.enabled[game.System], s.Emulator) {
			continue
		}
		attempt := r.check(game, s)
		if attempt.Playable {
			return attempt
		}
		if closest == nil || len(attempt.Missing) < len(closest.Missing) {
			closest = &attempt
		}
	}
	if closest != nil {
		return *closest
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

	var problems []string
	romsDir := r.store.SystemRomsDir(game.System)
	if rom, ok := r.find([]string{romsDir}, s.Name); ok {
		res.RomPath = rom
	} else {
		res.Missing = append(res.Missing, s.Name+".zip")
		problems = append(problems, fmt.Sprintf("%s.zip is not in %s", s.Name, romsDir))
	}

	searchDirs := r.requirementDirs(game.System, s.Emulator)
	var missingRequired []string
	for _, req := range s.Requires {
		if _, ok := r.find(searchDirs, req); !ok {
			missingRequired = append(missingRequired, req+".zip")
		}
	}
	if len(missingRequired) > 0 {
		res.Missing = append(res.Missing, missingRequired...)
		problems = append(problems, fmt.Sprintf("%s needs %s in one of %s",
			s.Emulator, strings.Join(missingRequired, ", "), strings.Join(searchDirs, ", ")))
	}

	res.Playable = len(res.Missing) == 0
	res.Reason = strings.Join(problems, "; ")
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

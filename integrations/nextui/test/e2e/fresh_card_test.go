package e2e

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/integrations/nextui/internal/mapping"
	"github.com/fnune/kyaraben/internal/folders"
)

var nextUISkeletonROMFolders = []string{
	"Game Boy Advance (GBA)",
	"Game Boy Color (GBC)",
	"Game Boy (GB)",
	"Nintendo Entertainment System (FC)",
	"Sega Genesis (MD)",
	"Sony PlayStation (PS)",
	"Super Nintendo Entertainment System (SFC)",
	"Amiga (PUAE)",
	"Amstrad CPC (CPC)",
	"Arcade (FBN)",
	"Atari 2600 (A2600)",
	"Atari 5200 (A5200)",
	"Atari 7800 (A7800)",
	"Atari Lynx (LYNX)",
	"Colecovision (COLECO)",
	"Commodore 128 (C128)",
	"Commodore 64 (C64)",
	"Commodore PET (PET)",
	"Commodore Plus4 (PLUS4)",
	"Commodore VIC20 (VIC)",
	"Doom (PRBOOM)",
	"Famicom Disk System (FDS)",
	"Game Boy Advance (MGBA)",
	"Microsoft MSX (MSX)",
	"Neo Geo Pocket Color (NGPC)",
	"Neo Geo Pocket (NGP)",
	"Pico-8 (P8)",
	"Pokémon mini (PKM)",
	"Sega 32X (32X)",
	"Sega CD (SEGACD)",
	"Sega Game Gear (GG)",
	"Sega Master System (SMS)",
	"Sega SG-1000 (SG1000)",
	"Super Game Boy (SGB)",
	"Super Nintendo Entertainment System (SUPA)",
	"TurboGrafx-16 (PCE)",
	"Virtual Boy (VB)",
}

func TestFirstLaunchOnAFreshNextUICardSyncsROMsIntoItsExistingFolders(t *testing.T) {
	card := map[string]any{"/data": &vfst.Dir{Perm: 0755}}
	for _, name := range nextUISkeletonROMFolders {
		card["/sdcard/Roms/"+name] = &vfst.Dir{Perm: 0755}
	}
	fs, cleanup, err := vfst.NewTestFS(card)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	cfg, err := config.NewConfigStore(fs, "/data").Load(config.DefaultConfigForCard(fs, "/sdcard"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	romFolderPrefix := folders.ID(folders.CategoryROMs, "")
	var checked int
	for _, m := range mapping.NewMapper("/sdcard", *cfg).SyncguestFolderMappings() {
		if !strings.HasPrefix(m.ID, romFolderPrefix) {
			continue
		}
		checked++
		if info, err := fs.Stat(m.Path); err != nil || !info.IsDir() {
			t.Errorf("%s syncs into %s, which a fresh NextUI card does not have", m.ID, filepath.Base(m.Path))
		}
	}
	if checked != len(cfg.ROMs) {
		t.Errorf("checked %d ROM folders, want %d", checked, len(cfg.ROMs))
	}
}

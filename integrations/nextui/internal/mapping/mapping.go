package mapping

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Category string

const (
	CategoryROMs        Category = "roms"
	CategorySaves       Category = "saves"
	CategoryBIOS        Category = "bios"
	CategoryScreenshots Category = "screenshots"
)

type SystemMapping struct {
	KyarabenID  string
	NextUITag   string
	DisplayName string
	Core        string
}

var baseSystems = []SystemMapping{
	{"nes", "FC", "Nintendo (FC)", "fceumm"},
	{"snes", "SFC", "Super Nintendo (SFC)", "snes9x"},
	{"gb", "GB", "Game Boy (GB)", "gambatte"},
	{"gbc", "GBC", "Game Boy Color (GBC)", "gambatte"},
	{"gba", "GBA", "Game Boy Advance (GBA)", "gpsp"},
	{"psx", "PS", "PlayStation (PS)", "pcsx_rearmed"},
	{"genesis", "MD", "Mega Drive (MD)", "picodrive"},
}

var extraSystems = []SystemMapping{
	{"gamegear", "GG", "Game Gear (GG)", "picodrive"},
	{"mastersystem", "SMS", "Master System (SMS)", "picodrive"},
	{"pcengine", "PCE", "PC Engine (PCE)", "mednafen_pce_fast"},
	{"ngp", "NGP", "Neo Geo Pocket (NGP)", "race"},
	{"atari2600", "A2600", "Atari 2600 (A2600)", "stella"},
	{"c64", "C64", "Commodore 64 (C64)", "vice_x64"},
	{"arcade", "FBN", "Arcade (FBN)", "fbneo"},
}

type Mapper struct {
	sdcardPath   string
	tagOverrides map[string]string
	systemByTag  map[string]SystemMapping
	systemByID   map[string]SystemMapping
}

func NewMapper(sdcardPath string, tagOverrides map[string]string) *Mapper {
	m := &Mapper{
		sdcardPath:   sdcardPath,
		tagOverrides: tagOverrides,
		systemByTag:  make(map[string]SystemMapping),
		systemByID:   make(map[string]SystemMapping),
	}

	allSystems := append(baseSystems, extraSystems...)
	for _, sys := range allSystems {
		m.systemByTag[sys.NextUITag] = sys
		m.systemByID[sys.KyarabenID] = sys
	}

	return m
}

func (m *Mapper) KyarabenFolderID(category Category, system string) string {
	if category == CategoryScreenshots {
		return "kyaraben-screenshots"
	}
	return fmt.Sprintf("kyaraben-%s-%s", category, system)
}

func (m *Mapper) NextUIPath(category Category, tag string) string {
	switch category {
	case CategoryROMs:
		sys, ok := m.systemByTag[tag]
		if !ok {
			return filepath.Join(m.sdcardPath, "Roms", tag)
		}
		return filepath.Join(m.sdcardPath, "Roms", sys.DisplayName)
	case CategorySaves:
		return filepath.Join(m.sdcardPath, "Saves", tag)
	case CategoryBIOS:
		return filepath.Join(m.sdcardPath, "Bios", tag)
	case CategoryScreenshots:
		return filepath.Join(m.sdcardPath, "Screenshots")
	default:
		return ""
	}
}

func (m *Mapper) TagToSystem(tag string) (string, bool) {
	if override, ok := m.tagOverrides[tag]; ok {
		return override, true
	}
	if sys, ok := m.systemByTag[tag]; ok {
		return sys.KyarabenID, true
	}
	return "", false
}

func (m *Mapper) SystemToTag(system string) (string, bool) {
	if sys, ok := m.systemByID[system]; ok {
		return sys.NextUITag, true
	}
	return "", false
}

func (m *Mapper) AllSystems() []SystemMapping {
	return append(baseSystems, extraSystems...)
}

func (m *Mapper) BaseSystems() []SystemMapping {
	return baseSystems
}

func (m *Mapper) SupportedTags() []string {
	var tags []string
	for tag := range m.systemByTag {
		tags = append(tags, tag)
	}
	for tag := range m.tagOverrides {
		tags = append(tags, tag)
	}
	return tags
}

func ExtractTagFromPath(path string) string {
	base := filepath.Base(path)
	if idx := strings.LastIndex(base, "("); idx != -1 {
		if end := strings.LastIndex(base, ")"); end > idx {
			return base[idx+1 : end]
		}
	}
	return base
}

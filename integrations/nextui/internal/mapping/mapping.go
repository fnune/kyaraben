package mapping

import (
	"fmt"
	"path/filepath"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
)

type Category string

const (
	CategoryROMs        Category = "roms"
	CategorySaves       Category = "saves"
	CategoryBIOS        Category = "bios"
	CategoryScreenshots Category = "screenshots"
)

type Mapper struct {
	sdcardPath string
	cfg        config.Config
}

func NewMapper(sdcardPath string, cfg config.Config) *Mapper {
	return &Mapper{
		sdcardPath: sdcardPath,
		cfg:        cfg,
	}
}

func (m *Mapper) KyarabenFolderID(category Category, system string) string {
	if category == CategoryScreenshots {
		return "kyaraben-screenshots"
	}
	return fmt.Sprintf("kyaraben-%s-%s", category, system)
}

func (m *Mapper) DevicePath(category Category, system string) string {
	var relativePath string

	switch category {
	case CategorySaves:
		relativePath = m.cfg.Saves[system]
	case CategoryROMs:
		relativePath = m.cfg.ROMs[system]
	case CategoryBIOS:
		relativePath = m.cfg.BIOS[system]
	case CategoryScreenshots:
		relativePath = m.cfg.Screenshots
	}

	if relativePath == "" {
		return ""
	}

	return filepath.Join(m.sdcardPath, relativePath)
}

func (m *Mapper) AllSystems() []string {
	seen := make(map[string]bool)
	var systems []string

	for system := range m.cfg.Saves {
		if !seen[system] {
			seen[system] = true
			systems = append(systems, system)
		}
	}
	for system := range m.cfg.ROMs {
		if !seen[system] {
			seen[system] = true
			systems = append(systems, system)
		}
	}

	return systems
}

func (m *Mapper) SystemsForCategory(category Category) []string {
	var mapping map[string]string

	switch category {
	case CategorySaves:
		mapping = m.cfg.Saves
	case CategoryROMs:
		mapping = m.cfg.ROMs
	case CategoryBIOS:
		mapping = m.cfg.BIOS
	case CategoryScreenshots:
		return nil
	}

	systems := make([]string, 0, len(mapping))
	for system := range mapping {
		systems = append(systems, system)
	}
	return systems
}

func (m *Mapper) FolderMappings() []FolderMapping {
	var mappings []FolderMapping

	for system, path := range m.cfg.Saves {
		mappings = append(mappings, FolderMapping{
			FolderID:   m.KyarabenFolderID(CategorySaves, system),
			DevicePath: filepath.Join(m.sdcardPath, path),
			Category:   CategorySaves,
			System:     system,
		})
	}

	for system, path := range m.cfg.ROMs {
		mappings = append(mappings, FolderMapping{
			FolderID:   m.KyarabenFolderID(CategoryROMs, system),
			DevicePath: filepath.Join(m.sdcardPath, path),
			Category:   CategoryROMs,
			System:     system,
		})
	}

	for system, path := range m.cfg.BIOS {
		mappings = append(mappings, FolderMapping{
			FolderID:   m.KyarabenFolderID(CategoryBIOS, system),
			DevicePath: filepath.Join(m.sdcardPath, path),
			Category:   CategoryBIOS,
			System:     system,
		})
	}

	if m.cfg.Screenshots != "" {
		mappings = append(mappings, FolderMapping{
			FolderID:   "kyaraben-screenshots",
			DevicePath: filepath.Join(m.sdcardPath, m.cfg.Screenshots),
			Category:   CategoryScreenshots,
		})
	}

	return mappings
}

type FolderMapping struct {
	FolderID   string
	DevicePath string
	Category   Category
	System     string
}

package mapping

import (
	"path/filepath"

	"github.com/fnune/kyaraben/integrations/nextui/internal/config"
	"github.com/fnune/kyaraben/internal/folders"
	"github.com/fnune/kyaraben/internal/syncguest"
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

func (m *Mapper) KyarabenFolderID(category folders.Category, system string) string {
	return folders.ID(category, system)
}

func (m *Mapper) DevicePath(category folders.Category, system string) string {
	var relativePath string

	switch category {
	case folders.CategorySaves:
		relativePath = m.cfg.Saves[system]
	case folders.CategoryROMs:
		relativePath = m.cfg.ROMs[system]
	case folders.CategoryBIOS:
		relativePath = m.cfg.BIOS[system]
	case folders.CategoryScreenshots:
		relativePath = m.cfg.Screenshots[system]
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

func (m *Mapper) SystemsForCategory(category folders.Category) []string {
	var mapping map[string]string

	switch category {
	case folders.CategorySaves:
		mapping = m.cfg.Saves
	case folders.CategoryROMs:
		mapping = m.cfg.ROMs
	case folders.CategoryBIOS:
		mapping = m.cfg.BIOS
	case folders.CategoryScreenshots:
		mapping = m.cfg.Screenshots
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
			FolderID:   folders.ID(folders.CategorySaves, system),
			DevicePath: filepath.Join(m.sdcardPath, path),
			Category:   folders.CategorySaves,
			System:     system,
		})
	}

	for system, path := range m.cfg.ROMs {
		mappings = append(mappings, FolderMapping{
			FolderID:   folders.ID(folders.CategoryROMs, system),
			DevicePath: filepath.Join(m.sdcardPath, path),
			Category:   folders.CategoryROMs,
			System:     system,
		})
	}

	for system, path := range m.cfg.BIOS {
		mappings = append(mappings, FolderMapping{
			FolderID:   folders.ID(folders.CategoryBIOS, system),
			DevicePath: filepath.Join(m.sdcardPath, path),
			Category:   folders.CategoryBIOS,
			System:     system,
		})
	}

	for emulator, path := range m.cfg.Screenshots {
		mappings = append(mappings, FolderMapping{
			FolderID:   folders.ID(folders.CategoryScreenshots, emulator),
			DevicePath: filepath.Join(m.sdcardPath, path),
			Category:   folders.CategoryScreenshots,
			System:     emulator,
		})
	}

	return mappings
}

type FolderMapping struct {
	FolderID   string
	DevicePath string
	Category   folders.Category
	System     string
}

func (m *Mapper) SyncguestFolderMappings() []syncguest.FolderMapping {
	mappings := m.FolderMappings()
	result := make([]syncguest.FolderMapping, len(mappings))
	for i, fm := range mappings {
		result[i] = syncguest.FolderMapping{
			ID:   fm.FolderID,
			Path: fm.DevicePath,
		}
	}
	return result
}

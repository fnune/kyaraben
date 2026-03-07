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

func (m *Mapper) SyncguestFolderMappings() []syncguest.FolderMapping {
	var mappings []syncguest.FolderMapping

	for system, path := range m.cfg.Saves {
		mappings = append(mappings, syncguest.FolderMapping{
			ID:   folders.ID(folders.CategorySaves, system),
			Path: filepath.Join(m.sdcardPath, path),
		})
	}

	for system, path := range m.cfg.ROMs {
		mappings = append(mappings, syncguest.FolderMapping{
			ID:   folders.ID(folders.CategoryROMs, system),
			Path: filepath.Join(m.sdcardPath, path),
		})
	}

	for system, path := range m.cfg.BIOS {
		mappings = append(mappings, syncguest.FolderMapping{
			ID:   folders.ID(folders.CategoryBIOS, system),
			Path: filepath.Join(m.sdcardPath, path),
		})
	}

	for emulator, path := range m.cfg.Screenshots {
		mappings = append(mappings, syncguest.FolderMapping{
			ID:   folders.ID(folders.CategoryScreenshots, emulator),
			Path: filepath.Join(m.sdcardPath, path),
		})
	}

	if m.cfg.Service.SyncStates {
		for emulator, path := range m.cfg.States {
			mappings = append(mappings, syncguest.FolderMapping{
				ID:   folders.ID(folders.CategoryStates, emulator),
				Path: filepath.Join(m.sdcardPath, path),
			})
		}
	}

	return mappings
}

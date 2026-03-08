package mapping

import (
	"path/filepath"

	"github.com/fnune/kyaraben/integrations/batocera/internal/config"
	"github.com/fnune/kyaraben/internal/folders"
	"github.com/fnune/kyaraben/internal/syncguest"
)

type Mapper struct {
	basePath string
	cfg      config.Config
}

func NewMapper(basePath string, cfg config.Config) *Mapper {
	return &Mapper{
		basePath: basePath,
		cfg:      cfg,
	}
}

func (m *Mapper) SyncguestFolderMappings() []syncguest.FolderMapping {
	var mappings []syncguest.FolderMapping

	for system, path := range m.cfg.Saves {
		mappings = append(mappings, syncguest.FolderMapping{
			ID:   folders.ID(folders.CategorySaves, system),
			Path: filepath.Join(m.basePath, path),
		})
	}

	for system, path := range m.cfg.ROMs {
		mappings = append(mappings, syncguest.FolderMapping{
			ID:   folders.ID(folders.CategoryROMs, system),
			Path: filepath.Join(m.basePath, path),
		})
	}

	for emulator, path := range m.cfg.Screenshots {
		mappings = append(mappings, syncguest.FolderMapping{
			ID:   folders.ID(folders.CategoryScreenshots, emulator),
			Path: filepath.Join(m.basePath, path),
		})
	}

	return mappings
}

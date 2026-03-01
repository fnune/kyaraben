package cleanup

import (
	"io/fs"
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
)

var log = logging.New("cleanup")

type Cleaner struct {
	fs       vfs.FS
	resolver model.BaseDirResolver
}

func New(fileSystem vfs.FS, resolver model.BaseDirResolver) *Cleaner {
	return &Cleaner{fs: fileSystem, resolver: resolver}
}

func NewDefault() *Cleaner {
	return &Cleaner{fs: vfs.OSFS, resolver: model.NewDefaultResolver()}
}

func (c *Cleaner) RemoveConfigDirs(configs []model.ManagedConfig) []string {
	seen := make(map[string]bool)
	var removed []string

	for _, cfg := range configs {
		dir, err := cfg.Target.ResolveDirWith(c.resolver)
		if err != nil {
			continue
		}
		if seen[dir] {
			continue
		}
		seen[dir] = true

		if !c.dirExists(dir) {
			continue
		}

		if err := c.forceRemoveAll(dir); err != nil {
			log.Info("Could not remove config directory %s: %v", dir, err)
			continue
		}
		removed = append(removed, dir)
	}

	return removed
}

func (c *Cleaner) CollectConfigDirs(configs []model.ManagedConfig) []string {
	seen := make(map[string]bool)
	var dirs []string

	for _, cfg := range configs {
		dir, err := cfg.Target.ResolveDirWith(c.resolver)
		if err != nil {
			continue
		}
		if !seen[dir] && c.dirExists(dir) {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}

	return dirs
}

func (c *Cleaner) dirExists(path string) bool {
	info, err := c.fs.Stat(path)
	return err == nil && info.IsDir()
}

func (c *Cleaner) forceRemoveAll(path string) error {
	if err := c.forceChmodRecursive(path); err != nil {
		return err
	}
	return c.fs.RemoveAll(path)
}

func (c *Cleaner) forceChmodRecursive(path string) error {
	info, err := c.fs.Lstat(path)
	if err != nil {
		return err
	}

	if info.Mode()&fs.ModeSymlink != 0 {
		return nil
	}

	if !info.IsDir() {
		return c.fs.Chmod(path, 0644)
	}

	if err := c.fs.Chmod(path, 0755); err != nil {
		return err
	}

	entries, err := c.fs.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := c.forceChmodRecursive(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

func RemoveConfigDirs(configs []model.ManagedConfig) []string {
	return NewDefault().RemoveConfigDirs(configs)
}

func CollectConfigDirs(configs []model.ManagedConfig) []string {
	return NewDefault().CollectConfigDirs(configs)
}

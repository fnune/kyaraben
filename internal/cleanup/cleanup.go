package cleanup

import (
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
)

var log = logging.New("cleanup")

func RemoveConfigDirs(configs []model.ManagedConfig) []string {
	seen := make(map[string]bool)
	var removed []string

	for _, cfg := range configs {
		dir, err := cfg.Target.ResolveDir()
		if err != nil {
			continue
		}
		if seen[dir] {
			continue
		}
		seen[dir] = true

		if !dirExists(dir) {
			continue
		}

		if err := forceRemoveAll(dir); err != nil {
			log.Info("Could not remove config directory %s: %v", dir, err)
			continue
		}
		removed = append(removed, dir)
	}

	return removed
}

func CollectConfigDirs(configs []model.ManagedConfig) []string {
	seen := make(map[string]bool)
	var dirs []string

	for _, cfg := range configs {
		dir, err := cfg.Target.ResolveDir()
		if err != nil {
			continue
		}
		if !seen[dir] && dirExists(dir) {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}

	return dirs
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func forceRemoveAll(path string) error {
	if err := forceChmodRecursive(path); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func forceChmodRecursive(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}

	if !info.IsDir() {
		return os.Chmod(path, 0644)
	}

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := forceChmodRecursive(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

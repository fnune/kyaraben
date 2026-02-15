package symlink

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
)

type Creator struct {
	fs vfs.FS
}

func NewCreator(fs vfs.FS) *Creator {
	return &Creator{fs: fs}
}

func NewDefaultCreator() *Creator {
	return &Creator{fs: vfs.OSFS}
}

func (c *Creator) Create(spec model.SymlinkSpec) error {
	return c.create(spec.Source, spec.Target)
}

func (c *Creator) create(source, target string) error {
	info, err := c.fs.Lstat(source)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existingTarget, err := c.fs.Readlink(source)
			if err != nil {
				return fmt.Errorf("reading symlink %s: %w", source, err)
			}
			if existingTarget == target {
				return nil
			}
			if err := c.fs.Remove(source); err != nil {
				return fmt.Errorf("removing old symlink %s: %w", source, err)
			}
		} else if info.IsDir() {
			entries, err := c.fs.ReadDir(source)
			if err != nil {
				return fmt.Errorf("reading directory %s: %w", source, err)
			}
			if len(entries) > 0 {
				return fmt.Errorf(
					"cannot create symlink at %s: directory contains files; "+
						"please move or remove your existing saves before running kyaraben apply",
					source,
				)
			}
			if err := c.fs.Remove(source); err != nil {
				return fmt.Errorf("removing empty directory %s: %w", source, err)
			}
		} else {
			return fmt.Errorf("cannot create symlink at %s: file exists", source)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking %s: %w", source, err)
	}

	if err := vfs.MkdirAll(c.fs, filepath.Dir(source), 0755); err != nil {
		return fmt.Errorf("creating parent directory for %s: %w", source, err)
	}

	if err := c.fs.Symlink(target, source); err != nil {
		return fmt.Errorf("creating symlink %s -> %s: %w", source, target, err)
	}

	return nil
}

func CreateAll(creator model.SymlinkCreator, specs []model.SymlinkSpec) error {
	for _, spec := range specs {
		if err := creator.Create(spec); err != nil {
			return err
		}
	}
	return nil
}

func Remove(fs vfs.FS, source string) error {
	info, err := fs.Lstat(source)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("checking %s: %w", source, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s is not a symlink, refusing to remove", source)
	}
	if err := fs.Remove(source); err != nil {
		return fmt.Errorf("removing symlink %s: %w", source, err)
	}
	return nil
}

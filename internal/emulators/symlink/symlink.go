package symlink

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/model"
)

type OSCreator struct{}

func (OSCreator) Create(spec model.SymlinkSpec) error {
	return create(spec.Source, spec.Target)
}

func create(source, target string) error {
	info, err := os.Lstat(source)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existingTarget, err := os.Readlink(source)
			if err != nil {
				return fmt.Errorf("reading symlink %s: %w", source, err)
			}
			if existingTarget == target {
				return nil
			}
			if err := os.Remove(source); err != nil {
				return fmt.Errorf("removing old symlink %s: %w", source, err)
			}
		} else if info.IsDir() {
			entries, err := os.ReadDir(source)
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
			if err := os.Remove(source); err != nil {
				return fmt.Errorf("removing empty directory %s: %w", source, err)
			}
		} else {
			return fmt.Errorf("cannot create symlink at %s: file exists", source)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking %s: %w", source, err)
	}

	if err := os.MkdirAll(filepath.Dir(source), 0755); err != nil {
		return fmt.Errorf("creating parent directory for %s: %w", source, err)
	}

	if err := os.Symlink(target, source); err != nil {
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

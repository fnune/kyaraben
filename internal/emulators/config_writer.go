package emulators

import (
	"fmt"
	"os"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/configformat"
	"github.com/fnune/kyaraben/internal/fileutil"
	"github.com/fnune/kyaraben/internal/model"
)

type ConfigWriter struct {
	fs       vfs.FS
	resolver model.BaseDirResolver
}

func NewConfigWriter(fs vfs.FS, resolver model.BaseDirResolver) *ConfigWriter {
	return &ConfigWriter{fs: fs, resolver: resolver}
}

func NewDefaultConfigWriter() *ConfigWriter {
	return NewConfigWriter(vfs.OSFS, model.NewDefaultResolver())
}

func (w *ConfigWriter) resolvePath(target model.ConfigTarget) (string, error) {
	return target.ResolveWith(w.resolver)
}

type ApplyResult struct {
	Path           string
	WrittenEntries map[string]string
	BackupPath     string
}

type ApplyOptions struct {
	CreateBackup bool
}

func (w *ConfigWriter) NeedsBackup(patch model.ConfigPatch) (string, bool, error) {
	path, err := w.resolvePath(patch.Target)
	if err != nil {
		return "", false, fmt.Errorf("resolving config path: %w", err)
	}

	_, err = w.fs.Stat(path)
	if os.IsNotExist(err) {
		return path, false, nil
	}
	if err != nil {
		return path, false, err
	}
	return path, true, nil
}

func (w *ConfigWriter) createBackup(path string) (string, error) {
	return fileutil.BackupWithTimestamp(path)
}

func (w *ConfigWriter) Apply(patch model.ConfigPatch) (ApplyResult, error) {
	return w.ApplyWithOptions(patch, ApplyOptions{})
}

func (w *ConfigWriter) ApplyWithOptions(patch model.ConfigPatch, opts ApplyOptions) (ApplyResult, error) {
	path, err := w.resolvePath(patch.Target)
	if err != nil {
		return ApplyResult{}, fmt.Errorf("resolving config path: %w", err)
	}

	var backupPath string
	if opts.CreateBackup {
		if _, err := w.fs.Stat(path); err == nil {
			backupPath, err = w.createBackup(path)
			if err != nil {
				return ApplyResult{}, fmt.Errorf("creating backup: %w", err)
			}
		}
	}

	if patch.Delete {
		_ = w.fs.Remove(path)
		return ApplyResult{Path: path, BackupPath: backupPath}, nil
	}

	if patch.ManagesWholeFile() {
		_ = w.fs.Remove(path)
	}

	handler := configformat.NewHandler(w.fs, patch.Target.Format)
	formatResult, err := handler.Apply(path, patch.Entries, patch.ManagedRegions)
	if err != nil {
		return ApplyResult{}, err
	}

	return ApplyResult{
		Path:           formatResult.Path,
		WrittenEntries: formatResult.WrittenEntries,
		BackupPath:     backupPath,
	}, nil
}

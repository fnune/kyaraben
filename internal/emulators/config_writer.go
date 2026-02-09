package emulators

import (
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/configformat"
	"github.com/fnune/kyaraben/internal/fileutil"
	"github.com/fnune/kyaraben/internal/model"
)

type ConfigWriter struct {
	resolver model.BaseDirResolver
}

func NewConfigWriter(resolver model.BaseDirResolver) *ConfigWriter {
	return &ConfigWriter{resolver: resolver}
}

func (w *ConfigWriter) resolvePath(target model.ConfigTarget) (string, error) {
	return target.ResolveWith(w.resolver)
}

type ApplyResult struct {
	Path         string
	BaselineHash string
	BackupPath   string
}

type ApplyOptions struct {
	CreateBackup bool
}

func (w *ConfigWriter) NeedsBackup(patch model.ConfigPatch) (string, bool, error) {
	path, err := w.resolvePath(patch.Target)
	if err != nil {
		return "", false, fmt.Errorf("resolving config path: %w", err)
	}

	_, err = os.Stat(path)
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
		if _, err := os.Stat(path); err == nil {
			backupPath, err = w.createBackup(path)
			if err != nil {
				return ApplyResult{}, fmt.Errorf("creating backup: %w", err)
			}
		}
	}

	handler := configformat.GetHandler(patch.Target.Format)
	formatResult, err := handler.Apply(path, patch.Entries)
	if err != nil {
		return ApplyResult{}, err
	}

	return ApplyResult{
		Path:         formatResult.Path,
		BaselineHash: formatResult.BaselineHash,
		BackupPath:   backupPath,
	}, nil
}

package fileutil

import (
	"fmt"
	"os"
	"time"
)

func BackupWithTimestamp(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file for backup: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s.bak", path, timestamp)

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("writing backup file: %w", err)
	}

	return backupPath, nil
}

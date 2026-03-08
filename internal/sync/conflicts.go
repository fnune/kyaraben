package sync

import (
	"io/fs"
	"path/filepath"
	"regexp"

	"github.com/twpayne/go-vfs/v5"
)

var conflictPattern = regexp.MustCompile(`\.sync-conflict-[A-Z0-9]{7}-\d{8}-\d{6}`)

func ScanForConflicts(fsys vfs.FS, folderPath string) ([]string, error) {
	var conflicts []string
	err := walkDir(fsys, folderPath, func(path string, d fs.DirEntry) {
		if conflictPattern.MatchString(d.Name()) {
			conflicts = append(conflicts, path)
		}
	})
	if err != nil {
		return nil, err
	}
	return conflicts, nil
}

func walkDir(fsys vfs.FS, path string, fn func(string, fs.DirEntry)) error {
	entries, err := fsys.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if err := walkDir(fsys, entryPath, fn); err != nil {
				continue
			}
		} else {
			fn(entryPath, entry)
		}
	}
	return nil
}

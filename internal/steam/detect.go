package steam

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/twpayne/go-vfs/v5"
)

type SteamInstall struct {
	RootPath string
	UserIDs  []string
}

type Detector struct {
	fs      vfs.FS
	homeDir string
}

func NewDetector(fs vfs.FS, homeDir string) *Detector {
	return &Detector{fs: fs, homeDir: homeDir}
}

func NewDefaultDetector() *Detector {
	homeDir, _ := os.UserHomeDir()
	return NewDetector(vfs.OSFS, homeDir)
}

var userIDPattern = regexp.MustCompile(`^\d+$`)

func (d *Detector) Detect() (*SteamInstall, error) {
	if d.homeDir == "" {
		return nil, nil
	}

	candidates := []string{
		filepath.Join(d.homeDir, ".steam", "steam"),
		filepath.Join(d.homeDir, ".local", "share", "Steam"),
	}

	for _, root := range candidates {
		info, err := d.fs.Stat(root)
		if err != nil || !info.IsDir() {
			continue
		}

		userDataDir := filepath.Join(root, "userdata")
		entries, err := d.fs.ReadDir(userDataDir)
		if err != nil {
			continue
		}

		var userIDs []string
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !userIDPattern.MatchString(name) {
				continue
			}
			configDir := filepath.Join(userDataDir, name, "config")
			if info, err := d.fs.Stat(configDir); err == nil && info.IsDir() {
				userIDs = append(userIDs, name)
			}
		}

		if len(userIDs) == 0 {
			continue
		}

		sort.Strings(userIDs)

		return &SteamInstall{
			RootPath: root,
			UserIDs:  userIDs,
		}, nil
	}

	return nil, nil
}

func (i *SteamInstall) ShortcutsPath(userID string) string {
	return filepath.Join(i.RootPath, "userdata", userID, "config", "shortcuts.vdf")
}

func (i *SteamInstall) GridDir(userID string) string {
	return filepath.Join(i.RootPath, "userdata", userID, "config", "grid")
}

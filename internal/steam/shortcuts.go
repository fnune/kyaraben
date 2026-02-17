package steam

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/logging"
)

var log = logging.New("steam")

type Manager struct {
	fs       vfs.FS
	detector *Detector
	install  *SteamInstall
}

func NewManager(fs vfs.FS, homeDir string) *Manager {
	detector := NewDetector(fs, homeDir)
	install, _ := detector.Detect()
	return &Manager{
		fs:       fs,
		detector: detector,
		install:  install,
	}
}

func NewDefaultManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	return NewManager(vfs.OSFS, homeDir)
}

func (m *Manager) IsAvailable() bool {
	return m.install != nil && len(m.install.UserIDs) > 0
}

func (m *Manager) Sync(entries []ShortcutEntry) error {
	if !m.IsAvailable() {
		return nil
	}

	for _, userID := range m.install.UserIDs {
		if err := m.syncForUser(userID, entries); err != nil {
			return fmt.Errorf("syncing shortcuts for user %s: %w", userID, err)
		}
	}

	return nil
}

func (m *Manager) syncForUser(userID string, entries []ShortcutEntry) error {
	shortcutsPath := m.install.ShortcutsPath(userID)

	existing, err := m.loadShortcuts(shortcutsPath)
	if err != nil {
		return fmt.Errorf("loading existing shortcuts: %w", err)
	}

	log.Debug("Loaded %d existing shortcuts", len(existing))
	for _, s := range existing {
		log.Debug("  Existing: %s (AppID %d)", s.AppName, s.AppID)
	}

	managed := make(map[uint32]bool)
	for _, entry := range entries {
		appID := GenerateAppID(entry.Exe, entry.AppName)
		managed[appID] = true
		log.Debug("Managing: %s (AppID %d)", entry.AppName, appID)
	}

	var updated []Shortcut
	for _, s := range existing {
		if managed[s.AppID] {
			log.Debug("Replacing managed entry: %s (AppID %d)", s.AppName, s.AppID)
			continue
		}
		log.Debug("Preserving non-managed entry: %s (AppID %d)", s.AppName, s.AppID)
		updated = append(updated, s)
	}

	for _, entry := range entries {
		appID := GenerateAppID(entry.Exe, entry.AppName)

		shortcut := Shortcut{
			AppID:              appID,
			AppName:            entry.AppName,
			Exe:                entry.Exe,
			StartDir:           entry.StartDir,
			Icon:               entry.Icon,
			ShortcutPath:       entry.ShortcutPath,
			LaunchOptions:      entry.LaunchOptions,
			AllowDesktopConfig: 1,
			AllowOverlay:       1,
			Tags:               entry.Tags,
		}

		updated = append(updated, shortcut)

		if entry.GridAssets != nil {
			if err := m.writeGridAssets(userID, appID, entry.GridAssets); err != nil {
				log.Debug("Failed to write grid assets for %s: %v", entry.AppName, err)
			}
		}
	}

	if err := m.saveShortcuts(shortcutsPath, updated); err != nil {
		return fmt.Errorf("saving shortcuts: %w", err)
	}

	log.Info("Synced %d shortcuts for Steam user %s", len(entries), userID)
	return nil
}

func (m *Manager) loadShortcuts(path string) ([]Shortcut, error) {
	data, err := m.fs.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return ParseShortcuts(bytes.NewReader(data))
}

func (m *Manager) saveShortcuts(path string, shortcuts []Shortcut) error {
	dir := filepath.Dir(path)
	if err := vfs.MkdirAll(m.fs, dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	var buf bytes.Buffer
	if err := WriteShortcuts(&buf, shortcuts); err != nil {
		return fmt.Errorf("encoding shortcuts: %w", err)
	}

	tempPath := path + ".tmp"
	if err := m.fs.WriteFile(tempPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := m.fs.Rename(tempPath, path); err != nil {
		_ = m.fs.Remove(tempPath)
		return fmt.Errorf("renaming file: %w", err)
	}

	return nil
}

func (m *Manager) writeGridAssets(userID string, appID uint32, assets *GridAssets) error {
	gridDir := m.install.GridDir(userID)
	if err := vfs.MkdirAll(m.fs, gridDir, 0755); err != nil {
		return fmt.Errorf("creating grid directory: %w", err)
	}

	type asset struct {
		data   []byte
		suffix string
	}
	assetList := []asset{
		{assets.Grid, ".png"},
		{assets.Hero, "_hero.png"},
		{assets.Logo, "_logo.png"},
		{assets.Capsule, "p.png"},
	}

	for _, a := range assetList {
		if len(a.data) == 0 {
			continue
		}
		filename := fmt.Sprintf("%d%s", appID, a.suffix)
		path := filepath.Join(gridDir, filename)
		if err := m.fs.WriteFile(path, a.data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}
	}

	return nil
}

func (m *Manager) RemoveShortcuts(appIDs []uint32) error {
	if !m.IsAvailable() {
		return nil
	}

	appIDSet := make(map[uint32]bool)
	for _, id := range appIDs {
		appIDSet[id] = true
	}

	for _, userID := range m.install.UserIDs {
		shortcutsPath := m.install.ShortcutsPath(userID)

		existing, err := m.loadShortcuts(shortcutsPath)
		if err != nil {
			continue
		}

		var updated []Shortcut
		var removed []uint32
		for _, s := range existing {
			if appIDSet[s.AppID] {
				removed = append(removed, s.AppID)
				continue
			}
			updated = append(updated, s)
		}

		if len(removed) == 0 {
			continue
		}

		if err := m.saveShortcuts(shortcutsPath, updated); err != nil {
			return fmt.Errorf("saving shortcuts for user %s: %w", userID, err)
		}

		gridDir := m.install.GridDir(userID)
		for _, appID := range removed {
			m.removeGridAssets(gridDir, appID)
		}

		log.Info("Removed %d shortcuts for Steam user %s", len(removed), userID)
	}

	return nil
}

func (m *Manager) removeGridAssets(gridDir string, appID uint32) {
	suffixes := []string{".png", "_hero.png", "_logo.png", "p.png"}
	for _, suffix := range suffixes {
		path := filepath.Join(gridDir, fmt.Sprintf("%d%s", appID, suffix))
		_ = m.fs.Remove(path)
	}
}

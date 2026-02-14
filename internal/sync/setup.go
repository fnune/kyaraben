package sync

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/packages"
)

type Setup struct {
	fs        vfs.FS
	installer packages.Installer
	stateDir  string
}

func NewSetup(fs vfs.FS, installer packages.Installer, stateDir string) *Setup {
	return &Setup{
		fs:        fs,
		installer: installer,
		stateDir:  stateDir,
	}
}

func NewDefaultSetup(installer packages.Installer, stateDir string) *Setup {
	return NewSetup(vfs.OSFS, installer, stateDir)
}

type SetupResult struct {
	SyncthingBinary string
	ConfigDir       string
	DataDir         string
	APIKey          string
}

func (s *Setup) Install(ctx context.Context, cfg model.SyncConfig, userStorePath string, allSystems []model.SystemID, onProgress func(packages.InstallProgress)) (*SetupResult, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	binary, err := s.installer.InstallEmulator(ctx, "syncthing", onProgress)
	if err != nil {
		return nil, fmt.Errorf("installing syncthing: %w", err)
	}

	configDir := filepath.Join(s.stateDir, "syncthing", "config")
	dataDir := filepath.Join(s.stateDir, "syncthing", "data")

	apiKey, err := s.loadOrGenerateAPIKey(configDir)
	if err != nil {
		return nil, fmt.Errorf("generating API key: %w", err)
	}

	configGen := NewConfigGenerator(s.fs, cfg, userStorePath, allSystems)
	configGen.SetAPIKey(apiKey)

	if err := configGen.WriteConfig(configDir); err != nil {
		return nil, fmt.Errorf("writing syncthing config: %w", err)
	}

	unitGen := NewSystemdUnit(s.fs)
	params := UnitParams{
		BinaryPath: binary.Path,
		ConfigDir:  configDir,
		DataDir:    dataDir,
		GUIPort:    cfg.Syncthing.GUIPort,
		APIKey:     apiKey,
	}

	if err := unitGen.Write(params); err != nil {
		return nil, fmt.Errorf("writing systemd unit: %w", err)
	}

	if err := unitGen.Enable(); err != nil {
		return nil, fmt.Errorf("enabling syncthing service: %w", err)
	}

	return &SetupResult{
		SyncthingBinary: binary.Path,
		ConfigDir:       configDir,
		DataDir:         dataDir,
		APIKey:          apiKey,
	}, nil
}

func (s *Setup) Disable() error {
	unitGen := NewSystemdUnit(s.fs)
	return unitGen.Disable()
}

func (s *Setup) IsEnabled() bool {
	unitGen := NewSystemdUnit(s.fs)
	return unitGen.IsEnabled()
}

func (s *Setup) loadOrGenerateAPIKey(configDir string) (string, error) {
	keyPath := filepath.Join(configDir, ".apikey")

	data, err := s.fs.ReadFile(keyPath)
	if err == nil && len(data) >= 32 {
		return string(data), nil
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return "", err
	}

	if err := vfs.MkdirAll(s.fs, configDir, 0700); err != nil {
		return "", fmt.Errorf("creating config dir: %w", err)
	}

	if err := s.fs.WriteFile(keyPath, []byte(apiKey), 0600); err != nil {
		return "", fmt.Errorf("saving API key: %w", err)
	}

	return apiKey, nil
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

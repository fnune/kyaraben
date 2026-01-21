package paths

import (
	"os"
	"path/filepath"
)

// DataDir returns XDG_DATA_HOME or ~/.local/share
func DataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

// StateDir returns XDG_STATE_HOME or ~/.local/state
func StateDir() (string, error) {
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state"), nil
}

// ConfigDir returns XDG_CONFIG_HOME or ~/.config
func ConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	return os.UserConfigDir()
}

// KyarabenDataDir returns the kyaraben data directory
func KyarabenDataDir() (string, error) {
	base, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "kyaraben"), nil
}

// KyarabenStateDir returns the kyaraben state directory
func KyarabenStateDir() (string, error) {
	base, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "kyaraben"), nil
}

// KyarabenConfigDir returns the kyaraben config directory
func KyarabenConfigDir() (string, error) {
	base, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "kyaraben"), nil
}

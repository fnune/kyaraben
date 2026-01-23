package model

import (
	"fmt"
	"os"
	"path/filepath"
)

type ConfigFormat string

const (
	ConfigFormatINI  ConfigFormat = "ini"
	ConfigFormatTOML ConfigFormat = "toml"
	ConfigFormatCFG  ConfigFormat = "cfg"
	ConfigFormatXML  ConfigFormat = "xml"
	ConfigFormatJSON ConfigFormat = "json"
)

type ConfigBaseDir string

const (
	ConfigBaseDirUserConfig ConfigBaseDir = "user_config"
	ConfigBaseDirUserData   ConfigBaseDir = "user_data"
	ConfigBaseDirHome       ConfigBaseDir = "home"
)

type ConfigTarget struct {
	RelPath string
	Format  ConfigFormat
	BaseDir ConfigBaseDir
}

func (ct ConfigTarget) Resolve() (string, error) {
	var baseDir string

	switch ct.BaseDir {
	case ConfigBaseDirUserConfig:
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("getting user config dir: %w", err)
		}
		baseDir = dir

	case ConfigBaseDirUserData:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home dir: %w", err)
		}
		baseDir = filepath.Join(home, ".local", "share")

	case ConfigBaseDirHome:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home dir: %w", err)
		}
		baseDir = home

	default:
		return "", fmt.Errorf("unknown config base dir: %s", ct.BaseDir)
	}

	return filepath.Join(baseDir, ct.RelPath), nil
}

type ConfigEntry struct {
	Section string
	Key     string
	Value   string
}

type ConfigPatch struct {
	Target  ConfigTarget
	Entries []ConfigEntry
}

package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ConfigFormat string

const (
	ConfigFormatINI  ConfigFormat = "ini"
	ConfigFormatTOML ConfigFormat = "toml"
	ConfigFormatCFG  ConfigFormat = "cfg"
	ConfigFormatXML  ConfigFormat = "xml"
	ConfigFormatJSON ConfigFormat = "json"
	ConfigFormatYAML ConfigFormat = "yaml"
	ConfigFormatRaw  ConfigFormat = "raw"
)

type ConfigBaseDir string

const (
	ConfigBaseDirUserConfig ConfigBaseDir = "user_config"
	ConfigBaseDirUserData   ConfigBaseDir = "user_data"
	ConfigBaseDirHome       ConfigBaseDir = "home"
	ConfigBaseDirOpaqueDir  ConfigBaseDir = "opaque_dir"
)

type ConfigTarget struct {
	RelPath string
	Format  ConfigFormat
	BaseDir ConfigBaseDir
}

type BaseDirResolver interface {
	UserConfigDir() (string, error)
	UserHomeDir() (string, error)
	UserDataDir() (string, error)
}

type OSBaseDirResolver struct{}

func (OSBaseDirResolver) UserConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	return os.UserConfigDir()
}

func (OSBaseDirResolver) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (OSBaseDirResolver) UserDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

func (ct ConfigTarget) ResolveWith(resolver BaseDirResolver) (string, error) {
	var baseDir string

	switch ct.BaseDir {
	case ConfigBaseDirOpaqueDir:
		return ct.RelPath, nil

	case ConfigBaseDirUserConfig:
		dir, err := resolver.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("getting user config dir: %w", err)
		}
		baseDir = dir

	case ConfigBaseDirUserData:
		home, err := resolver.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home dir: %w", err)
		}
		baseDir = filepath.Join(home, ".local", "share")

	case ConfigBaseDirHome:
		home, err := resolver.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home dir: %w", err)
		}
		baseDir = home

	default:
		return "", fmt.Errorf("unknown config base dir: %s", ct.BaseDir)
	}

	return filepath.Join(baseDir, ct.RelPath), nil
}

func (ct ConfigTarget) Resolve() (string, error) {
	return ct.ResolveWith(OSBaseDirResolver{})
}

type ConfigEntry struct {
	Path      []string
	Value     string
	Unmanaged bool // Only set if key doesn't exist; user changes are preserved
}

func (e ConfigEntry) Key() string {
	if len(e.Path) == 0 {
		return ""
	}
	return e.Path[len(e.Path)-1]
}

func (e ConfigEntry) Parent() []string {
	if len(e.Path) <= 1 {
		return nil
	}
	return e.Path[:len(e.Path)-1]
}

func (e ConfigEntry) FullPath() string {
	return strings.Join(e.Path, ".")
}

type ConfigPatch struct {
	Target  ConfigTarget
	Entries []ConfigEntry
}

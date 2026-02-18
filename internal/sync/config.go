package sync

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
)

type FolderType string

const (
	FolderTypeSendReceive FolderType = "sendreceive"
	FolderTypeSendOnly    FolderType = "sendonly"
	FolderTypeReceiveOnly FolderType = "receiveonly"
)

type SyncthingXMLConfig struct {
	XMLName  xml.Name    `xml:"configuration"`
	Version  int         `xml:"version,attr"`
	Folders  []XMLFolder `xml:"folder"`
	Devices  []XMLDevice `xml:"device"`
	GUI      XMLGUI      `xml:"gui"`
	Options  XMLOptions  `xml:"options"`
	Defaults XMLDefaults `xml:"defaults"`
}

type XMLFolder struct {
	ID               string            `xml:"id,attr"`
	Label            string            `xml:"label,attr"`
	Path             string            `xml:"path,attr"`
	Type             FolderType        `xml:"type,attr"`
	Devices          []XMLFolderDevice `xml:"device"`
	FSWatcherEnabled bool              `xml:"fsWatcherEnabled"`
	IgnorePerms      bool              `xml:"ignorePerms"`
	Versioning       XMLVersioning     `xml:"versioning"`
}

type XMLFolderDevice struct {
	ID string `xml:"id,attr"`
}

type XMLDevice struct {
	ID                string   `xml:"id,attr"`
	Name              string   `xml:"name,attr"`
	Compression       string   `xml:"compression,attr"`
	Introducer        bool     `xml:"introducer,attr"`
	AutoAcceptFolders bool     `xml:"autoAcceptFolders,attr"`
	Addresses         []string `xml:"address,omitempty"`
}

type XMLGUI struct {
	Enabled bool   `xml:"enabled,attr"`
	Address string `xml:"address"`
	APIKey  string `xml:"apikey"`
	Theme   string `xml:"theme"`
}

type XMLOptions struct {
	ListenAddresses       []string `xml:"listenAddress"`
	GlobalAnnounceEnabled bool     `xml:"globalAnnounceEnabled"`
	LocalAnnounceEnabled  bool     `xml:"localAnnounceEnabled"`
	LocalAnnouncePort     int      `xml:"localAnnouncePort"`
	RelaysEnabled         bool     `xml:"relaysEnabled"`
	URAccepted            int      `xml:"urAccepted"`
	AutoUpgradeIntervalH  int      `xml:"autoUpgradeIntervalH"`
}

type XMLDefaults struct {
	Folder XMLDefaultFolder `xml:"folder"`
}

type XMLDefaultFolder struct {
	Path string `xml:"path,attr"`
}

type XMLVersioning struct {
	Type   string               `xml:"type,attr"`
	Params []XMLVersioningParam `xml:"param"`
}

type XMLVersioningParam struct {
	Key string `xml:"key,attr"`
	Val string `xml:"val,attr"`
}

type ConfigGenerator struct {
	fs         vfs.FS
	syncConfig model.SyncConfig
	userStore  string
	deviceID   string
	apiKey     string
	allSystems []model.SystemID
}

func NewConfigGenerator(fs vfs.FS, syncConfig model.SyncConfig, userStore string, allSystems []model.SystemID) *ConfigGenerator {
	return &ConfigGenerator{
		fs:         fs,
		syncConfig: syncConfig,
		userStore:  userStore,
		allSystems: allSystems,
	}
}

func NewDefaultConfigGenerator(syncConfig model.SyncConfig, userStore string, allSystems []model.SystemID) *ConfigGenerator {
	return NewConfigGenerator(vfs.OSFS, syncConfig, userStore, allSystems)
}

func (g *ConfigGenerator) SetDeviceID(id string) {
	g.deviceID = id
}

func (g *ConfigGenerator) SetAPIKey(key string) {
	g.apiKey = key
}

func (g *ConfigGenerator) Generate() (*SyncthingXMLConfig, error) {
	folders, err := g.generateFolders()
	if err != nil {
		return nil, fmt.Errorf("generating folders: %w", err)
	}

	devices := g.generateDevices()

	config := &SyncthingXMLConfig{
		Version: 37,
		Folders: folders,
		Devices: devices,
		GUI: XMLGUI{
			Enabled: true,
			Address: fmt.Sprintf("127.0.0.1:%d", g.syncConfig.Syncthing.GUIPort),
			APIKey:  g.apiKey,
			Theme:   "default",
		},
		Options: XMLOptions{
			ListenAddresses: []string{
				fmt.Sprintf("tcp://0.0.0.0:%d", g.syncConfig.Syncthing.ListenPort),
				fmt.Sprintf("quic://0.0.0.0:%d", g.syncConfig.Syncthing.ListenPort),
			},
			GlobalAnnounceEnabled: true,
			LocalAnnounceEnabled:  true,
			LocalAnnouncePort:     g.syncConfig.Syncthing.DiscoveryPort,
			RelaysEnabled:         g.syncConfig.Syncthing.RelayEnabled,
			URAccepted:            -1,
			AutoUpgradeIntervalH:  0,
		},
		Defaults: XMLDefaults{
			Folder: XMLDefaultFolder{
				Path: g.userStore,
			},
		},
	}

	return config, nil
}

func (g *ConfigGenerator) generateFolders() ([]XMLFolder, error) {
	var folders []XMLFolder

	isPrimary := g.syncConfig.Mode == model.SyncModePrimary

	folderTypes := map[string]struct {
		subdirs       []string
		primaryType   FolderType
		secondaryType FolderType
		versioning    bool
	}{
		"roms":        {subdirs: g.systemSubdirs(), primaryType: FolderTypeSendOnly, secondaryType: FolderTypeReceiveOnly, versioning: false},
		"bios":        {subdirs: g.systemSubdirs(), primaryType: FolderTypeSendOnly, secondaryType: FolderTypeReceiveOnly, versioning: false},
		"saves":       {subdirs: g.systemSubdirs(), primaryType: FolderTypeSendReceive, secondaryType: FolderTypeSendReceive, versioning: true},
		"states":      {subdirs: g.systemSubdirs(), primaryType: FolderTypeSendReceive, secondaryType: FolderTypeSendReceive, versioning: true},
		"screenshots": {subdirs: nil, primaryType: FolderTypeSendReceive, secondaryType: FolderTypeSendReceive, versioning: false},
	}

	deviceRefs := g.folderDeviceRefs()

	for category, spec := range folderTypes {
		folderType := spec.primaryType
		if !isPrimary {
			folderType = spec.secondaryType
		}

		if spec.subdirs != nil {
			for _, subdir := range spec.subdirs {
				folderID := fmt.Sprintf("kyaraben-%s-%s", category, subdir)
				path := filepath.Join(g.userStore, category, subdir)

				folder := XMLFolder{
					ID:               folderID,
					Label:            folderID,
					Path:             path,
					Type:             folderType,
					Devices:          deviceRefs,
					FSWatcherEnabled: true,
					IgnorePerms:      true,
				}

				if spec.versioning {
					folder.Versioning = g.versioningConfig()
				}

				folders = append(folders, folder)
			}
		} else {
			folderID := fmt.Sprintf("kyaraben-%s", category)
			path := filepath.Join(g.userStore, category)

			folder := XMLFolder{
				ID:               folderID,
				Label:            folderID,
				Path:             path,
				Type:             folderType,
				Devices:          deviceRefs,
				FSWatcherEnabled: true,
				IgnorePerms:      true,
			}

			if spec.versioning {
				folder.Versioning = g.versioningConfig()
			}

			folders = append(folders, folder)
		}
	}

	return folders, nil
}

func (g *ConfigGenerator) systemSubdirs() []string {
	subdirs := make([]string, len(g.allSystems))
	for i, sys := range g.allSystems {
		subdirs[i] = string(sys)
	}
	return subdirs
}

func (g *ConfigGenerator) generateDevices() []XMLDevice {
	var devices []XMLDevice

	if g.deviceID != "" {
		devices = append(devices, XMLDevice{
			ID:          g.deviceID,
			Name:        "this-device",
			Compression: "metadata",
		})
	}

	return devices
}

func (g *ConfigGenerator) folderDeviceRefs() []XMLFolderDevice {
	var refs []XMLFolderDevice

	if g.deviceID != "" {
		refs = append(refs, XMLFolderDevice{ID: g.deviceID})
	}

	return refs
}

func (g *ConfigGenerator) versioningConfig() XMLVersioning {
	return XMLVersioning{
		Type: "staggered",
		Params: []XMLVersioningParam{
			{Key: "cleanInterval", Val: "3600"},
			{Key: "maxAge", Val: "2592000"},
		},
	}
}

func (g *ConfigGenerator) WriteConfig(configDir string) error {
	config, err := g.Generate()
	if err != nil {
		return err
	}

	if err := vfs.MkdirAll(g.fs, configDir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.xml")

	data, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	xmlHeader := []byte(xml.Header)
	data = append(xmlHeader, data...)

	if err := g.fs.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	log.Info("Wrote syncthing config to %s", configPath)

	if err := g.writeIgnoreFiles(config.Folders); err != nil {
		return fmt.Errorf("writing ignore files: %w", err)
	}

	return nil
}

func (g *ConfigGenerator) writeIgnoreFiles(folders []XMLFolder) error {
	if len(g.syncConfig.Ignore.Patterns) == 0 {
		return nil
	}

	var content strings.Builder
	for _, pattern := range g.syncConfig.Ignore.Patterns {
		content.WriteString(pattern)
		content.WriteString("\n")
	}
	ignoreContent := []byte(content.String())

	for _, folder := range folders {
		if err := vfs.MkdirAll(g.fs, folder.Path, 0755); err != nil {
			return fmt.Errorf("creating folder %s: %w", folder.Path, err)
		}

		ignorePath := filepath.Join(folder.Path, ".stignore")
		if err := g.fs.WriteFile(ignorePath, ignoreContent, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", ignorePath, err)
		}
	}

	log.Info("Wrote .stignore files to %d folders", len(folders))
	return nil
}

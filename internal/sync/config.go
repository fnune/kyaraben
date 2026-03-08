package sync

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/folders"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/syncthing"
)

type FolderType string

const (
	FolderTypeSendReceive FolderType = "sendreceive"
)

const (
	ConfigSchemaVersion = 1
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
	NATEnabled            bool     `xml:"natEnabled"`
	RelaysEnabled         bool     `xml:"relaysEnabled"`
	CrashReportingEnabled bool     `xml:"crashReportingEnabled"`
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
	fs           vfs.FS
	syncConfig   model.SyncConfig
	collection   string
	deviceID     string
	apiKey       string
	allSystems   []model.SystemID
	allEmulators []folders.EmulatorInfo
	allFrontends []model.FrontendID
}

func NewConfigGenerator(fs vfs.FS, syncConfig model.SyncConfig, collection string, allSystems []model.SystemID, allEmulators []folders.EmulatorInfo, allFrontends []model.FrontendID) *ConfigGenerator {
	return &ConfigGenerator{
		fs:           fs,
		syncConfig:   syncConfig,
		collection:   collection,
		allSystems:   allSystems,
		allEmulators: allEmulators,
		allFrontends: allFrontends,
	}
}

func NewDefaultConfigGenerator(syncConfig model.SyncConfig, collection string, allSystems []model.SystemID, allEmulators []folders.EmulatorInfo, allFrontends []model.FrontendID) *ConfigGenerator {
	return NewConfigGenerator(vfs.OSFS, syncConfig, collection, allSystems, allEmulators, allFrontends)
}

func (g *ConfigGenerator) SetDeviceID(id string) {
	g.deviceID = id
}

func (g *ConfigGenerator) SetAPIKey(key string) {
	g.apiKey = key
}

func (g *ConfigGenerator) GenerateBootstrap() *SyncthingXMLConfig {
	return &SyncthingXMLConfig{
		Version: 37,
		Folders: nil,
		Devices: nil,
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
			GlobalAnnounceEnabled: g.syncConfig.Syncthing.GlobalDiscoveryEnabled,
			LocalAnnounceEnabled:  true,
			LocalAnnouncePort:     g.syncConfig.Syncthing.DiscoveryPort,
			RelaysEnabled:         g.syncConfig.Syncthing.RelayEnabled,
			URAccepted:            -1,
			AutoUpgradeIntervalH:  0,
		},
		Defaults: XMLDefaults{
			Folder: XMLDefaultFolder{
				Path: g.collection,
			},
		},
	}
}

func (g *ConfigGenerator) FolderCreateRequests() []syncthing.FolderCreateRequest {
	specs := folders.GenerateSpecs(folders.HostInput{
		Systems:          g.allSystems,
		Emulators:        g.allEmulators,
		Frontends:        g.allFrontends,
		FrontendSuffixes: g.frontendSuffixes,
	})

	requests := make([]syncthing.FolderCreateRequest, len(specs))
	for i, spec := range specs {
		requests[i] = syncthing.FolderCreateRequest{
			ID:    spec.ID,
			Label: spec.ID,
			Path:  g.folderPath(spec),
			Type:  string(FolderTypeSendReceive),
		}
	}
	return requests
}

func (g *ConfigGenerator) folderPath(spec folders.Spec) string {
	switch spec.Category {
	case folders.CategoryROMs:
		return filepath.Join(g.collection, "roms", string(spec.System))
	case folders.CategoryBIOS:
		return filepath.Join(g.collection, "bios", string(spec.System))
	case folders.CategorySaves:
		return filepath.Join(g.collection, "saves", string(spec.System))
	case folders.CategoryStates:
		return filepath.Join(g.collection, "states", string(spec.Emulator))
	case folders.CategoryScreenshots:
		return filepath.Join(g.collection, "screenshots", string(spec.Emulator))
	default:
		if spec.Frontend != "" {
			return g.frontendPath(spec)
		}
		return g.collection
	}
}

func (g *ConfigGenerator) frontendPath(spec folders.Spec) string {
	suffix := strings.TrimPrefix(spec.ID, fmt.Sprintf("kyaraben-frontends-%s-", spec.Frontend))
	parts := strings.SplitN(suffix, "-", 2)
	if len(parts) == 2 {
		return filepath.Join(g.collection, "frontends", string(spec.Frontend), parts[0], parts[1])
	}
	return filepath.Join(g.collection, "frontends", string(spec.Frontend), suffix)
}

func (g *ConfigGenerator) frontendSuffixes(fe model.FrontendID, systems []model.SystemID) []string {
	var suffixes []string
	for _, subType := range []string{"gamelists", "media"} {
		for _, sys := range systems {
			suffixes = append(suffixes, fmt.Sprintf("%s-%s", subType, sys))
		}
	}
	return suffixes
}

func (g *ConfigGenerator) WriteConfig(configDir string) error {
	if err := vfs.MkdirAll(g.fs, configDir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.xml")

	existingData, err := g.fs.ReadFile(configPath)
	if err != nil {
		return g.writeBootstrapConfig(configPath)
	}

	var existing SyncthingXMLConfig
	if err := xml.Unmarshal(existingData, &existing); err != nil {
		log.Info("Could not parse existing config.xml, writing fresh: %v", err)
		return g.writeBootstrapConfig(configPath)
	}

	return g.writeMergedConfig(configPath, &existing)
}

func (g *ConfigGenerator) writeBootstrapConfig(configPath string) error {
	config := g.GenerateBootstrap()
	return g.writeConfig(configPath, config, "bootstrap")
}

func (g *ConfigGenerator) writeMergedConfig(configPath string, existing *SyncthingXMLConfig) error {
	bootstrap := g.GenerateBootstrap()

	merged := &SyncthingXMLConfig{
		Version:  existing.Version,
		Folders:  existing.Folders,
		Devices:  existing.Devices,
		GUI:      bootstrap.GUI,
		Options:  bootstrap.Options,
		Defaults: bootstrap.Defaults,
	}

	return g.writeConfig(configPath, merged, "merged")
}

func (g *ConfigGenerator) writeConfig(configPath string, config *SyncthingXMLConfig, configType string) error {
	data, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	xmlHeader := []byte(xml.Header)
	data = append(xmlHeader, data...)

	if err := g.fs.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	log.Info("Wrote syncthing %s config to %s", configType, configPath)
	return nil
}

func (g *ConfigGenerator) WriteIgnoreFiles(folderRequests []syncthing.FolderCreateRequest) error {
	if len(g.syncConfig.Ignore.Patterns) == 0 {
		return nil
	}

	var content strings.Builder
	for _, pattern := range g.syncConfig.Ignore.Patterns {
		content.WriteString(pattern)
		content.WriteString("\n")
	}
	ignoreContent := []byte(content.String())

	for _, folder := range folderRequests {
		if err := vfs.MkdirAll(g.fs, folder.Path, 0755); err != nil {
			return fmt.Errorf("creating folder %s: %w", folder.Path, err)
		}

		ignorePath := filepath.Join(folder.Path, ".stignore")
		if err := g.fs.WriteFile(ignorePath, ignoreContent, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", ignorePath, err)
		}
	}

	log.Info("Wrote .stignore files to %d folders", len(folderRequests))
	return nil
}

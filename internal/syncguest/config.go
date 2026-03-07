package syncguest

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

type FolderMapping struct {
	ID   string
	Path string
}

type XMLConfig struct {
	XMLName xml.Name    `xml:"configuration"`
	Version int         `xml:"version,attr"`
	Folders []XMLFolder `xml:"folder"`
	Devices []XMLDevice `xml:"device"`
	GUI     XMLGUI      `xml:"gui"`
	Options XMLOptions  `xml:"options"`
}

type XMLFolder struct {
	ID               string            `xml:"id,attr"`
	Label            string            `xml:"label,attr"`
	Path             string            `xml:"path,attr"`
	Type             string            `xml:"type,attr"`
	Devices          []XMLFolderDevice `xml:"device"`
	FSWatcherEnabled bool              `xml:"fsWatcherEnabled"`
	IgnorePerms      bool              `xml:"ignorePerms"`
}

type XMLFolderDevice struct {
	ID string `xml:"id,attr"`
}

type XMLDevice struct {
	ID          string `xml:"id,attr"`
	Name        string `xml:"name,attr"`
	Compression string `xml:"compression,attr"`
}

type XMLGUI struct {
	Enabled bool   `xml:"enabled,attr"`
	Address string `xml:"address"`
	APIKey  string `xml:"apikey"`
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

func (m *Manager) ConfigureFolders(folders []FolderMapping) error {
	configDir := filepath.Join(m.config.DataDir, "syncthing")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	deviceID, apiKey, err := m.loadOrCreateIdentity(configDir)
	if err != nil {
		return fmt.Errorf("load identity: %w", err)
	}

	existingDevices, err := m.loadExistingDevices(configDir, deviceID)
	if err != nil {
		m.logger.Debug("could not load existing devices: %v", err)
	}

	config := m.generateConfig(folders, deviceID, apiKey, existingDevices)

	configPath := filepath.Join(configDir, "config.xml")
	data, err := xml.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	data = append([]byte(xml.Header), data...)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (m *Manager) loadOrCreateIdentity(configDir string) (deviceID, apiKey string, err error) {
	configPath := filepath.Join(configDir, "config.xml")
	data, err := os.ReadFile(configPath)
	if err == nil {
		var existing XMLConfig
		if xml.Unmarshal(data, &existing) == nil {
			for _, dev := range existing.Devices {
				if dev.Name == "this-device" {
					deviceID = dev.ID
				}
			}
			apiKey = existing.GUI.APIKey
		}
	}

	if deviceID == "" {
		deviceID = "PENDING"
	}
	if apiKey == "" {
		apiKey = generateAPIKey()
	}

	return deviceID, apiKey, nil
}

func (m *Manager) loadExistingDevices(configDir, selfID string) ([]XMLDevice, error) {
	configPath := filepath.Join(configDir, "config.xml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var existing XMLConfig
	if err := xml.Unmarshal(data, &existing); err != nil {
		return nil, err
	}

	var devices []XMLDevice
	for _, dev := range existing.Devices {
		if dev.ID != selfID && dev.Name != "this-device" {
			devices = append(devices, dev)
		}
	}

	return devices, nil
}

func (m *Manager) generateConfig(folders []FolderMapping, deviceID, apiKey string, existingDevices []XMLDevice) *XMLConfig {
	var deviceRefs []XMLFolderDevice
	deviceRefs = append(deviceRefs, XMLFolderDevice{ID: deviceID})
	for _, dev := range existingDevices {
		deviceRefs = append(deviceRefs, XMLFolderDevice{ID: dev.ID})
	}

	var xmlFolders []XMLFolder
	for _, f := range folders {
		xmlFolders = append(xmlFolders, XMLFolder{
			ID:               f.ID,
			Label:            f.ID,
			Path:             f.Path,
			Type:             "sendreceive",
			Devices:          deviceRefs,
			FSWatcherEnabled: true,
			IgnorePerms:      true,
		})
	}

	var devices []XMLDevice
	devices = append(devices, XMLDevice{
		ID:          deviceID,
		Name:        "this-device",
		Compression: "metadata",
	})
	devices = append(devices, existingDevices...)

	stConfig := m.config.Syncthing
	return &XMLConfig{
		Version: 37,
		Folders: xmlFolders,
		Devices: devices,
		GUI: XMLGUI{
			Enabled: true,
			Address: fmt.Sprintf("127.0.0.1:%d", stConfig.GUIPort),
			APIKey:  apiKey,
		},
		Options: XMLOptions{
			ListenAddresses: []string{
				fmt.Sprintf("tcp://0.0.0.0:%d", stConfig.ListenPort),
				fmt.Sprintf("quic://0.0.0.0:%d", stConfig.ListenPort),
			},
			GlobalAnnounceEnabled: stConfig.GlobalDiscoveryEnabled,
			LocalAnnounceEnabled:  true,
			LocalAnnouncePort:     stConfig.DiscoveryPort,
			RelaysEnabled:         stConfig.RelayEnabled,
			URAccepted:            -1,
			AutoUpgradeIntervalH:  0,
		},
	}
}

func generateAPIKey() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[i%len(chars)]
	}
	return string(b)
}

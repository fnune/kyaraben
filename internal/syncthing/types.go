package syncthing

import "time"

type SystemStatus struct {
	MyID      string `json:"myID"`
	StartTime string `json:"startTime"`
	Uptime    int    `json:"uptime"`
}

type ConnectionInfo struct {
	Connected bool   `json:"connected"`
	Address   string `json:"address"`
	Paused    bool   `json:"paused"`
}

type FolderStatus struct {
	State       string `json:"state"`
	Error       string `json:"error"`
	GlobalFiles int    `json:"globalFiles"`
	GlobalBytes int64  `json:"globalBytes"`
	LocalFiles  int    `json:"localFiles"`
	LocalBytes  int64  `json:"localBytes"`
	NeedFiles   int    `json:"needFiles"`
	NeedBytes   int64  `json:"needBytes"`
	PullErrors  int    `json:"pullErrors"`
	InSyncFiles int    `json:"inSyncFiles"`
	InSyncBytes int64  `json:"inSyncBytes"`
}

type CompletionResponse struct {
	Completion  float64 `json:"completion"`
	GlobalBytes int64   `json:"globalBytes"`
	NeedBytes   int64   `json:"needBytes"`
	GlobalItems int     `json:"globalItems"`
	NeedItems   int     `json:"needItems"`
	NeedDeletes int     `json:"needDeletes"`
}

type LocalChange struct {
	Action   string
	Type     string
	Path     string
	Modified string
	Size     int64
}

type ConfiguredDevice struct {
	ID     string
	Name   string
	Paused bool
}

type DiscoveredDevice struct {
	DeviceID  string
	Addresses []string
}

type PendingDevice struct {
	DeviceID string
	Name     string
	Address  string
	Time     time.Time
}

type PendingFolder struct {
	ID        string
	Label     string
	OfferedBy string
}

type FolderConfig struct {
	ID      string   `json:"id"`
	Path    string   `json:"path"`
	Type    string   `json:"type"`
	Devices []string `json:"-"`
}

type FolderSharingDrift struct {
	FolderID         string
	MissingDeviceIDs []string
}

type SyncProgressInfo struct {
	NeedFiles   int64
	NeedBytes   int64
	GlobalBytes int64
	Percent     int
}

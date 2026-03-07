package sync

import (
	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/syncthing"
)

var log = logging.New("sync")
var stLog = log.WithPrefix("[syncthing]")

func init() {
	syncthing.SetLogger(&syncthingLogAdapter{})
}

type syncthingLogAdapter struct{}

func (a *syncthingLogAdapter) Debug(format string, args ...any) { stLog.Debug(format, args...) }
func (a *syncthingLogAdapter) Info(format string, args ...any)  { stLog.Info(format, args...) }
func (a *syncthingLogAdapter) Error(format string, args ...any) { stLog.Error(format, args...) }

type Client struct {
	*syncthing.Client
	config model.SyncConfig
}

func NewClient(config model.SyncConfig) *Client {
	stConfig := syncthing.Config{
		ListenPort:             config.Syncthing.ListenPort,
		DiscoveryPort:          config.Syncthing.DiscoveryPort,
		GUIPort:                config.Syncthing.GUIPort,
		RelayEnabled:           config.Syncthing.RelayEnabled,
		GlobalDiscoveryEnabled: config.Syncthing.GlobalDiscoveryEnabled,
		BaseURL:                config.Syncthing.BaseURL,
	}
	return &Client{
		Client: syncthing.NewClient(stConfig),
		config: config,
	}
}

func (c *Client) SyncConfig() model.SyncConfig {
	return c.config
}

package syncthing

type Config struct {
	ListenPort             int
	DiscoveryPort          int
	GUIPort                int
	RelayEnabled           bool
	GlobalDiscoveryEnabled bool
	BaseURL                string
}

func DefaultConfig() Config {
	return Config{
		ListenPort:             22100,
		DiscoveryPort:          21127,
		GUIPort:                8484,
		RelayEnabled:           true,
		GlobalDiscoveryEnabled: true,
	}
}

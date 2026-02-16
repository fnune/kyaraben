package model

type SteamShortcutInfo struct {
	AppName       string
	LaunchOptions string
	Tags          []string
	GridAssets    *SteamGridAssets
}

type SteamGridAssets struct {
	Grid    []byte
	Hero    []byte
	Logo    []byte
	Capsule []byte
}

type SteamShortcutProvider interface {
	SteamShortcut(binDir string) *SteamShortcutInfo
}

package steam

type ShortcutManager interface {
	IsAvailable() bool
	Sync(entries []ShortcutEntry) error
}

type ShortcutEntry struct {
	AppName       string
	Exe           string
	StartDir      string
	Icon          string
	LaunchOptions string
	Tags          []string
	GridAssets    *GridAssets
}

type GridAssets struct {
	Grid    []byte
	Hero    []byte
	Logo    []byte
	Capsule []byte
}

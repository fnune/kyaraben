package steam

type ShortcutManager interface {
	IsAvailable() bool
	Sync(entries []ShortcutEntry) (changed bool, err error)
}

type ShortcutEntry struct {
	AppName       string
	Exe           string
	StartDir      string
	Icon          string
	ShortcutPath  string
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

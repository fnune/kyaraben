package esde

import (
	"os"
	"path/filepath"

	"github.com/beevik/etree"
	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/model"
)

var log = logging.New("esde")

var statsElements = []string{"playcount", "playtime", "lastplayed"}

type GamelistSync struct {
	fs      vfs.FS
	esdeDir string
	store   model.StoreReader
}

func NewGamelistSync(fs vfs.FS, esdeDir string, store model.StoreReader) *GamelistSync {
	return &GamelistSync{fs: fs, esdeDir: esdeDir, store: store}
}

func NewDefaultGamelistSync(store model.StoreReader) (*GamelistSync, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &GamelistSync{
		fs:      vfs.OSFS,
		esdeDir: filepath.Join(home, "ES-DE"),
		store:   store,
	}, nil
}

func (g *GamelistSync) SyncSystem(sys model.SystemID) error {
	if err := g.importGamelist(sys); err != nil {
		log.Warn("import gamelist for %s: %v", sys, err)
	}
	if err := g.exportGamelist(sys); err != nil {
		log.Warn("export gamelist for %s: %v", sys, err)
	}
	return nil
}

func (g *GamelistSync) ImportAll(systems []model.SystemID) {
	for _, sys := range systems {
		if err := g.importGamelist(sys); err != nil {
			log.Warn("import gamelist for %s: %v", sys, err)
		}
	}
}

func (g *GamelistSync) ExportAll(systems []model.SystemID) {
	for _, sys := range systems {
		if err := g.exportGamelist(sys); err != nil {
			log.Warn("export gamelist for %s: %v", sys, err)
		}
	}
}

func (g *GamelistSync) importGamelist(sys model.SystemID) error {
	syncPath := filepath.Join(g.store.FrontendGamelistDir(model.FrontendIDESDE, sys), "gamelist.xml")
	esdePath := filepath.Join(g.esdeDir, "gamelists", string(sys), "gamelist.xml")

	synced, err := g.loadGamelist(syncPath)
	if err != nil {
		return nil
	}

	local, _ := g.loadGamelist(esdePath)
	merged := g.merge(synced, local)

	if err := vfs.MkdirAll(g.fs, filepath.Dir(esdePath), 0755); err != nil {
		return err
	}

	merged.Indent(2)
	return merged.WriteToFile(esdePath)
}

func (g *GamelistSync) exportGamelist(sys model.SystemID) error {
	esdePath := filepath.Join(g.esdeDir, "gamelists", string(sys), "gamelist.xml")
	syncPath := filepath.Join(g.store.FrontendGamelistDir(model.FrontendIDESDE, sys), "gamelist.xml")

	doc, err := g.loadGamelist(esdePath)
	if err != nil {
		return nil
	}

	g.stripStats(doc)

	if err := vfs.MkdirAll(g.fs, filepath.Dir(syncPath), 0755); err != nil {
		return err
	}

	doc.Indent(2)
	return doc.WriteToFile(syncPath)
}

func (g *GamelistSync) loadGamelist(path string) (*etree.Document, error) {
	data, err := g.fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	return doc, nil
}

func (g *GamelistSync) stripStats(doc *etree.Document) {
	for _, game := range doc.FindElements("//game") {
		for _, tag := range statsElements {
			if elem := game.SelectElement(tag); elem != nil {
				game.RemoveChild(elem)
			}
		}
	}
}

type gameStats struct {
	playcount  string
	playtime   string
	lastplayed string
}

func (g *GamelistSync) merge(synced, local *etree.Document) *etree.Document {
	if local == nil {
		return synced
	}

	localStats := make(map[string]gameStats)
	for _, game := range local.FindElements("//game") {
		pathElem := game.SelectElement("path")
		if pathElem == nil {
			continue
		}
		path := pathElem.Text()
		stats := gameStats{}
		if elem := game.SelectElement("playcount"); elem != nil {
			stats.playcount = elem.Text()
		}
		if elem := game.SelectElement("playtime"); elem != nil {
			stats.playtime = elem.Text()
		}
		if elem := game.SelectElement("lastplayed"); elem != nil {
			stats.lastplayed = elem.Text()
		}
		if stats.playcount != "" || stats.playtime != "" || stats.lastplayed != "" {
			localStats[path] = stats
		}
	}

	for _, game := range synced.FindElements("//game") {
		pathElem := game.SelectElement("path")
		if pathElem == nil {
			continue
		}
		path := pathElem.Text()
		stats, ok := localStats[path]
		if !ok {
			continue
		}
		if stats.playcount != "" {
			g.setOrCreateElement(game, "playcount", stats.playcount)
		}
		if stats.playtime != "" {
			g.setOrCreateElement(game, "playtime", stats.playtime)
		}
		if stats.lastplayed != "" {
			g.setOrCreateElement(game, "lastplayed", stats.lastplayed)
		}
	}

	return synced
}

func (g *GamelistSync) setOrCreateElement(parent *etree.Element, tag, value string) {
	elem := parent.SelectElement(tag)
	if elem == nil {
		elem = parent.CreateElement(tag)
	}
	elem.SetText(value)
}

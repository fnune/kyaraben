package collections

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
)

type EmulatorLookup func(model.EmulatorID) (model.Emulator, error)

type Service struct {
	fs          vfs.FS
	config      *model.KyarabenConfig
	store       model.StoreReader
	getEmulator EmulatorLookup
	source      *Source
	overrides   *OverrideStore
	configDir   string
	commandPort int
	binDir      string
	coresDir    string
}

type ServiceDeps struct {
	FS            vfs.FS
	Config        *model.KyarabenConfig
	Store         model.StoreReader
	GetEmulator   EmulatorLookup
	Source        *Source
	Overrides     *OverrideStore
	UserConfigDir string
	CommandPort   int
	BinDir        string
	CoresDir      string
}

func NewService(d ServiceDeps) *Service {
	return &Service{
		fs:          d.FS,
		config:      d.Config,
		store:       d.Store,
		getEmulator: d.GetEmulator,
		source:      d.Source,
		overrides:   d.Overrides,
		configDir:   d.UserConfigDir,
		commandPort: d.CommandPort,
		binDir:      d.BinDir,
		coresDir:    d.CoresDir,
	}
}

func (s *Service) IDs() []string {
	ids := make([]string, 0, len(s.config.Collections))
	for id := range s.config.Collections {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (s *Service) collectionConfig(id string) (model.CollectionConfig, error) {
	cfg, ok := s.config.Collections[id]
	if !ok {
		return model.CollectionConfig{}, fmt.Errorf("collection %q is not configured; add [collections.%s] with a url to your config", id, id)
	}
	return cfg, nil
}

func (s *Service) Load(ctx context.Context, id string, maxAge time.Duration) (*Loaded, error) {
	cfg, err := s.collectionConfig(id)
	if err != nil {
		return nil, err
	}
	return s.source.Load(ctx, id, cfg.URL, maxAge)
}

type GameStatus struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Resolution
}

type Status struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	FromCache bool         `json:"fromCache"`
	FetchErr  string       `json:"fetchError,omitempty"`
	Games     []GameStatus `json:"games"`
}

func (s *Service) Status(ctx context.Context, id string, maxAge time.Duration) (*Status, error) {
	loaded, err := s.Load(ctx, id, maxAge)
	if err != nil {
		return nil, err
	}
	cfg, _ := s.collectionConfig(id)
	resolver := s.resolver()

	st := &Status{ID: id, Name: loaded.Manifest.Name, FromCache: loaded.FromCache}
	if loaded.FetchErr != nil {
		st.FetchErr = loaded.FetchErr.Error()
	}
	for _, g := range loaded.Manifest.Games {
		st.Games = append(st.Games, GameStatus{
			ID:         g.ID,
			Title:      g.Title,
			Resolution: resolver.Resolve(g, cfg.Games[g.ID]),
		})
	}
	return st, nil
}

func (s *Service) Sync(ctx context.Context, id string, force bool) ([]FileResult, error) {
	loaded, err := s.Load(ctx, id, 0)
	if err != nil {
		return nil, err
	}
	cfg, _ := s.collectionConfig(id)
	resolver := s.resolver()

	var files []OverrideFile
	for _, g := range loaded.Manifest.Games {
		res := resolver.Resolve(g, cfg.Games[g.ID])
		gameFiles, err := GameOverrides(s.configDir, g, res, s.deviceSettings(cfg, g.ID), s.fs)
		if err != nil {
			return nil, fmt.Errorf("game %s: %w", g.ID, err)
		}
		files = append(files, gameFiles...)
	}
	return s.overrides.Write(id, files, true, force)
}

type LaunchPlan struct {
	CollectionID string       `json:"collection"`
	GameID       string       `json:"game"`
	Title        string       `json:"title"`
	Resolution   Resolution   `json:"resolution"`
	Command      []string     `json:"command"`
	Menu         bool         `json:"menu"`
	CommandPort  int          `json:"-"`
	Files        []FileResult `json:"files"`
}

type LaunchOptions struct {
	// Menu opens RetroArch's quick menu over the game, to change its settings.
	Menu bool
}

func (s *Service) Plan(ctx context.Context, id, gameID string, opts LaunchOptions) (*LaunchPlan, []OverrideFile, error) {
	loaded, err := s.Load(ctx, id, time.Minute)
	if err != nil {
		return nil, nil, err
	}
	game, ok := loaded.Manifest.Game(gameID)
	if !ok {
		return nil, nil, fmt.Errorf("collection %q has no game %q", id, gameID)
	}
	cfg, _ := s.collectionConfig(id)

	res := s.resolver().Resolve(game, cfg.Games[gameID])
	if !res.Playable {
		return nil, nil, fmt.Errorf("%s is not playable: %s", game.Title, res.Reason)
	}

	device := s.deviceSettings(cfg, gameID)
	if opts.Menu {
		device.CommandPort = s.commandPort
	}
	files, err := GameOverrides(s.configDir, game, res, device, s.fs)
	if err != nil {
		return nil, nil, err
	}
	command, err := s.command(res)
	if err != nil {
		return nil, nil, err
	}
	return &LaunchPlan{
		CollectionID: id,
		GameID:       gameID,
		Title:        game.Title,
		Resolution:   res,
		Command:      command,
		Menu:         opts.Menu,
		CommandPort:  device.CommandPort,
	}, files, nil
}

func (s *Service) Prepare(ctx context.Context, id, gameID string, opts LaunchOptions) (*LaunchPlan, error) {
	plan, files, err := s.Plan(ctx, id, gameID, opts)
	if err != nil {
		return nil, err
	}
	plan.Files, err = s.overrides.Write(id, files, false, false)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) command(res Resolution) ([]string, error) {
	emu, err := s.getEmulator(res.Emulator)
	if err != nil {
		return nil, err
	}
	binary := filepath.Join(s.binDir, emu.Launcher.Binary)
	if _, err := s.fs.Stat(binary); err != nil {
		return nil, fmt.Errorf("%s is not installed; run kyaraben apply", emu.Launcher.Binary)
	}
	core := filepath.Join(s.coresDir, emu.Launcher.CoreName+".so")
	if _, err := s.fs.Stat(core); err != nil {
		return nil, fmt.Errorf("core %s is not installed; run kyaraben apply", emu.Launcher.CoreName)
	}
	return []string{binary, "-f", "-L", core, res.RomPath}, nil
}

func (s *Service) RemoveAllOverrides() ([]FileResult, error) {
	return s.overrides.RemoveAll()
}

func (s *Service) resolver() *Resolver {
	return NewResolver(s.fs, s.store, s.config.Systems)
}

func (s *Service) deviceSettings(cfg model.CollectionConfig, gameID string) DeviceSettings {
	return DeviceSettings{Vertical: cfg.VerticalFor(gameID), VRR: cfg.VRR}
}

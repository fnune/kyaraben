package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fnune/kyaraben/internal/collections"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/version"
)

type CollectionsCmd struct {
	List   CollectionsListCmd   `cmd:"" help:"List configured collections."`
	Status CollectionsStatusCmd `cmd:"" help:"Show which games in a collection are playable on this device."`
	Sync   CollectionsSyncCmd   `cmd:"" help:"Fetch manifests and write per-game emulator settings."`
	Serve  CollectionsServeCmd  `cmd:"" help:"Serve collections to their web pages on 127.0.0.1."`
}

type LaunchCmd struct {
	Collection string `arg:"" help:"Collection ID."`
	Game       string `arg:"" help:"Game ID within the collection."`
	DryRun     bool   `help:"Print the command and per-game settings without writing or running anything."`
}

func (c *Context) collectionsDir() (string, error) {
	return collections.Dir(c.Paths)
}

func (c *Context) NewCollectionsService() (*collections.Service, error) {
	cfg, err := c.LoadConfig()
	if err != nil {
		return nil, err
	}
	return c.newCollectionsService(cfg)
}

func (c *Context) newCollectionsService(cfg *model.KyarabenConfig) (*collections.Service, error) {
	store, err := c.NewCollection(cfg)
	if err != nil {
		return nil, err
	}
	return collections.NewDefaultService(c.FS, c.Paths, c.Resolver, cfg, store, c.NewRegistry().GetEmulator)
}

// setUpCollections runs after apply. Collections are optional, so failures
// are reported without failing the apply.
func setUpCollections(ctx *Context, cfg *model.KyarabenConfig) {
	units, err := collections.NewDefaultUnits(ctx.FS, ctx.Paths)
	if err != nil {
		fmt.Printf("Warning: collections: %v\n", err)
		return
	}
	if len(cfg.Collections) == 0 && !units.Installed() {
		return
	}
	svc, err := ctx.newCollectionsService(cfg)
	if err != nil {
		fmt.Printf("Warning: collections: %v\n", err)
		return
	}
	if len(cfg.Collections) > 0 {
		fmt.Println("Setting up collections...")
	}
	report := collections.SetUp(context.Background(), svc, units, ctx.Paths, len(cfg.Collections) > 0)
	for _, id := range svc.IDs() {
		if results, ok := report.Results[id]; ok {
			printFileResults(id, results)
		}
	}
	for _, w := range report.Warnings {
		fmt.Printf("Warning: %s\n", w)
	}
	fmt.Println()
}

type CollectionsListCmd struct{}

func (cmd *CollectionsListCmd) Run(ctx *Context) error {
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}
	if len(cfg.Collections) == 0 {
		fmt.Println("No collections configured. Add one to config.toml:")
		fmt.Println()
		fmt.Println("  [collections.shmups]")
		fmt.Println(`  url = "https://example.com/collection.json"`)
		return nil
	}
	svc, err := ctx.NewCollectionsService()
	if err != nil {
		return err
	}
	for _, id := range svc.IDs() {
		fmt.Printf("%s  %s\n", id, cfg.Collections[id].URL)
	}
	return nil
}

type CollectionsStatusCmd struct {
	Collection string `arg:"" help:"Collection ID."`
	JSON       bool   `help:"Print machine-readable JSON."`
}

func (cmd *CollectionsStatusCmd) Run(ctx *Context) error {
	svc, err := ctx.NewCollectionsService()
	if err != nil {
		return err
	}
	st, err := svc.Status(context.Background(), cmd.Collection, 0)
	if err != nil {
		return err
	}
	if cmd.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(st)
	}

	if st.FetchErr != "" {
		fmt.Printf("Using cached manifest: %s\n\n", st.FetchErr)
	}
	playable := 0
	for _, g := range st.Games {
		if g.Playable {
			playable++
			fmt.Printf("  ✓ %-28s %s (%s)\n", g.ID, g.Set, g.Emulator)
		} else {
			fmt.Printf("  ✗ %-28s %s\n", g.ID, g.Reason)
		}
	}
	fmt.Printf("\n%s: %d of %d games playable\n", st.Name, playable, len(st.Games))
	return nil
}

type CollectionsSyncCmd struct {
	Collection string `arg:"" optional:"" help:"Collection ID (default: all)."`
	Force      bool   `help:"Overwrite per-game files you edited."`
}

func (cmd *CollectionsSyncCmd) Run(ctx *Context) error {
	svc, err := ctx.NewCollectionsService()
	if err != nil {
		return err
	}
	ids := svc.IDs()
	if cmd.Collection != "" {
		ids = []string{cmd.Collection}
	}
	for _, id := range ids {
		results, err := svc.Sync(context.Background(), id, cmd.Force)
		if err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
		printFileResults(id, results)
	}
	return nil
}

func printFileResults(id string, results []collections.FileResult) {
	counts := map[collections.FileOutcome]int{}
	for _, r := range results {
		counts[r.Outcome]++
		switch r.Outcome {
		case collections.FileForeign:
			fmt.Printf("  kept %s: it was not created by kyaraben\n", r.Path)
		case collections.FileModified:
			fmt.Printf("  kept %s: it was edited (use --force to replace)\n", r.Path)
		}
	}
	fmt.Printf("%s: %d written, %d unchanged, %d removed\n", id,
		counts[collections.FileWritten], counts[collections.FileUnchanged], counts[collections.FileRemoved])
}

func (cmd *LaunchCmd) Run(ctx *Context) error {
	svc, err := ctx.NewCollectionsService()
	if err != nil {
		return err
	}
	if cmd.DryRun {
		plan, files, err := svc.Plan(context.Background(), cmd.Collection, cmd.Game)
		if err != nil {
			return err
		}
		for _, f := range files {
			fmt.Printf("# %s\n%s\n", f.Path, f.Content)
		}
		fmt.Println(strings.Join(plan.Command, " "))
		return nil
	}

	plan, err := svc.Prepare(context.Background(), cmd.Collection, cmd.Game)
	if err != nil {
		return err
	}
	for _, f := range plan.Files {
		if f.Outcome == collections.FileForeign || f.Outcome == collections.FileModified {
			fmt.Fprintf(os.Stderr, "Note: kept %s (%s); the collection's settings for this game are not applied\n", f.Path, f.Outcome)
		}
	}
	dir, err := ctx.collectionsDir()
	if err != nil {
		return err
	}
	sessions := collections.NewSessions(collections.ExecStarter{}, filepath.Join(dir, "logs"))
	session, err := sessions.Launch(plan)
	if err != nil {
		return err
	}
	fmt.Printf("Playing %s (%s, %s)\n", plan.Title, plan.Resolution.Set, plan.Resolution.Emulator)
	for {
		s, _ := sessions.Get(session.ID)
		if !s.Running {
			fmt.Printf("Played for %s\n", s.EndedAt.Sub(s.StartedAt).Round(time.Second))
			if s.Error != "" {
				return fmt.Errorf("emulator exited: %s", s.Error)
			}
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
}

type CollectionsServeCmd struct {
	Port int `help:"Port to listen on (default depends on the instance)."`
}

func (cmd *CollectionsServeCmd) Run(ctx *Context) error {
	svc, err := ctx.NewCollectionsService()
	if err != nil {
		return err
	}
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}
	dir, err := ctx.collectionsDir()
	if err != nil {
		return err
	}

	port := cmd.Port
	if port == 0 {
		port = collections.Port(ctx.Paths)
	}
	listener, err := collectionsListener(port)
	if err != nil {
		return err
	}

	sessions := collections.NewSessions(collections.ExecStarter{}, filepath.Join(dir, "logs"))
	server := &http.Server{
		Handler:           collections.NewServer(svc, sessions, cfg.Collections, port, version.Get()).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-sigCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	fmt.Printf("Serving collections on http://127.0.0.1:%d\n", port)
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// collectionsListener prefers a socket passed in by systemd socket activation
// (sd_listen_fds protocol, first fd is 3) so the service starts on first use.
func collectionsListener(port int) (net.Listener, error) {
	if os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid()) && os.Getenv("LISTEN_FDS") != "" && os.Getenv("LISTEN_FDS") != "0" {
		f := os.NewFile(3, "systemd-socket")
		defer func() { _ = f.Close() }()
		return net.FileListener(f)
	}
	return net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
}

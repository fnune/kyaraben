package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/daemon"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/nix"
)

type DaemonCmd struct{}

func (cmd *DaemonCmd) Run(ctx *Context) error {
	registry := ctx.NewRegistry()
	nixClient, err := ctx.NewNixClient()
	if err != nil {
		return fmt.Errorf("creating nix client: %w", err)
	}
	flakeGenerator := nix.NewFlakeGenerator(registry)
	configWriter := emulators.NewConfigWriter()
	launcherManager := launcher.NewManager()

	d := daemon.New(ctx.ConfigPath, registry, nixClient, flakeGenerator, configWriter, launcherManager)

	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Send ready event
	if err := encoder.Encode(daemon.Event{
		Type: daemon.EventReady,
		Data: map[string]string{"version": "0.1.0"},
	}); err != nil {
		return fmt.Errorf("sending ready event: %w", err)
	}

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var cmd daemon.Command
		if err := json.Unmarshal(line, &cmd); err != nil {
			_ = encoder.Encode(daemon.Event{
				Type: daemon.EventError,
				Data: map[string]string{"error": fmt.Sprintf("invalid command: %v", err)},
			})
			continue
		}

		// Emit function streams events immediately to stdout
		emit := func(event daemon.Event) {
			_ = encoder.Encode(event)
		}

		events := d.HandleWithEmit(cmd, emit)
		for _, event := range events {
			if err := encoder.Encode(event); err != nil {
				return fmt.Errorf("sending event: %w", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	return nil
}

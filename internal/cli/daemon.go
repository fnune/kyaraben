package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

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
	var encoderMu sync.Mutex

	sendEvent := func(event daemon.Event) {
		encoderMu.Lock()
		defer encoderMu.Unlock()
		_ = encoder.Encode(event)
	}

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
			sendEvent(daemon.Event{
				Type: daemon.EventError,
				Data: map[string]string{"error": fmt.Sprintf("invalid command: %v", err)},
			})
			continue
		}

		if cmd.Type == daemon.CmdApply {
			go func() {
				events := d.HandleWithEmit(cmd, sendEvent)
				for _, event := range events {
					sendEvent(event)
				}
			}()
		} else {
			events := d.HandleWithEmit(cmd, sendEvent)
			for _, event := range events {
				sendEvent(event)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	return nil
}

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
	launcherManager, err := launcher.NewManager()
	if err != nil {
		return fmt.Errorf("creating launcher manager: %w", err)
	}

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
		Type: daemon.EventTypeReady,
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
				Type: daemon.EventTypeError,
				Data: map[string]string{"error": fmt.Sprintf("invalid command: %v", err)},
			})
			continue
		}

		var events []daemon.Event
		switch cmd.Type {
		case daemon.CommandTypeApply:
			go func() {
				events := d.HandleWithEmit(cmd, sendEvent)
				for _, event := range events {
					sendEvent(event)
				}
			}()
			continue
		case daemon.CommandTypeSetConfig:
			var setConfigCmd daemon.SetConfigCommand
			if err := json.Unmarshal(line, &setConfigCmd); err != nil {
				sendEvent(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid set_config command: %v", err)},
				})
				continue
			}
			events = d.HandleSetConfig(setConfigCmd, sendEvent)
		case daemon.CommandTypeSyncAddDevice:
			var syncAddCmd daemon.SyncAddDeviceCommand
			if err := json.Unmarshal(line, &syncAddCmd); err != nil {
				sendEvent(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid sync_add_device command: %v", err)},
				})
				continue
			}
			events = d.HandleSyncAddDevice(syncAddCmd, sendEvent)
		case daemon.CommandTypeSyncRemoveDevice:
			var syncRemoveCmd daemon.SyncRemoveDeviceCommand
			if err := json.Unmarshal(line, &syncRemoveCmd); err != nil {
				sendEvent(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid sync_remove_device command: %v", err)},
				})
				continue
			}
			events = d.HandleSyncRemoveDevice(syncRemoveCmd, sendEvent)
		default:
			events = d.HandleWithEmit(cmd, sendEvent)
		}

		for _, event := range events {
			sendEvent(event)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	return nil
}

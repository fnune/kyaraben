package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/daemon"
	"github.com/fnune/kyaraben/internal/emulators"
	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
)

type DaemonCmd struct{}

func (cmd *DaemonCmd) Run(ctx *Context) error {
	registry := ctx.NewRegistry()
	installer, err := ctx.NewInstaller()
	if err != nil {
		return fmt.Errorf("creating installer: %w", err)
	}
	configWriter := emulators.NewDefaultConfigWriter()
	launcherManager, err := launcher.NewManager(vfs.OSFS, ctx.GetPaths(), model.NewDefaultResolver())
	if err != nil {
		return fmt.Errorf("creating launcher manager: %w", err)
	}

	manifestPath, err := ctx.GetPaths().ManifestPath()
	if err != nil {
		return fmt.Errorf("getting manifest path: %w", err)
	}

	stateDir, err := ctx.stateDir()
	if err != nil {
		return fmt.Errorf("getting state dir: %w", err)
	}

	configPath, err := ctx.GetConfigPath()
	if err != nil {
		return fmt.Errorf("getting config path: %w", err)
	}

	d := daemon.NewDefault(ctx.GetPaths(), configPath, stateDir, manifestPath, registry, installer, configWriter, launcherManager)

	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	var encoderMu sync.Mutex

	sendEvent := func(event daemon.Event) {
		encoderMu.Lock()
		defer encoderMu.Unlock()
		_ = encoder.Encode(event)
	}

	sendEventWithID := func(event daemon.Event, id string) {
		event.ID = id
		sendEvent(event)
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

		cmdID := cmd.ID
		emitWithID := func(event daemon.Event) {
			sendEventWithID(event, cmdID)
		}

		var events []daemon.Event
		switch cmd.Type {
		case daemon.CommandTypeApply:
			go func(id string) {
				emitForApply := func(event daemon.Event) {
					sendEventWithID(event, id)
				}
				events := d.HandleWithEmit(cmd, emitForApply)
				for _, event := range events {
					sendEventWithID(event, id)
				}
			}(cmdID)
			continue
		case daemon.CommandTypeSetConfig:
			var setConfigCmd daemon.SetConfigCommand
			if err := json.Unmarshal(line, &setConfigCmd); err != nil {
				sendEventWithID(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid set_config command: %v", err)},
				}, cmdID)
				continue
			}
			events = d.HandleSetConfig(setConfigCmd, emitWithID)
		case daemon.CommandTypeSyncRemoveDevice:
			var syncRemoveCmd daemon.SyncRemoveDeviceCommand
			if err := json.Unmarshal(line, &syncRemoveCmd); err != nil {
				sendEventWithID(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid sync_remove_device command: %v", err)},
				}, cmdID)
				continue
			}
			events = d.HandleSyncRemoveDevice(syncRemoveCmd, emitWithID)
		case daemon.CommandTypeSyncRevertFolder:
			var revertCmd daemon.SyncRevertFolderCommand
			if err := json.Unmarshal(line, &revertCmd); err != nil {
				sendEventWithID(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid sync_revert_folder command: %v", err)},
				}, cmdID)
				continue
			}
			events = d.HandleSyncRevertFolder(revertCmd, emitWithID)
		case daemon.CommandTypeSyncLocalChanges:
			var changesCmd daemon.SyncLocalChangesCommand
			if err := json.Unmarshal(line, &changesCmd); err != nil {
				sendEventWithID(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid sync_local_changes command: %v", err)},
				}, cmdID)
				continue
			}
			events = d.HandleSyncLocalChanges(changesCmd, emitWithID)
		case daemon.CommandTypeSyncStartPairing:
			go func(id string) {
				emitForPairing := func(event daemon.Event) {
					sendEventWithID(event, id)
				}
				events := d.HandleSyncStartPairing(cmd, emitForPairing)
				for _, event := range events {
					sendEventWithID(event, id)
				}
			}(cmdID)
			continue
		case daemon.CommandTypeSyncEnable:
			var enableCmd daemon.SyncEnableCommand
			if err := json.Unmarshal(line, &enableCmd); err != nil {
				sendEventWithID(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid sync_enable command: %v", err)},
				}, cmdID)
				continue
			}
			go func(id string, c daemon.SyncEnableCommand) {
				emitForEnable := func(event daemon.Event) {
					sendEventWithID(event, id)
				}
				events := d.HandleSyncEnable(c, emitForEnable)
				for _, event := range events {
					sendEventWithID(event, id)
				}
			}(cmdID, enableCmd)
			continue
		case daemon.CommandTypeSyncJoinPeer:
			var joinCmd daemon.SyncJoinPeerCommand
			if err := json.Unmarshal(line, &joinCmd); err != nil {
				sendEventWithID(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid sync_join_peer command: %v", err)},
				}, cmdID)
				continue
			}
			go func(id string, cmd daemon.SyncJoinPeerCommand) {
				emitForPairing := func(event daemon.Event) {
					sendEventWithID(event, id)
				}
				events := d.HandleSyncJoinPeer(cmd, emitForPairing)
				for _, event := range events {
					sendEventWithID(event, id)
				}
			}(cmdID, joinCmd)
			continue
		case daemon.CommandTypeInstallKyaraben:
			var installCmd daemon.InstallKyarabenCommand
			if err := json.Unmarshal(line, &installCmd); err != nil {
				sendEventWithID(daemon.Event{
					Type: daemon.EventTypeError,
					Data: map[string]string{"error": fmt.Sprintf("invalid install_kyaraben command: %v", err)},
				}, cmdID)
				continue
			}
			events = d.HandleInstallKyaraben(installCmd, emitWithID)
		default:
			events = d.HandleWithEmit(cmd, emitWithID)
		}

		for _, event := range events {
			sendEventWithID(event, cmdID)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	return nil
}

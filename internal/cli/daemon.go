package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fnune/kyaraben/internal/daemon"
)

// DaemonCmd runs kyaraben in daemon mode.
type DaemonCmd struct{}

// Run executes the daemon command.
func (cmd *DaemonCmd) Run(ctx *Context) error {
	d := daemon.New(ctx.ConfigPath)

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
			encoder.Encode(daemon.Event{
				Type: daemon.EventError,
				Data: map[string]string{"error": fmt.Sprintf("invalid command: %v", err)},
			})
			continue
		}

		events := d.Handle(cmd)
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

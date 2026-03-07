package pairing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
	"github.com/fnune/kyaraben/internal/syncthing"
)

type Flow struct {
	client      syncthing.SyncClient
	relayClient *syncthing.RelayClient
	ui          ui.UI
}

func NewFlow(client syncthing.SyncClient, relayClient *syncthing.RelayClient, ui ui.UI) *Flow {
	return &Flow{
		client:      client,
		relayClient: relayClient,
		ui:          ui,
	}
}

func (f *Flow) Run(ctx context.Context) error {
	items := []ui.MenuItem{
		{Label: "Generate pairing code", Value: "generate"},
		{Label: "Enter pairing code", Value: "enter"},
	}

	idx, action, err := f.ui.Menu().Show(items, ui.MenuOptions{
		Title:    "Pair new device",
		ShowBack: true,
	})
	if err != nil {
		return err
	}
	if action == ui.ActionBack {
		return nil
	}

	switch items[idx].Value {
	case "generate":
		return f.generateCode(ctx)
	case "enter":
		return f.enterCode(ctx)
	}

	return nil
}

func (f *Flow) generateCode(ctx context.Context) error {
	deviceID, err := f.client.GetDeviceID(ctx)
	if err != nil {
		return fmt.Errorf("get device ID: %w", err)
	}

	session, err := f.relayClient.CreateSession(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("create pairing session: %w", err)
	}

	if err := f.ui.Presenter().ShowMessage("Pairing code", session.Code); err != nil {
		return err
	}

	pollCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	responderID, err := f.pollForResponse(pollCtx, session.Code)
	if err != nil {
		_ = f.ui.Presenter().Close()
		return fmt.Errorf("waiting for peer: %w", err)
	}

	_ = f.ui.Presenter().Close()

	if err := f.client.AddDevice(ctx, responderID, ""); err != nil {
		return fmt.Errorf("add device: %w", err)
	}

	if err := f.client.ShareFoldersWithDevice(ctx, responderID); err != nil {
		return fmt.Errorf("share folders: %w", err)
	}

	if err := f.ui.Presenter().ShowMessage("Paired", "Device paired successfully"); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	_ = f.ui.Presenter().Close()

	return nil
}

func (f *Flow) enterCode(ctx context.Context) error {
	code, err := f.ui.Keyboard().GetInput(ui.KeyboardOptions{
		Title:     "Enter pairing code",
		MaxLength: 6,
		Uppercase: true,
	})
	if err != nil {
		return err
	}
	if code == "" {
		return nil
	}

	code = strings.ToUpper(strings.TrimSpace(code))

	deviceID, err := f.client.GetDeviceID(ctx)
	if err != nil {
		return fmt.Errorf("get device ID: %w", err)
	}

	session, err := f.relayClient.GetSession(ctx, code)
	if err != nil {
		return fmt.Errorf("get pairing session: %w", err)
	}

	if err := f.relayClient.SubmitResponse(ctx, code, deviceID); err != nil {
		return fmt.Errorf("submit response: %w", err)
	}

	initiatorID := session.DeviceID
	if err := f.client.AddDevice(ctx, initiatorID, ""); err != nil {
		return fmt.Errorf("add device: %w", err)
	}

	if err := f.client.ShareFoldersWithDevice(ctx, initiatorID); err != nil {
		return fmt.Errorf("share folders: %w", err)
	}

	if err := f.ui.Presenter().ShowMessage("Paired", "Device paired successfully"); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	_ = f.ui.Presenter().Close()

	return nil
}

func (f *Flow) pollForResponse(ctx context.Context, code string) (string, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			resp, err := f.relayClient.GetResponse(ctx, code)
			if err != nil {
				continue
			}
			if resp != nil && resp.Ready && resp.DeviceID != "" {
				return resp.DeviceID, nil
			}
		}
	}
}

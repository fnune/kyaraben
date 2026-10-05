package sync

import (
	"context"
	"fmt"

	"github.com/fnune/kyaraben/internal/syncthing"
)

func ShareFoldersWithConfiguredDevices(ctx context.Context, client syncthing.SyncClient) (int, error) {
	localDeviceID, err := client.GetDeviceID(ctx)
	if err != nil {
		return 0, fmt.Errorf("getting local device ID: %w", err)
	}

	devices, err := client.GetConfiguredDevices(ctx)
	if err != nil {
		return 0, fmt.Errorf("getting configured devices: %w", err)
	}
	if len(devices) == 0 {
		return 0, nil
	}

	deviceIDs := make([]string, len(devices))
	for i, dev := range devices {
		deviceIDs[i] = dev.ID
	}

	folders, err := client.GetFoldersWithDevices(ctx)
	if err != nil {
		return 0, fmt.Errorf("getting folders with devices: %w", err)
	}

	drift := ComputeFolderSharingDrift(folders, deviceIDs, localDeviceID)
	if len(drift) == 0 {
		return 0, nil
	}

	if err := client.ReconcileFolderSharing(ctx, drift); err != nil {
		return 0, fmt.Errorf("sharing folders: %w", err)
	}
	return len(drift), nil
}

func ComputeFolderSharingDrift(
	folders []FolderConfig,
	configuredDeviceIDs []string,
	localDeviceID string,
) []FolderSharingDrift {
	var drift []FolderSharingDrift

	for _, folder := range folders {
		folderDeviceSet := make(map[string]bool)
		for _, deviceID := range folder.Devices {
			folderDeviceSet[deviceID] = true
		}

		var missing []string
		for _, deviceID := range configuredDeviceIDs {
			if deviceID == localDeviceID {
				continue
			}
			if !folderDeviceSet[deviceID] {
				missing = append(missing, deviceID)
			}
		}

		if len(missing) > 0 {
			drift = append(drift, FolderSharingDrift{
				FolderID:         folder.ID,
				MissingDeviceIDs: missing,
			})
		}
	}

	return drift
}

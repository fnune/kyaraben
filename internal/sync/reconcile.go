package sync

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

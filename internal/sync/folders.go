package sync

// ComputeMissingDevices returns which devices are missing from which folders.
// The returned map is folder ID -> slice of missing device IDs.
func ComputeMissingDevices(folders []FolderConfig, wantDeviceIDs []string) map[string][]string {
	if len(wantDeviceIDs) == 0 {
		return nil
	}

	wantSet := make(map[string]bool, len(wantDeviceIDs))
	for _, id := range wantDeviceIDs {
		wantSet[id] = true
	}

	result := make(map[string][]string)
	for _, folder := range folders {
		hasDevice := make(map[string]bool, len(folder.Devices))
		for _, id := range folder.Devices {
			hasDevice[id] = true
		}

		var missing []string
		for _, id := range wantDeviceIDs {
			if !hasDevice[id] {
				missing = append(missing, id)
			}
		}

		if len(missing) > 0 {
			result[folder.ID] = missing
		}
	}

	return result
}

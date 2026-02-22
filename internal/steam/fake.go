package steam

type FakeManager struct {
	Available     bool
	SyncedEntries []ShortcutEntry
	RemovedAppIDs []uint32
	SyncError     error
	RemoveError   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{Available: true}
}

func (f *FakeManager) IsAvailable() bool {
	return f.Available
}

func (f *FakeManager) Sync(entries []ShortcutEntry) (bool, error) {
	if f.SyncError != nil {
		return false, f.SyncError
	}
	changed := len(entries) != len(f.SyncedEntries)
	f.SyncedEntries = entries
	return changed, nil
}

func (f *FakeManager) RemoveShortcuts(appIDs []uint32) error {
	if f.RemoveError != nil {
		return f.RemoveError
	}
	f.RemovedAppIDs = append(f.RemovedAppIDs, appIDs...)
	return nil
}

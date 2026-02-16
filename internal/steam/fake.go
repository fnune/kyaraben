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

func (f *FakeManager) Sync(entries []ShortcutEntry) error {
	if f.SyncError != nil {
		return f.SyncError
	}
	f.SyncedEntries = entries
	return nil
}

func (f *FakeManager) RemoveShortcuts(appIDs []uint32) error {
	if f.RemoveError != nil {
		return f.RemoveError
	}
	f.RemovedAppIDs = append(f.RemovedAppIDs, appIDs...)
	return nil
}

package model

// HotkeyID identifies a hotkey action.
// Constants use the full type name as prefix (HotkeyID*) because tygo's
// enum_style: union requires the prefix to match the type name exactly.
type HotkeyID string

const (
	HotkeyIDSaveState        HotkeyID = "savestate"
	HotkeyIDLoadState        HotkeyID = "loadstate"
	HotkeyIDNextSlot         HotkeyID = "nextslot"
	HotkeyIDPrevSlot         HotkeyID = "prevslot"
	HotkeyIDFastForward      HotkeyID = "fastforward"
	HotkeyIDRewind           HotkeyID = "rewind"
	HotkeyIDPause            HotkeyID = "pause"
	HotkeyIDScreenshot       HotkeyID = "screenshot"
	HotkeyIDQuit             HotkeyID = "quit"
	HotkeyIDToggleFullscreen HotkeyID = "fullscreen"
	HotkeyIDOpenMenu         HotkeyID = "menu"
)

// HotkeyMappings declares which hotkeys an emulator supports.
// Each non-nil field indicates the emulator supports that hotkey.
type HotkeyMappings struct {
	SaveState        *HotkeyKey
	LoadState        *HotkeyKey
	NextSlot         *HotkeyKey
	PrevSlot         *HotkeyKey
	FastForward      *HotkeyKey
	Rewind           *HotkeyKey
	Pause            *HotkeyKey
	Screenshot       *HotkeyKey
	Quit             *HotkeyKey
	ToggleFullscreen *HotkeyKey
	OpenMenu         *HotkeyKey
}

// HotkeyKey holds the emulator-specific config key for a hotkey.
type HotkeyKey struct {
	Key string
}

// SupportedHotkeys returns the list of hotkey IDs this mapping supports.
func (m *HotkeyMappings) SupportedHotkeys() []HotkeyID {
	var ids []HotkeyID
	if m.SaveState != nil {
		ids = append(ids, HotkeyIDSaveState)
	}
	if m.LoadState != nil {
		ids = append(ids, HotkeyIDLoadState)
	}
	if m.NextSlot != nil {
		ids = append(ids, HotkeyIDNextSlot)
	}
	if m.PrevSlot != nil {
		ids = append(ids, HotkeyIDPrevSlot)
	}
	if m.FastForward != nil {
		ids = append(ids, HotkeyIDFastForward)
	}
	if m.Rewind != nil {
		ids = append(ids, HotkeyIDRewind)
	}
	if m.Pause != nil {
		ids = append(ids, HotkeyIDPause)
	}
	if m.Screenshot != nil {
		ids = append(ids, HotkeyIDScreenshot)
	}
	if m.Quit != nil {
		ids = append(ids, HotkeyIDQuit)
	}
	if m.ToggleFullscreen != nil {
		ids = append(ids, HotkeyIDToggleFullscreen)
	}
	if m.OpenMenu != nil {
		ids = append(ids, HotkeyIDOpenMenu)
	}
	return ids
}

// Binding returns the HotkeyBinding for a given ID from a HotkeyConfig.
func (id HotkeyID) Binding(hk HotkeyConfig) HotkeyBinding {
	switch id {
	case HotkeyIDSaveState:
		return hk.SaveState
	case HotkeyIDLoadState:
		return hk.LoadState
	case HotkeyIDNextSlot:
		return hk.NextSlot
	case HotkeyIDPrevSlot:
		return hk.PrevSlot
	case HotkeyIDFastForward:
		return hk.FastForward
	case HotkeyIDRewind:
		return hk.Rewind
	case HotkeyIDPause:
		return hk.Pause
	case HotkeyIDScreenshot:
		return hk.Screenshot
	case HotkeyIDQuit:
		return hk.Quit
	case HotkeyIDToggleFullscreen:
		return hk.ToggleFullscreen
	case HotkeyIDOpenMenu:
		return hk.OpenMenu
	}
	return HotkeyBinding{}
}

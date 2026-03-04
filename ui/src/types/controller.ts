export const SDL_BUTTONS = [
  { value: 'A', label: 'South (A)' },
  { value: 'B', label: 'East (B)' },
  { value: 'X', label: 'West (X)' },
  { value: 'Y', label: 'North (Y)' },
  { value: 'LeftShoulder', label: 'LB' },
  { value: 'RightShoulder', label: 'RB' },
  { value: 'LeftTrigger', label: 'LT' },
  { value: 'RightTrigger', label: 'RT' },
  { value: 'DPadUp', label: 'D-pad up' },
  { value: 'DPadDown', label: 'D-pad down' },
  { value: 'DPadLeft', label: 'D-pad left' },
  { value: 'DPadRight', label: 'D-pad right' },
  { value: 'LeftStick', label: 'L3' },
  { value: 'RightStick', label: 'R3' },
  { value: 'Back', label: 'Back/Select' },
  { value: 'Start', label: 'Start' },
  { value: 'Guide', label: 'Guide' },
] as const

export const MODIFIER_BUTTONS = [
  { value: 'Back', label: 'Back/Select' },
  { value: 'Start', label: 'Start' },
  { value: 'Guide', label: 'Guide' },
] as const

export const HOTKEY_ACTIONS = [
  { key: 'saveState', label: 'Save state' },
  { key: 'loadState', label: 'Load state' },
  { key: 'nextSlot', label: 'Next slot' },
  { key: 'prevSlot', label: 'Previous slot' },
  { key: 'fastForward', label: 'Fast forward' },
  { key: 'rewind', label: 'Rewind' },
  { key: 'pause', label: 'Pause' },
  { key: 'screenshot', label: 'Screenshot' },
  { key: 'quit', label: 'Quit' },
  { key: 'toggleFullscreen', label: 'Toggle fullscreen' },
  { key: 'openMenu', label: 'Open menu' },
] as const

export type HotkeyActionKey = (typeof HOTKEY_ACTIONS)[number]['key']

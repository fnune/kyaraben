import {
  HOTKEY_ACTIONS,
  type HotkeyActionKey,
  MODIFIER_BUTTONS,
  SDL_BUTTONS,
} from '@shared/controller'
import { useMemo } from 'react'
import { Select } from '@/lib/Select'

export interface HotkeyConfig {
  modifier: string
  saveState: string
  loadState: string
  nextSlot: string
  prevSlot: string
  fastForward: string
  rewind: string
  pause: string
  screenshot: string
  quit: string
  toggleFullscreen: string
  openMenu: string
}

export interface HotkeySettingsProps {
  readonly hotkeys: HotkeyConfig
  readonly onModifierChange: (value: string) => void
  readonly onActionChange: (key: HotkeyActionKey, value: string) => void
}

export function HotkeySettings({ hotkeys, onModifierChange, onActionChange }: HotkeySettingsProps) {
  const duplicates = useMemo(() => {
    const seen = new Map<string, HotkeyActionKey[]>()
    for (const action of HOTKEY_ACTIONS) {
      const button = hotkeys[action.key]
      const existing = seen.get(button) ?? []
      existing.push(action.key)
      seen.set(button, existing)
    }
    const result = new Set<HotkeyActionKey>()
    for (const keys of seen.values()) {
      if (keys.length > 1) {
        for (const key of keys) {
          result.add(key)
        }
      }
    }
    return result
  }, [hotkeys])

  const modifierLabel = MODIFIER_BUTTONS.find((b) => b.value === hotkeys.modifier)?.label ?? 'Back'

  return (
    <div>
      <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block">
        Hotkeys
      </span>

      <p className="text-sm text-on-surface-muted mt-1 mb-3">
        Hold the modifier button and press an action button to trigger hotkeys during gameplay.
      </p>

      <div className="space-y-4">
        <div className="flex items-center gap-3">
          <span className="text-sm text-on-surface min-w-[100px]">Modifier</span>
          <Select
            value={hotkeys.modifier}
            options={MODIFIER_BUTTONS.map((b) => ({ value: b.value, label: b.label }))}
            onChange={onModifierChange}
            size="sm"
          />
        </div>

        <div className="border-t border-outline pt-3 space-y-2">
          {HOTKEY_ACTIONS.map((action) => {
            const isDuplicate = duplicates.has(action.key)
            const isSameAsModifier = hotkeys[action.key] === hotkeys.modifier
            const hasError = isDuplicate || isSameAsModifier
            return (
              <div key={action.key} className="flex items-center gap-3">
                <span
                  className={`text-sm min-w-[120px] ${hasError ? 'text-status-error' : 'text-on-surface'}`}
                >
                  {action.label}
                </span>
                <span className="text-xs text-on-surface-muted">{modifierLabel} +</span>
                <Select
                  value={hotkeys[action.key]}
                  options={SDL_BUTTONS.map((b) => ({ value: b.value, label: b.label }))}
                  onChange={(value) => onActionChange(action.key, value)}
                  size="sm"
                  error={hasError}
                />
                {isSameAsModifier && (
                  <span className="text-xs text-status-error">Same as modifier</span>
                )}
                {isDuplicate && !isSameAsModifier && (
                  <span className="text-xs text-status-error">Duplicate</span>
                )}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

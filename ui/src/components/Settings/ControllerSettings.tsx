import { RadioCard } from '@/lib/RadioCard'

export interface ControllerSettingsProps {
  readonly nintendoConfirm: string
  readonly onNintendoConfirmChange: (value: string) => void
}

type NintendoConfirmOption = 'east' | 'south'

const NINTENDO_CONFIRM_OPTIONS: {
  value: NintendoConfirmOption
  title: string
  description: string
}[] = [
  {
    value: 'east',
    title: 'East button confirms',
    description: 'Match the position of the original consoles.',
  },
  {
    value: 'south',
    title: 'South button confirms',
    description: 'Consistent with non-Nintendo consoles.',
  },
]

export function ControllerSettings({
  nintendoConfirm,
  onNintendoConfirmChange,
}: ControllerSettingsProps) {
  const selectedValue = nintendoConfirm === 'south' ? 'south' : 'east'

  return (
    <div>
      <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block">
        Controller
      </span>

      <p className="text-sm text-on-surface-muted mt-1 mb-2">
        Confirm button position for NES, SNES, Game Boy, GBA, DS, 3DS, GameCube, Wii, Wii U, and
        Switch.
      </p>

      <div className="mt-2 grid grid-cols-1 sm:grid-cols-2 gap-2">
        {NINTENDO_CONFIRM_OPTIONS.map((option) => (
          <RadioCard
            key={option.value}
            title={option.title}
            description={option.description}
            selected={selectedValue === option.value}
            onSelect={() => onNintendoConfirmChange(option.value)}
            className="flex-1 min-w-0 p-3"
          />
        ))}
      </div>
    </div>
  )
}

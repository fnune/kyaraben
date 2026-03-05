import { Modal } from '@/lib/Modal'
import type { EmulatorID, SystemID } from '@/types/daemon'

export interface EmulatorSettingsModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorId: EmulatorID
  readonly emulatorName: string
  readonly systemId: SystemID
  readonly supportsPreset: boolean
  readonly preset: string | null
  readonly graphics: { preset: string }
  readonly onPresetChange: (value: string | null) => void
  readonly supportsResume: boolean
  readonly resume: string | null
  readonly savestate: { resume: string }
  readonly onResumeChange: (value: string | null) => void
}

type PresetOption = 'modern-pixels' | 'upscaled' | 'pseudo-authentic' | 'manual' | 'default'
type ResumeOption = 'on' | 'off' | 'manual' | 'default'

function presetToOption(value: string | null): PresetOption {
  if (
    value === 'modern-pixels' ||
    value === 'upscaled' ||
    value === 'pseudo-authentic' ||
    value === 'manual'
  )
    return value
  return 'default'
}

function presetOptionToValue(option: PresetOption): string | null {
  return option === 'default' ? null : option
}

function resumeToOption(value: string | null): ResumeOption {
  if (value === 'on' || value === 'off' || value === 'manual') return value
  return 'default'
}

function resumeOptionToValue(option: ResumeOption): string | null {
  return option === 'default' ? null : option
}

function resolvePreset(override: string | null, global: string): string {
  if (override !== null) return override
  if (global) return global
  return 'manual'
}

function resolveResume(override: string | null, global: string): string {
  if (override !== null) return override
  if (global) return global
  return 'manual'
}

function formatPresetLabel(preset: string): string {
  switch (preset) {
    case 'modern-pixels':
      return 'Modern pixels'
    case 'upscaled':
      return 'Upscaled'
    case 'pseudo-authentic':
      return 'Pseudo-authentic'
    default:
      return 'Manual'
  }
}

export function EmulatorSettingsModal({
  open,
  onClose,
  emulatorName,
  supportsPreset,
  preset,
  graphics,
  onPresetChange,
  supportsResume,
  resume,
  savestate,
  onResumeChange,
}: EmulatorSettingsModalProps) {
  const currentPresetOption = presetToOption(preset)
  const resolvedPreset = resolvePreset(preset, graphics.preset)
  const currentResumeOption = resumeToOption(resume)
  const resolvedResume = resolveResume(resume, savestate.resume)

  const handlePresetOptionChange = (option: PresetOption) => {
    onPresetChange(presetOptionToValue(option))
  }

  const handleResumeOptionChange = (option: ResumeOption) => {
    onResumeChange(resumeOptionToValue(option))
  }

  const getPresetDefaultLabel = () => {
    if (!graphics.preset || graphics.preset === 'manual') return 'Default'
    return `Default (${formatPresetLabel(graphics.preset).toLowerCase()})`
  }

  const getPresetDescription = () => {
    if (resolvedPreset === 'manual') {
      return 'Kyaraben will not modify display settings.'
    }
    return `Kyaraben will apply the ${formatPresetLabel(resolvedPreset).toLowerCase()} preset.`
  }

  const getResumeDefaultLabel = () => {
    if (!savestate.resume || savestate.resume === 'manual') return 'Default'
    if (savestate.resume === 'recommended') return 'Default (recommended)'
    return `Default (${savestate.resume})`
  }

  const getResumeDescription = () => {
    if (resolvedResume === 'manual') {
      return 'Kyaraben will not modify resume settings.'
    }
    if (resolvedResume === 'on' || resolvedResume === 'recommended') {
      return 'Kyaraben will enable auto-resume (save on exit, load on launch).'
    }
    return 'Kyaraben will disable auto-resume.'
  }

  return (
    <Modal open={open} onClose={onClose} title={`${emulatorName} settings`}>
      <div className="space-y-4">
        {supportsPreset && (
          <div>
            <p className="text-sm text-on-surface-muted mb-2">Display preset</p>
            <div className="flex flex-wrap gap-2">
              <PresetButton
                label="Modern pixels"
                selected={currentPresetOption === 'modern-pixels'}
                onClick={() => handlePresetOptionChange('modern-pixels')}
              />
              <PresetButton
                label="Upscaled"
                selected={currentPresetOption === 'upscaled'}
                onClick={() => handlePresetOptionChange('upscaled')}
              />
              <PresetButton
                label="Pseudo-authentic"
                selected={currentPresetOption === 'pseudo-authentic'}
                onClick={() => handlePresetOptionChange('pseudo-authentic')}
              />
              <PresetButton
                label="Manual"
                selected={currentPresetOption === 'manual'}
                onClick={() => handlePresetOptionChange('manual')}
              />
              <PresetButton
                label={getPresetDefaultLabel()}
                selected={currentPresetOption === 'default'}
                onClick={() => handlePresetOptionChange('default')}
              />
            </div>
            <p className="text-xs text-on-surface-dim mt-2">{getPresetDescription()}</p>
          </div>
        )}
        {supportsResume && (
          <div>
            <p className="text-sm text-on-surface-muted mb-2">Resume</p>
            <div className="flex rounded-element overflow-hidden border border-outline">
              <SegmentedButton
                label="On"
                selected={currentResumeOption === 'on'}
                onClick={() => handleResumeOptionChange('on')}
              />
              <SegmentedButton
                label="Off"
                selected={currentResumeOption === 'off'}
                onClick={() => handleResumeOptionChange('off')}
              />
              <SegmentedButton
                label="Manual"
                selected={currentResumeOption === 'manual'}
                onClick={() => handleResumeOptionChange('manual')}
              />
              <SegmentedButton
                label={getResumeDefaultLabel()}
                selected={currentResumeOption === 'default'}
                onClick={() => handleResumeOptionChange('default')}
              />
            </div>
            <p className="text-xs text-on-surface-dim mt-2">{getResumeDescription()}</p>
          </div>
        )}
      </div>
    </Modal>
  )
}

function PresetButton({
  label,
  selected,
  onClick,
}: {
  readonly label: string
  readonly selected: boolean
  readonly onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`px-3 py-1.5 text-sm rounded-element border transition-colors ${
        selected
          ? 'bg-accent text-on-accent border-accent'
          : 'bg-surface-raised text-on-surface-secondary border-outline hover:bg-surface-raised/80'
      }`}
    >
      {label}
    </button>
  )
}

function SegmentedButton({
  label,
  selected,
  onClick,
}: {
  readonly label: string
  readonly selected: boolean
  readonly onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex-1 px-4 py-2 text-sm transition-colors ${
        selected
          ? 'bg-accent text-on-accent'
          : 'bg-surface-raised text-on-surface-secondary hover:bg-surface-raised/80'
      }`}
    >
      {label}
    </button>
  )
}

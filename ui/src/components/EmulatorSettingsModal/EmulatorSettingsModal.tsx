import type { EmulatorID, SystemID } from '@shared/daemon'
import { Modal } from '@/lib/Modal'

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

type PresetOption = 'clean' | 'retro' | 'manual' | 'default'
type ResumeOption = 'autosave' | 'autoload' | 'off' | 'manual' | 'default'

function presetToOption(value: string | null): PresetOption {
  if (value === 'clean' || value === 'retro' || value === 'manual') return value
  return 'default'
}

function presetOptionToValue(option: PresetOption): string | null {
  return option === 'default' ? null : option
}

function resumeToOption(value: string | null): ResumeOption {
  if (value === 'autosave' || value === 'autoload' || value === 'off' || value === 'manual')
    return value
  // Overrides written before the autosave/autoload split use "on" for both.
  if (value === 'on') return 'autosave'
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
  const value = override ?? global
  if (value === 'autoload' || value === 'off' || value === 'manual') return value
  return 'autosave'
}

function formatResumeLabel(resume: string): string {
  switch (resume) {
    case 'autoload':
      return 'Autoload'
    case 'off':
      return 'Off'
    case 'manual':
      return 'Manual'
    default:
      return 'Autosave'
  }
}

function formatPresetLabel(preset: string): string {
  switch (preset) {
    case 'clean':
      return 'Clean'
    case 'retro':
      return 'Retro'
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
      return 'Your display settings will be preserved.'
    }
    return `Kyaraben will apply the ${formatPresetLabel(resolvedPreset).toLowerCase()} preset.`
  }

  const getResumeDefaultLabel = () => {
    if (savestate.resume === 'manual') return 'Default'
    return `Default (${formatResumeLabel(savestate.resume).toLowerCase()})`
  }

  const getResumeDescription = () => {
    switch (resolvedResume) {
      case 'manual':
        return 'Kyaraben will not modify resume settings.'
      case 'off':
        return 'Kyaraben will disable auto-resume.'
      case 'autoload':
        return 'Kyaraben will savestate on exit and load it back on launch.'
      default:
        return 'Kyaraben will savestate on exit, but not load it back on launch.'
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={`${emulatorName} settings`}>
      <div className="space-y-4">
        {supportsPreset && (
          <div>
            <p className="text-sm text-on-surface-muted mb-2">Display preset</p>
            <div className="flex flex-wrap gap-2">
              <PresetButton
                label="Retro"
                selected={currentPresetOption === 'retro'}
                onClick={() => handlePresetOptionChange('retro')}
              />
              <PresetButton
                label="Clean"
                selected={currentPresetOption === 'clean'}
                onClick={() => handlePresetOptionChange('clean')}
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
                label="Autosave"
                selected={currentResumeOption === 'autosave'}
                onClick={() => handleResumeOptionChange('autosave')}
              />
              <SegmentedButton
                label="Autoload"
                selected={currentResumeOption === 'autoload'}
                onClick={() => handleResumeOptionChange('autoload')}
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

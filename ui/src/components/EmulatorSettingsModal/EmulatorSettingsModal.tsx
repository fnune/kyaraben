import { Modal } from '@/lib/Modal'
import type { EmulatorID, SystemID } from '@/types/daemon'

export interface EmulatorSettingsModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorId: EmulatorID
  readonly emulatorName: string
  readonly systemId: SystemID
  readonly supportsShaders: boolean
  readonly shaders: string | null
  readonly graphics: { shaders: string }
  readonly onShaderChange: (value: string | null) => void
  readonly supportsResume: boolean
  readonly resume: string | null
  readonly savestate: { resume: string }
  readonly onResumeChange: (value: string | null) => void
}

type SettingOption = 'on' | 'off' | 'manual' | 'default'

function valueToOption(value: string | null): SettingOption {
  if (value === 'on' || value === 'off' || value === 'manual') return value
  return 'default'
}

function optionToValue(option: SettingOption): string | null {
  return option === 'default' ? null : option
}

function resolveSetting(override: string | null, global: string): string {
  if (override !== null) return override
  if (global) return global
  return 'manual'
}

const LCD_SYSTEMS: readonly SystemID[] = [
  'gb',
  'gbc',
  'gba',
  'nds',
  'n3ds',
  'ngp',
  'gamegear',
  'psp',
  'psvita',
  'switch',
]

function getDisplayType(systemId: SystemID): 'crt' | 'lcd' {
  return LCD_SYSTEMS.includes(systemId) ? 'lcd' : 'crt'
}

function getShaderInfo(emulatorId: EmulatorID, displayType: 'crt' | 'lcd'): string {
  if (emulatorId === 'dolphin') {
    return 'CRT shader (crt-lottes-fast)'
  }
  return displayType === 'crt' ? 'CRT shader (crt-mattias)' : 'LCD shader (lcd-grid-v2)'
}

export function EmulatorSettingsModal({
  open,
  onClose,
  emulatorId,
  emulatorName,
  systemId,
  supportsShaders,
  shaders,
  graphics,
  onShaderChange,
  supportsResume,
  resume,
  savestate,
  onResumeChange,
}: EmulatorSettingsModalProps) {
  const displayType = getDisplayType(systemId)
  const currentShaderOption = valueToOption(shaders)
  const resolvedShaders = resolveSetting(shaders, graphics.shaders)
  const currentResumeOption = valueToOption(resume)
  const resolvedResume = resolveSetting(resume, savestate.resume)

  const handleShaderOptionChange = (option: SettingOption) => {
    onShaderChange(optionToValue(option))
  }

  const handleResumeOptionChange = (option: SettingOption) => {
    onResumeChange(optionToValue(option))
  }

  const getShaderDefaultLabel = () => {
    if (!graphics.shaders || graphics.shaders === 'manual') return 'Default'
    if (graphics.shaders === 'recommended') return 'Default (recommended)'
    return `Default (${graphics.shaders})`
  }

  const getShaderDescription = () => {
    if (resolvedShaders === 'manual') {
      return 'Kyaraben will not modify shader settings.'
    }
    if (resolvedShaders === 'on' || resolvedShaders === 'recommended') {
      return <>Kyaraben will enable: {getShaderInfo(emulatorId, displayType)}.</>
    }
    return 'Kyaraben will disable shaders.'
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
    <Modal open={open} onClose={onClose} title={`${emulatorName} Settings`}>
      <div className="space-y-4">
        {supportsShaders && (
          <div>
            <p className="text-sm text-on-surface-muted mb-2">Shaders</p>
            <div className="flex rounded-element overflow-hidden border border-outline">
              <SegmentedButton
                label="On"
                selected={currentShaderOption === 'on'}
                onClick={() => handleShaderOptionChange('on')}
              />
              <SegmentedButton
                label="Off"
                selected={currentShaderOption === 'off'}
                onClick={() => handleShaderOptionChange('off')}
              />
              <SegmentedButton
                label="Manual"
                selected={currentShaderOption === 'manual'}
                onClick={() => handleShaderOptionChange('manual')}
              />
              <SegmentedButton
                label={getShaderDefaultLabel()}
                selected={currentShaderOption === 'default'}
                onClick={() => handleShaderOptionChange('default')}
              />
            </div>
            <p className="text-xs text-on-surface-dim mt-2">{getShaderDescription()}</p>
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

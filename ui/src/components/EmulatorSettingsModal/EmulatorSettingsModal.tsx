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
}

type ShaderOption = 'on' | 'off' | 'manual' | 'default'

function shaderValueToOption(value: string | null): ShaderOption {
  if (value === 'on') return 'on'
  if (value === 'off') return 'off'
  if (value === 'manual') return 'manual'
  return 'default'
}

function optionToShaderValue(option: ShaderOption): string | null {
  if (option === 'on') return 'on'
  if (option === 'off') return 'off'
  if (option === 'manual') return 'manual'
  return null
}

function resolveShaders(shaders: string | null, graphics: { shaders: string }): string {
  if (shaders !== null) return shaders
  if (graphics.shaders) return graphics.shaders
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
}: EmulatorSettingsModalProps) {
  const displayType = getDisplayType(systemId)
  const currentOption = shaderValueToOption(shaders)
  const resolvedShaders = resolveShaders(shaders, graphics)

  const handleOptionChange = (option: ShaderOption) => {
    onShaderChange(optionToShaderValue(option))
  }

  const getDefaultLabel = () => {
    if (!graphics.shaders || graphics.shaders === 'manual') return 'Default'
    if (graphics.shaders === 'recommended') return 'Default (recommended)'
    return `Default (${graphics.shaders})`
  }

  const getDescription = () => {
    if (resolvedShaders === 'manual') {
      return 'Kyaraben will not modify shader settings.'
    }
    if (resolvedShaders === 'on' || resolvedShaders === 'recommended') {
      return <>Kyaraben will enable: {getShaderInfo(emulatorId, displayType)}.</>
    }
    return 'Kyaraben will disable shaders.'
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
                selected={currentOption === 'on'}
                onClick={() => handleOptionChange('on')}
              />
              <SegmentedButton
                label="Off"
                selected={currentOption === 'off'}
                onClick={() => handleOptionChange('off')}
              />
              <SegmentedButton
                label="Manual"
                selected={currentOption === 'manual'}
                onClick={() => handleOptionChange('manual')}
              />
              <SegmentedButton
                label={getDefaultLabel()}
                selected={currentOption === 'default'}
                onClick={() => handleOptionChange('default')}
              />
            </div>
            <p className="text-xs text-on-surface-dim mt-2">{getDescription()}</p>
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

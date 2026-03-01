import { Modal } from '@/lib/Modal'
import type { EmulatorID, SystemID } from '@/types/daemon'

export interface EmulatorSettingsModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly emulatorId: EmulatorID
  readonly emulatorName: string
  readonly systemId: SystemID
  readonly supportsShaders: boolean
  readonly shaders: boolean | null
  readonly onShaderChange: (value: boolean | null) => void
}

type ShaderOption = 'on' | 'off' | 'manual'

function shaderValueToOption(value: boolean | null): ShaderOption {
  if (value === true) return 'on'
  if (value === false) return 'off'
  return 'manual'
}

function optionToShaderValue(option: ShaderOption): boolean | null {
  if (option === 'on') return true
  if (option === 'off') return false
  return null
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
  onShaderChange,
}: EmulatorSettingsModalProps) {
  const displayType = getDisplayType(systemId)
  const currentOption = shaderValueToOption(shaders)

  const handleOptionChange = (option: ShaderOption) => {
    onShaderChange(optionToShaderValue(option))
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
            </div>
            <p className="text-xs text-on-surface-dim mt-2">
              {currentOption === 'manual' ? (
                'Kyaraben will not modify shader settings.'
              ) : currentOption === 'on' ? (
                <>Kyaraben will enable: {getShaderInfo(emulatorId, displayType)}.</>
              ) : (
                'Kyaraben will disable shaders.'
              )}
            </p>
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

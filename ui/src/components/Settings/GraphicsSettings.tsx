import { RadioCard } from '@/lib/RadioCard'

export interface GraphicsSettingsProps {
  readonly shaders: string
  readonly onShadersChange: (value: string) => void
}

type ShaderOption = 'on' | 'off' | 'manual'

const SHADER_OPTIONS: { value: ShaderOption; title: string; description: string }[] = [
  { value: 'on', title: 'Shaders on', description: 'CRT for consoles, LCD for handhelds.' },
  { value: 'off', title: 'Shaders off', description: 'Disable shaders.' },
  { value: 'manual', title: 'Shaders manual', description: 'Configure shaders yourself.' },
]

export function GraphicsSettings({ shaders, onShadersChange }: GraphicsSettingsProps) {
  const selectedValue = shaders === 'on' || shaders === 'off' ? shaders : 'manual'

  return (
    <div>
      <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block">
        Graphics
      </span>

      <div className="mt-2 grid grid-cols-1 sm:grid-cols-3 gap-2">
        {SHADER_OPTIONS.map((option) => (
          <RadioCard
            key={option.value}
            title={option.title}
            description={option.description}
            selected={selectedValue === option.value}
            onSelect={() => onShadersChange(option.value)}
            className="flex-1 min-w-0 p-3"
          />
        ))}
      </div>
    </div>
  )
}

import { RadioCard } from '@/lib/RadioCard'

export interface SavestateSettingsProps {
  readonly resume: string
  readonly onResumeChange: (value: string) => void
}

type ResumeOption = 'recommended' | 'off' | 'manual'

const RESUME_OPTIONS: { value: ResumeOption; title: string; description: string }[] = [
  {
    value: 'recommended',
    title: 'Resume recommended',
    description: 'Autosave on exit, autoload on launch.',
  },
  { value: 'off', title: 'Resume off', description: 'Disable auto-resume.' },
  { value: 'manual', title: 'Resume manual', description: 'Configure yourself.' },
]

export function SavestateSettings({ resume, onResumeChange }: SavestateSettingsProps) {
  const selectedValue = resume === 'recommended' || resume === 'off' ? resume : 'manual'

  return (
    <div>
      <span className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest block">
        Savestates
      </span>

      <div className="mt-2 grid grid-cols-1 sm:grid-cols-3 gap-2">
        {RESUME_OPTIONS.map((option) => (
          <RadioCard
            key={option.value}
            title={option.title}
            description={option.description}
            selected={selectedValue === option.value}
            onSelect={() => onResumeChange(option.value)}
            className="flex-1 min-w-0 p-3"
          />
        ))}
      </div>
    </div>
  )
}

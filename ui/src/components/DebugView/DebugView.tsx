import { SYSTEM_LOGOS } from '@/components/SystemLogo/SystemLogo'
import { fontPairs } from '@/lib/fonts'
import { useTheme } from '@/lib/ThemeContext'
import { themes } from '@/lib/themes'
import type { SystemID } from '@/types/daemon'

const SAMPLE_SYSTEMS: { id: SystemID; name: string }[] = [
  { id: 'snes', name: 'Super Nintendo' },
  { id: 'psx', name: 'PlayStation' },
  { id: 'genesis', name: 'Mega Drive' },
  { id: 'n64', name: 'Nintendo 64' },
  { id: 'dreamcast', name: 'Dreamcast' },
  { id: 'gba', name: 'Game Boy Advance' },
]

export function DebugView() {
  const { theme, fontPair, setThemeId, setFontPairId } = useTheme()

  return (
    <div className="p-6 space-y-8">
      <div>
        <h2 className="font-heading text-lg font-semibold text-on-surface tracking-wide uppercase">
          Design exploration
        </h2>
        <p className="text-sm text-on-surface-muted mt-1 italic">
          Switch themes and font pairs to preview different design directions.
        </p>
      </div>

      <div className="grid grid-cols-1 min-[720px]:grid-cols-2 gap-6">
        <fieldset>
          <legend className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest mb-3">
            Color theme
          </legend>
          <div className="space-y-2">
            {themes.map((t) => (
              <label
                key={t.id}
                className={`flex items-center gap-3 p-3 rounded-element border cursor-pointer transition-colors ${
                  theme.id === t.id
                    ? 'border-accent bg-accent-muted'
                    : 'border-outline hover:border-outline-strong'
                }`}
              >
                <input
                  type="radio"
                  name="theme"
                  value={t.id}
                  checked={theme.id === t.id}
                  onChange={() => setThemeId(t.id)}
                  className="sr-only"
                />
                <div className="flex gap-1 shrink-0">
                  <span
                    className="w-4 h-4 rounded-full border border-outline"
                    style={{ backgroundColor: t.tokens.surface }}
                  />
                  <span
                    className="w-4 h-4 rounded-full border border-outline"
                    style={{ backgroundColor: t.tokens.surfaceAlt }}
                  />
                  <span
                    className="w-4 h-4 rounded-full border border-outline"
                    style={{ backgroundColor: t.tokens.accent }}
                  />
                </div>
                <div className="flex-1 min-w-0">
                  <span className="text-sm font-medium text-on-surface">{t.name}</span>
                  <span className="text-xs text-on-surface-muted ml-2">{t.description}</span>
                </div>
              </label>
            ))}
          </div>
        </fieldset>

        <fieldset>
          <legend className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest mb-3">
            Font pair
          </legend>
          <div className="space-y-2">
            {fontPairs.map((f) => (
              <label
                key={f.id}
                className={`flex items-center gap-3 p-3 rounded-element border cursor-pointer transition-colors ${
                  fontPair.id === f.id
                    ? 'border-accent bg-accent-muted'
                    : 'border-outline hover:border-outline-strong'
                }`}
              >
                <input
                  type="radio"
                  name="fontPair"
                  value={f.id}
                  checked={fontPair.id === f.id}
                  onChange={() => setFontPairId(f.id)}
                  className="sr-only"
                />
                <div className="flex-1 min-w-0">
                  <span className="text-sm font-medium text-on-surface">{f.name}</span>
                  <span className="text-xs text-on-surface-muted ml-2">{f.description}</span>
                </div>
              </label>
            ))}
          </div>
        </fieldset>
      </div>

      <div>
        <h3 className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest mb-4">
          Logo grid preview
        </h3>
        <div className="grid grid-cols-3 min-[720px]:grid-cols-6 gap-4">
          {SAMPLE_SYSTEMS.map((sys) => (
            <div
              key={sys.id}
              className="flex flex-col items-center gap-2 p-4 rounded-element bg-surface-alt border border-outline"
            >
              <div className="w-16 h-10 flex items-center justify-center">
                <img
                  src={SYSTEM_LOGOS[sys.id]}
                  alt={sys.name}
                  className="max-w-full max-h-full object-contain"
                  style={{ filter: 'var(--t-logo-filter)' }}
                />
              </div>
              <span className="text-xs text-on-surface-muted text-center font-mono uppercase tracking-wider">
                {sys.id}
              </span>
            </div>
          ))}
        </div>
      </div>

      <div>
        <h3 className="text-xs font-semibold text-on-surface-dim uppercase tracking-widest mb-4">
          Typography specimen
        </h3>
        <div className="space-y-4 p-6 rounded-element bg-surface-alt border border-outline">
          <h4 className="font-heading text-2xl font-semibold text-on-surface">Kyaraben</h4>
          <p className="font-heading text-lg text-on-surface-secondary italic">
            Emulation, managed.
          </p>
          <p className="text-sm text-on-surface-muted leading-relaxed">
            Configure your systems, install emulators, and synchronize saves across devices.
            Everything stays under your control.
          </p>
          <div className="flex gap-3 pt-2">
            <button
              type="button"
              className="px-4 py-2 rounded-control bg-accent text-white text-sm font-medium"
            >
              Apply
            </button>
            <button
              type="button"
              className="px-4 py-2 rounded-control bg-surface-raised text-on-surface-secondary text-sm"
            >
              Discard
            </button>
          </div>
          <div className="pt-4 border-t border-outline">
            <p className="font-mono text-xs text-on-surface-dim">
              MONO: 0123456789 ~/Emulation v0.1.0
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

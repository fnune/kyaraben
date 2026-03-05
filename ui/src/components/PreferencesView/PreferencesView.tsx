import { useMemo } from 'react'
import { HotkeySettings } from '@/components/Settings/HotkeySettings'
import { useConfig } from '@/lib/ConfigContext'
import { RadioCard } from '@/lib/RadioCard'
import type { EmulatorRef, System } from '@/types/daemon'
import { PreferenceSection } from './PreferenceSection'

type ShaderOption = 'recommended' | 'off' | 'manual'
type ResumeOption = 'recommended' | 'off' | 'manual'
type ConfirmOption = 'east' | 'south'

const ALL_HOTKEYS = [
  'savestate',
  'loadstate',
  'nextslot',
  'prevslot',
  'fastforward',
  'rewind',
  'pause',
  'screenshot',
  'quit',
  'fullscreen',
  'menu',
] as const

type HotkeyID = (typeof ALL_HOTKEYS)[number]

const HOTKEY_LABELS: Record<HotkeyID, string> = {
  savestate: 'Save',
  loadstate: 'Load',
  nextslot: 'Next slot',
  prevslot: 'Prev slot',
  fastforward: 'Fast-forward',
  rewind: 'Rewind',
  pause: 'Pause',
  screenshot: 'Screenshot',
  quit: 'Quit',
  fullscreen: 'Full-screen',
  menu: 'Menu',
}

interface EmulatorHotkeyInfo {
  name: string
  supportedHotkeys: Set<string>
  isRetroArch: boolean
}

function byCountDescThenName<T>(getCount: (item: T) => number, getName: (item: T) => string) {
  return (a: T, b: T) => getCount(b) - getCount(a) || getName(a).localeCompare(getName(b))
}

function getEmulatorHotkeyInfo(systems: readonly System[]): EmulatorHotkeyInfo[] {
  const seen = new Set<string>()
  const emulators: EmulatorHotkeyInfo[] = []
  const retroArchEmulators: EmulatorRef[] = []

  for (const system of systems) {
    for (const emu of system.emulators) {
      if (seen.has(emu.id)) continue
      seen.add(emu.id)

      const isRetroArch = emu.name.includes('RetroArch')
      if (isRetroArch) {
        retroArchEmulators.push(emu)
      } else {
        emulators.push({
          name: emu.name,
          supportedHotkeys: new Set(emu.supportedHotkeys ?? []),
          isRetroArch: false,
        })
      }
    }
  }

  if (retroArchEmulators.length > 0) {
    const firstRA = retroArchEmulators[0]
    emulators.unshift({
      name: 'RetroArch cores',
      supportedHotkeys: new Set(firstRA?.supportedHotkeys ?? []),
      isRetroArch: true,
    })
  }

  emulators.sort(
    byCountDescThenName(
      (e) => e.supportedHotkeys.size,
      (e) => e.name,
    ),
  )
  return emulators
}

function getResumeEmulatorNames(systems: readonly System[]): string[] {
  const seen = new Set<string>()
  const names: string[] = []
  let hasRetroArch = false

  for (const system of systems) {
    for (const emu of system.emulators) {
      if (seen.has(emu.id)) continue
      seen.add(emu.id)

      if (!emu.resumeRecommended) continue

      const isRetroArch = emu.name.includes('RetroArch')
      if (isRetroArch) {
        hasRetroArch = true
      } else {
        names.push(emu.name)
      }
    }
  }

  const result: string[] = []
  if (hasRetroArch) {
    result.push('RetroArch cores')
  }
  result.push(...names.sort())
  return result
}

function getNintendoDiamondSystemNames(systems: readonly System[]): string[] {
  return systems.filter((s) => s.nintendoDiamond).map((s) => s.name)
}

function getSortedHotkeys(emulatorInfo: EmulatorHotkeyInfo[]): HotkeyID[] {
  const supportCount = new Map<HotkeyID, number>()
  for (const id of ALL_HOTKEYS) {
    supportCount.set(id, emulatorInfo.filter((e) => e.supportedHotkeys.has(id)).length)
  }
  return [...ALL_HOTKEYS].sort(
    byCountDescThenName(
      (id) => supportCount.get(id) ?? 0,
      (id) => id,
    ),
  )
}

export function PreferencesView() {
  const config = useConfig()
  const {
    configState,
    systems,
    setGraphicsShaders,
    setSavestateResume,
    setControllerNintendoConfirm,
    setHotkeyModifier,
    setHotkeyAction,
    resetHotkeys,
  } = config

  const shaders = configState.graphicsShaders
  const resume = configState.savestateResume
  const nintendoConfirm = configState.controllerNintendoConfirm
  const hotkeys = configState.hotkeys

  const emulatorHotkeyInfo = useMemo(() => getEmulatorHotkeyInfo(systems), [systems])
  const resumeEmulatorNames = useMemo(() => getResumeEmulatorNames(systems), [systems])
  const nintendoDiamondSystemNames = useMemo(
    () => getNintendoDiamondSystemNames(systems),
    [systems],
  )
  const sortedHotkeys = useMemo(() => getSortedHotkeys(emulatorHotkeyInfo), [emulatorHotkeyInfo])

  const selectedShaders: ShaderOption =
    shaders === 'recommended' || shaders === 'off' ? shaders : 'manual'
  const selectedResume: ResumeOption =
    resume === 'recommended' || resume === 'off' ? resume : 'manual'
  const selectedConfirm: ConfirmOption = nintendoConfirm === 'south' ? 'south' : 'east'

  return (
    <div className="p-6 pb-24">
      <PreferenceSection
        title="Display"
        intro={
          <>
            <p className="text-sm text-on-surface mb-4">
              Shaders add visual effects that mimic original display hardware. Without shaders, you
              see raw pixels scaled up, which can look harsh on modern screens.
            </p>
            <div className="grid grid-cols-3 gap-3">
              <div className="text-center">
                <img
                  src="https://placehold.co/200x150?text=No+shader"
                  alt="Game without shaders showing raw pixels"
                  className="w-full rounded border border-outline mb-2"
                />
                <span className="text-xs text-on-surface-muted">No shader</span>
              </div>
              <div className="text-center">
                <img
                  src="https://placehold.co/200x150?text=CRT+shader"
                  alt="Game with CRT shader showing scanlines and bloom"
                  className="w-full rounded border border-outline mb-2"
                />
                <span className="text-xs text-on-surface-muted">CRT shader</span>
              </div>
              <div className="text-center">
                <img
                  src="https://placehold.co/200x150?text=LCD+shader"
                  alt="Game with LCD shader showing smoothed pixels"
                  className="w-full rounded border border-outline mb-2"
                />
                <span className="text-xs text-on-surface-muted">LCD shader</span>
              </div>
            </div>
          </>
        }
        controls={
          <>
            <RadioCard
              title="Recommended"
              description="CRT shaders for home consoles. LCD shaders for handhelds. Kyaraben picks the right shader for each system."
              selected={selectedShaders === 'recommended'}
              onSelect={() => setGraphicsShaders('recommended')}
              className="w-full p-3"
              wrap
            />
            <RadioCard
              title="Off"
              description="No shaders. Games display with raw pixels scaled to your screen."
              selected={selectedShaders === 'off'}
              onSelect={() => setGraphicsShaders('off')}
              className="w-full p-3"
              wrap
            />
            <RadioCard
              title="Manual"
              description="Kyaraben won't configure shaders. Set them up yourself in each emulator."
              selected={selectedShaders === 'manual'}
              onSelect={() => setGraphicsShaders('manual')}
              className="w-full p-3"
              wrap
            />
          </>
        }
        support={
          <p className="text-xs text-on-surface-muted">Emulator support details coming soon.</p>
        }
      />

      <PreferenceSection
        title="Resume"
        intro={
          <>
            <p className="text-sm text-on-surface mb-4">
              Auto-resume uses savestates to pick up exactly where you left off. When you quit a
              game, Kyaraben creates a savestate. When you launch it again, that savestate loads
              automatically and you continue from the exact moment you stopped.
            </p>
            <p className="text-sm text-on-surface-muted mb-4">
              This is separate from in-game saves. Savestates capture the entire emulator state,
              including unsaved progress.
            </p>
            <div className="flex items-center justify-center gap-4 py-4">
              <div className="text-center px-4 py-3 bg-surface rounded border border-outline">
                <div className="text-2xl mb-1">🎮</div>
                <span className="text-xs text-on-surface-muted">Playing</span>
              </div>
              <div className="text-on-surface-muted">→</div>
              <div className="text-center px-4 py-3 bg-surface rounded border border-outline">
                <div className="text-2xl mb-1">💾</div>
                <span className="text-xs text-on-surface-muted">Quit (autosave)</span>
              </div>
              <div className="text-on-surface-muted">→</div>
              <div className="text-center px-4 py-3 bg-surface rounded border border-outline">
                <div className="text-2xl mb-1">▶️</div>
                <span className="text-xs text-on-surface-muted">Launch (autoload)</span>
              </div>
            </div>
          </>
        }
        controls={
          <>
            <RadioCard
              title="Recommended"
              description="Autosave when you quit. Autoload when you launch. Never lose progress."
              selected={selectedResume === 'recommended'}
              onSelect={() => setSavestateResume('recommended')}
              className="w-full p-3"
              wrap
            />
            <RadioCard
              title="Off"
              description="No automatic savestates. Rely on in-game saves only."
              selected={selectedResume === 'off'}
              onSelect={() => setSavestateResume('off')}
              className="w-full p-3"
              wrap
            />
            <RadioCard
              title="Manual"
              description="Kyaraben won't configure auto-resume. Set it up yourself in each emulator."
              selected={selectedResume === 'manual'}
              onSelect={() => setSavestateResume('manual')}
              className="w-full p-3"
              wrap
            />
          </>
        }
        support={
          resumeEmulatorNames.length > 0 ? (
            <>
              <p className="text-xs text-on-surface-muted mb-2">
                Supported emulators when set to recommended:
              </p>
              <p className="text-xs text-on-surface-muted">{resumeEmulatorNames.join(', ')}.</p>
            </>
          ) : undefined
        }
      />

      <PreferenceSection
        title="Controller"
        intro={
          <>
            <p className="text-sm text-on-surface mb-4">
              Nintendo puts the confirm button on the east (right) side of the face buttons.
              PlayStation and Xbox put it on the south (bottom). If you use an Xbox or PlayStation
              controller for Nintendo games, choose which convention feels right to you.
            </p>
            <div className="grid grid-cols-2 gap-6 py-2">
              <div className="text-center">
                <img
                  src="https://placehold.co/180x120?text=East+confirms"
                  alt="Controller diagram showing east button highlighted as confirm"
                  className="w-full max-w-[180px] mx-auto rounded border border-outline mb-2"
                />
                <span className="text-xs text-on-surface-muted">
                  East confirms (matches Nintendo layout)
                </span>
              </div>
              <div className="text-center">
                <img
                  src="https://placehold.co/180x120?text=South+confirms"
                  alt="Controller diagram showing south button highlighted as confirm"
                  className="w-full max-w-[180px] mx-auto rounded border border-outline mb-2"
                />
                <span className="text-xs text-on-surface-muted">
                  South confirms (matches Xbox/PlayStation)
                </span>
              </div>
            </div>
          </>
        }
        controls={
          <>
            <RadioCard
              title="East button confirms"
              description="Matches original Nintendo consoles. On an Xbox controller, B confirms."
              selected={selectedConfirm === 'east'}
              onSelect={() => setControllerNintendoConfirm('east')}
              className="w-full p-3"
              wrap
            />
            <RadioCard
              title="South button confirms"
              description="Consistent with PlayStation and Xbox. On an Xbox controller, A confirms."
              selected={selectedConfirm === 'south'}
              onSelect={() => setControllerNintendoConfirm('south')}
              className="w-full p-3"
              wrap
            />
          </>
        }
        support={
          nintendoDiamondSystemNames.length > 0 ? (
            <>
              <p className="text-xs text-on-surface-muted mb-2">
                Affects: {nintendoDiamondSystemNames.join(', ')}.
              </p>
              <p className="text-xs text-on-surface-muted">
                N64 is not affected because its controller has a unique vertical A/B layout rather
                than a diamond.
              </p>
            </>
          ) : undefined
        }
      />

      <PreferenceSection
        title="Hotkeys"
        headerAction={
          <button
            type="button"
            onClick={resetHotkeys}
            className="text-sm text-accent hover:text-accent-hover"
          >
            Reset defaults
          </button>
        }
        intro={
          <p className="text-sm text-on-surface">
            Controller hotkeys let you save states, load states, fast forward, and more without
            leaving your game. Hold a modifier button and press an action button to trigger the
            hotkey. Hotkeys use physical button positions and are not affected by the confirm button
            setting above.
          </p>
        }
        controls={
          <HotkeySettings
            hotkeys={hotkeys}
            onModifierChange={setHotkeyModifier}
            onActionChange={setHotkeyAction}
          />
        }
        support={
          emulatorHotkeyInfo.length > 0 ? (
            <>
              <p className="text-xs text-on-surface-muted mb-3">
                Not all emulators support all hotkeys. The table below shows which hotkeys work with
                each emulator.
              </p>
              <div className="overflow-x-auto">
                <table className="w-full text-xs text-on-surface-muted">
                  <thead>
                    <tr className="border-b border-outline">
                      <th className="text-left py-1.5 pr-3 font-medium">Emulator</th>
                      {sortedHotkeys.map((id) => (
                        <th key={id} className="px-1 py-1.5 font-medium text-center">
                          {HOTKEY_LABELS[id]}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {emulatorHotkeyInfo.map((emu) => (
                      <tr key={emu.name} className="border-b border-outline last:border-b-0">
                        <td className="py-1.5 pr-3">{emu.name}</td>
                        {sortedHotkeys.map((id) => (
                          <td key={id} className="px-1 py-1.5 text-center">
                            {emu.supportedHotkeys.has(id) ? (
                              <span className="text-status-success">&#10003;</span>
                            ) : (
                              <span className="text-on-surface-muted/30">-</span>
                            )}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          ) : undefined
        }
      />
    </div>
  )
}

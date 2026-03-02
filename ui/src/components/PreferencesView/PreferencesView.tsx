import { useConfig } from '@/lib/ConfigContext'
import { RadioCard } from '@/lib/RadioCard'

type ShaderOption = 'recommended' | 'off' | 'manual'
type ResumeOption = 'recommended' | 'off' | 'manual'
type ConfirmOption = 'east' | 'south'

export function PreferencesView() {
  const config = useConfig()
  const { configState, setGraphicsShaders, setSavestateResume, setControllerNintendoConfirm } =
    config

  const shaders = configState.graphicsShaders
  const resume = configState.savestateResume
  const nintendoConfirm = configState.controllerNintendoConfirm

  const selectedShaders: ShaderOption =
    shaders === 'recommended' || shaders === 'off' ? shaders : 'manual'
  const selectedResume: ResumeOption =
    resume === 'recommended' || resume === 'off' ? resume : 'manual'
  const selectedConfirm: ConfirmOption = nintendoConfirm === 'south' ? 'south' : 'east'

  return (
    <div className="p-6 pb-24">
      <section className="mb-10">
        <h2 className="text-sm font-semibold text-on-surface-dim uppercase tracking-widest mb-4">
          Display
        </h2>

        <div className="bg-surface-alt rounded-card border border-outline overflow-hidden">
          <div className="p-4 border-b border-outline">
            <p className="text-sm text-on-surface mb-4">
              Shaders add visual effects that mimic original display hardware. Without shaders, you
              see raw pixels scaled up, which can look harsh on modern screens.
            </p>

            <div className="grid grid-cols-3 gap-3 mb-4">
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
          </div>

          <div className="p-4 space-y-3">
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
          </div>
        </div>
      </section>

      <section className="mb-10">
        <h2 className="text-sm font-semibold text-on-surface-dim uppercase tracking-widest mb-4">
          Resume
        </h2>

        <div className="bg-surface-alt rounded-card border border-outline overflow-hidden">
          <div className="p-4 border-b border-outline">
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
          </div>

          <div className="p-4 space-y-3">
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
          </div>

          <div className="p-4 border-t border-outline">
            <p className="text-xs text-on-surface-muted mb-2">
              Supported emulators when set to recommended:
            </p>
            <p className="text-xs text-on-surface-muted">
              RetroArch cores (NES, SNES, Genesis, N64, Saturn, PC Engine, Neo Geo Pocket, GB, GBC,
              GBA, DS, 3DS, Arcade, Atari 2600, C64), DuckStation (PS1), Flycast (Dreamcast).
            </p>
          </div>
        </div>
      </section>

      <section className="mb-10">
        <h2 className="text-sm font-semibold text-on-surface-dim uppercase tracking-widest mb-4">
          Controller
        </h2>

        <div className="bg-surface-alt rounded-card border border-outline overflow-hidden">
          <div className="p-4 border-b border-outline">
            <p className="text-sm text-on-surface mb-4">
              Nintendo puts the confirm button on the east (right) side of the face buttons.
              PlayStation and Xbox put it on the south (bottom). If you use an Xbox or PlayStation
              controller for Nintendo games, choose which convention feels right to you.
            </p>
            <p className="text-sm text-on-surface-muted mb-4">
              Affects: NES, SNES, Game Boy, GBA, DS, 3DS, GameCube, Wii, Wii U, and Switch.
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
          </div>

          <div className="p-4 space-y-3">
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
          </div>
        </div>
      </section>
    </div>
  )
}

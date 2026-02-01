import { EmulatorSubcard } from '@/components/EmulatorSubcard/EmulatorSubcard'
import { SYSTEM_LOGOS } from '@/components/SystemLogo/SystemLogo'
import type { DoctorResponse, EmulatorID, System, SystemID } from '@/types/daemon'

const SYSTEM_YEARS: Record<SystemID, number> = {
  nes: 1983,
  snes: 1990,
  n64: 1996,
  gb: 1989,
  gbc: 1998,
  gba: 2001,
  nds: 2004,
  '3ds': 2011,
  gamecube: 2001,
  wii: 2006,
  wiiu: 2012,
  switch: 2017,
  psx: 1994,
  ps2: 2000,
  ps3: 2006,
  psp: 2004,
  psvita: 2011,
  genesis: 1988,
  saturn: 1994,
  dreamcast: 1998,
}

export interface SystemCardProps {
  readonly system: System
  readonly enabledEmulators: ReadonlySet<EmulatorID>
  readonly emulatorVersions: ReadonlyMap<EmulatorID, string | null>
  readonly installedVersions: ReadonlyMap<EmulatorID, string>
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onEmulatorToggle: (emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
}

export function SystemCard({
  system,
  enabledEmulators,
  emulatorVersions,
  installedVersions,
  provisions,
  userStore,
  onEmulatorToggle,
  onVersionChange,
}: SystemCardProps) {
  const logo = SYSTEM_LOGOS[system.id]
  const year = SYSTEM_YEARS[system.id]

  return (
    <article className="border border-gray-200 rounded-xl overflow-hidden bg-white">
      <div className="relative h-20 bg-gradient-to-r from-gray-100 to-gray-50 overflow-hidden">
        <div className="absolute inset-y-0 left-0 flex flex-col justify-center px-5 z-10">
          <h3 className="text-lg font-semibold text-gray-900">{system.name}</h3>
          <p className="text-sm text-gray-500">
            {system.manufacturer} · {year}
          </p>
        </div>

        <div
          className="absolute -right-4 top-1/2 -translate-y-1/2 h-28 w-40 opacity-[0.08]"
          style={{
            maskImage: 'linear-gradient(to left, black 40%, transparent 100%)',
            WebkitMaskImage: 'linear-gradient(to left, black 40%, transparent 100%)',
          }}
        >
          <img
            src={logo}
            alt=""
            className="h-full w-full object-contain object-right"
            style={{
              filter: 'brightness(0)',
            }}
          />
        </div>
      </div>

      <div className="p-2 space-y-2">
        {system.emulators.map((emulator) => (
          <EmulatorSubcard
            key={emulator.id}
            emulator={emulator}
            systemId={system.id}
            enabled={enabledEmulators.has(emulator.id)}
            pinnedVersion={emulatorVersions.get(emulator.id) ?? null}
            installedVersion={installedVersions.get(emulator.id) ?? null}
            provisions={provisions[emulator.id] ?? []}
            userStore={userStore}
            onToggle={(enabled) => onEmulatorToggle(emulator.id, enabled)}
            onVersionChange={(version) => onVersionChange(emulator.id, version)}
          />
        ))}
      </div>
    </article>
  )
}

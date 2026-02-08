import { EmulatorSubcard } from '@/components/EmulatorSubcard/EmulatorSubcard'
import { SYSTEM_LOGOS } from '@/components/SystemLogo/SystemLogo'
import { launchEmulator } from '@/lib/daemon'
import type {
  DoctorResponse,
  EmulatorID,
  EmulatorPaths,
  ManagedConfigInfo,
  System,
  SystemID,
} from '@/types/daemon'

export const SYSTEM_YEARS: Record<SystemID, number> = {
  nes: 1983,
  snes: 1990,
  n64: 1996,
  gb: 1989,
  gbc: 1998,
  gba: 2001,
  nds: 2004,
  n3ds: 2011,
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
  readonly installedExecLines: ReadonlyMap<EmulatorID, string>
  readonly managedConfigs: ReadonlyMap<EmulatorID, ManagedConfigInfo[]>
  readonly installedPaths: ReadonlyMap<EmulatorID, Record<string, EmulatorPaths>>
  readonly provisions: DoctorResponse
  readonly sharedPackages: ReadonlySet<string>
  readonly onEmulatorToggle: (emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
}

export function SystemCard({
  system,
  enabledEmulators,
  emulatorVersions,
  installedVersions,
  installedExecLines,
  managedConfigs,
  installedPaths,
  provisions,
  sharedPackages,
  onEmulatorToggle,
  onVersionChange,
}: SystemCardProps) {
  const logo = SYSTEM_LOGOS[system.id]
  const year = SYSTEM_YEARS[system.id]

  return (
    <article className="border border-outline rounded-card border-t-2 border-t-accent overflow-hidden bg-surface">
      <div className="relative flex items-center h-14 bg-surface-alt">
        <img
          src={logo}
          alt=""
          className="absolute right-4 h-8 w-auto opacity-10"
          style={{ filter: 'var(--t-logo-filter)' }}
        />
        <div
          className="relative z-10 pl-4 pr-12"
          style={{
            background: 'linear-gradient(to right, var(--t-surface-alt) 70%, transparent)',
          }}
        >
          <h3 className="font-heading text-base font-semibold text-on-surface whitespace-nowrap">
            {system.name}
          </h3>
          <p className="text-xs text-on-surface-muted whitespace-nowrap italic">
            {system.manufacturer} · {year}
          </p>
        </div>
      </div>

      <div className="p-2 space-y-2 bg-surface">
        {system.emulators.map((emulator) => {
          const execLine = installedExecLines.get(emulator.id)
          const emuManagedConfigs = managedConfigs.get(emulator.id)
          const emuPaths = installedPaths.get(emulator.id)?.[system.id]
          const packageName = emulator.packageName ?? emulator.id
          const isSharedPackage = sharedPackages.has(packageName)
          return (
            <EmulatorSubcard
              key={emulator.id}
              emulator={emulator}
              enabled={enabledEmulators.has(emulator.id)}
              pinnedVersion={emulatorVersions.get(emulator.id) ?? null}
              installedVersion={installedVersions.get(emulator.id) ?? null}
              provisions={provisions[emulator.id] ?? []}
              sharedPackage={isSharedPackage}
              onToggle={(enabled) => onEmulatorToggle(emulator.id, enabled)}
              onVersionChange={(version) => onVersionChange(emulator.id, version)}
              {...(emuManagedConfigs && { managedConfigs: emuManagedConfigs })}
              {...(emuPaths && { paths: emuPaths })}
              {...(execLine && {
                execLine,
                onLaunch: () => launchEmulator(execLine),
              })}
            />
          )
        })}
      </div>
    </article>
  )
}

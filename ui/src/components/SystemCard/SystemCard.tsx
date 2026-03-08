import type {
  DoctorResponse,
  EmulatorID,
  EmulatorPaths,
  ManagedConfigInfo,
  System,
  SystemID,
} from '@shared/daemon'
import { VERSION_DEFAULT } from '@shared/ui'
import { forwardRef } from 'react'
import { ESDE_LOGOS } from '@/assets/esde'
import { EmulatorSubcard } from '@/components/EmulatorSubcard/EmulatorSubcard'
import { launchEmulator } from '@/lib/daemon'

export const SYSTEM_YEARS: Record<SystemID, number | null> = {
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
  mastersystem: 1985,
  gamegear: 1990,
  saturn: 1994,
  dreamcast: 1998,
  pcengine: 1987,
  ngp: 1998,
  neogeo: 1990,
  xbox: 2001,
  xbox360: 2005,
  atari2600: 1977,
  c64: 1982,
  arcade: null,
}

export const LOGO_OPACITIES: Partial<Record<SystemID, number>> = {
  wii: 0.35,
  wiiu: 0.35,
  psp: 0.18,
  ps2: 0.18,
  gba: 0.07,
  saturn: 0.13,
}

export interface SystemCardProps {
  readonly system: System
  readonly systemEnabledEmulators: ReadonlySet<EmulatorID>
  readonly globalEnabledEmulators: ReadonlySet<EmulatorID>
  readonly defaultEmulatorId: EmulatorID | null
  readonly emulatorVersions: ReadonlyMap<EmulatorID, string>
  readonly emulatorPresets: ReadonlyMap<EmulatorID, string | null>
  readonly emulatorResume: ReadonlyMap<EmulatorID, string | null>
  readonly graphics: { preset: string }
  readonly savestate: { resume: string }
  readonly installedVersions: ReadonlyMap<EmulatorID, string>
  readonly installedExecLines: ReadonlyMap<EmulatorID, string>
  readonly managedConfigs: ReadonlyMap<EmulatorID, ManagedConfigInfo[]>
  readonly installedPaths: ReadonlyMap<EmulatorID, Record<string, EmulatorPaths>>
  readonly provisions: DoctorResponse
  readonly sharedPackages: ReadonlySet<string>
  readonly onEmulatorToggle: (systemId: SystemID, emulatorId: EmulatorID, enabled: boolean) => void
  readonly onSetDefaultEmulator: (systemId: SystemID, emulatorId: EmulatorID) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string) => void
  readonly onPresetChange: (emulatorId: EmulatorID, preset: string | null) => void
  readonly onResumeChange: (emulatorId: EmulatorID, resume: string | null) => void
}

export const SystemCard = forwardRef<HTMLElement, SystemCardProps>(function SystemCard(
  {
    system,
    systemEnabledEmulators,
    globalEnabledEmulators,
    defaultEmulatorId,
    emulatorVersions,
    emulatorPresets,
    emulatorResume,
    graphics,
    savestate,
    installedVersions,
    installedExecLines,
    managedConfigs,
    installedPaths,
    provisions,
    sharedPackages,
    onEmulatorToggle,
    onSetDefaultEmulator,
    onVersionChange,
    onPresetChange,
    onResumeChange,
  },
  ref,
) {
  const logo = ESDE_LOGOS[system.id]
  const year = SYSTEM_YEARS[system.id]
  const logoOpacity = LOGO_OPACITIES[system.id] ?? 0.1
  const hasInstalledEmulator = system.emulators.some((emu) => installedVersions.has(emu.id))
  const hasAlternatives = systemEnabledEmulators.size > 1

  return (
    <article
      ref={ref}
      className={`border border-outline rounded-card overflow-hidden bg-surface ${hasInstalledEmulator ? 'border-t-2 border-t-accent' : ''}`}
    >
      <div className="flex items-center justify-between h-14 bg-surface-alt px-4">
        <div>
          <h3 className="font-heading text-base font-semibold text-on-surface whitespace-nowrap">
            {system.name}
          </h3>
          <p className="text-xs text-on-surface-muted whitespace-nowrap italic">
            {system.manufacturer}
            {year && ` · ${year}`}
          </p>
        </div>
        <img
          src={logo}
          alt=""
          className="h-6 w-auto saturate-0 will-change-transform"
          style={{ opacity: logoOpacity }}
        />
      </div>
      <div className="p-2 space-y-2 bg-surface">
        {system.emulators.map((emulator) => {
          const execLine = installedExecLines.get(emulator.id)
          const emuManagedConfigs = managedConfigs.get(emulator.id)
          const emuPaths = installedPaths.get(emulator.id)?.[system.id]
          const packageName = emulator.packageName ?? emulator.id
          const isSharedPackage = sharedPackages.has(packageName)
          const isEnabled = systemEnabledEmulators.has(emulator.id)
          const isEnabledElsewhere = !isEnabled && globalEnabledEmulators.has(emulator.id)
          return (
            <EmulatorSubcard
              key={emulator.id}
              emulator={emulator}
              systemId={system.id}
              enabled={isEnabled}
              enabledElsewhere={isEnabledElsewhere}
              isDefault={emulator.id === defaultEmulatorId}
              hasAlternatives={hasAlternatives}
              selectedVersion={emulatorVersions.get(emulator.id) ?? VERSION_DEFAULT}
              installedVersion={installedVersions.get(emulator.id) ?? null}
              provisions={provisions[`${system.id}:${emulator.id}`] ?? []}
              sharedPackage={isSharedPackage}
              preset={emulatorPresets.get(emulator.id) ?? null}
              graphics={graphics}
              resume={emulatorResume.get(emulator.id) ?? null}
              savestate={savestate}
              onToggle={(enabled) => onEmulatorToggle(system.id, emulator.id, enabled)}
              onSetDefault={() => onSetDefaultEmulator(system.id, emulator.id)}
              onVersionChange={(version) => onVersionChange(emulator.id, version)}
              onPresetChange={(preset) => onPresetChange(emulator.id, preset)}
              onResumeChange={(resume) => onResumeChange(emulator.id, resume)}
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
})

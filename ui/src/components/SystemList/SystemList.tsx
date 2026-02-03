import { EmulatorRow, SystemRow } from '@/components/SystemRow/SystemRow'
import type { DoctorResponse, EmulatorID, System, SystemID } from '@/types/daemon'
import type { Manufacturer } from '@/types/model.gen'
import { MANUFACTURER_ORDER } from '@/types/ui'

const SYSTEM_RELEASE_YEAR: Record<SystemID, number> = {
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

export interface SystemListProps {
  readonly systems: readonly System[]
  readonly systemEmulators: ReadonlyMap<SystemID, EmulatorID[]>
  readonly emulatorVersions: ReadonlyMap<EmulatorID, string | null>
  readonly installedVersions: ReadonlyMap<EmulatorID, string>
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onEnableDefault: (systemId: SystemID) => void
  readonly onEmulatorToggle: (systemId: SystemID, emulatorId: EmulatorID, enabled: boolean) => void
  readonly onVersionChange: (emulatorId: EmulatorID, version: string | null) => void
}

function groupByManufacturer(systems: readonly System[]): Map<Manufacturer, System[]> {
  const groups = new Map<Manufacturer, System[]>()

  for (const manufacturer of MANUFACTURER_ORDER) {
    groups.set(manufacturer, [])
  }

  for (const system of systems) {
    const group = groups.get(system.manufacturer)
    if (group) {
      group.push(system)
    }
  }

  for (const group of groups.values()) {
    group.sort((a, b) => (SYSTEM_RELEASE_YEAR[a.id] ?? 9999) - (SYSTEM_RELEASE_YEAR[b.id] ?? 9999))
  }

  return groups
}

function getPackageName(emulatorId: EmulatorID): string {
  if (emulatorId.includes(':')) {
    return emulatorId.split(':')[0] ?? emulatorId
  }
  return emulatorId
}

function getEmulatorSharingInfo(
  systems: readonly System[],
  systemEmulators: ReadonlyMap<SystemID, EmulatorID[]>,
  installedVersions: ReadonlyMap<EmulatorID, string>,
  currentSystemId: SystemID,
  currentEmulatorId: EmulatorID,
): { sharedWith: string[]; installedFor: string[] } {
  const currentPackage = getPackageName(currentEmulatorId)
  const sharedWith: string[] = []
  const installedFor: string[] = []

  for (const system of systems) {
    if (system.id === currentSystemId) continue

    const emulators = systemEmulators.get(system.id)
    if (!emulators) continue

    for (const selectedEmulator of emulators) {
      const selectedPackage = getPackageName(selectedEmulator)
      if (selectedPackage === currentPackage) {
        sharedWith.push(system.label)
        for (const [installedId] of installedVersions) {
          if (getPackageName(installedId) === currentPackage) {
            installedFor.push(system.label)
            break
          }
        }
        break
      }
    }
  }

  return { sharedWith, installedFor }
}

export function SystemList({
  systems,
  systemEmulators,
  emulatorVersions,
  installedVersions,
  provisions,
  userStore,
  onEnableDefault,
  onEmulatorToggle,
  onVersionChange,
}: SystemListProps) {
  const grouped = groupByManufacturer(systems)

  return (
    <div className="space-y-6">
      {MANUFACTURER_ORDER.map((manufacturer) => {
        const manufacturerSystems = grouped.get(manufacturer)
        if (!manufacturerSystems || manufacturerSystems.length === 0) {
          return null
        }

        return (
          <section key={manufacturer}>
            <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-2 px-3">
              {manufacturer}
            </h2>
            <div className="space-y-2">
              {manufacturerSystems.map((system) => {
                const enabledEmulators = systemEmulators.get(system.id) ?? []

                return (
                  <div
                    key={system.id}
                    className="bg-white rounded-lg border border-gray-200 overflow-hidden"
                  >
                    <SystemRow system={system} onEnableDefault={onEnableDefault} />
                    {system.emulators.map((emulator) => {
                      const isEmulatorEnabled = enabledEmulators.includes(emulator.id)
                      const pinnedVersion = emulatorVersions.get(emulator.id) ?? null
                      const installedVersion = installedVersions.get(emulator.id) ?? null

                      const { sharedWith, installedFor } = getEmulatorSharingInfo(
                        systems,
                        systemEmulators,
                        installedVersions,
                        system.id,
                        emulator.id,
                      )

                      return (
                        <EmulatorRow
                          key={emulator.id}
                          systemId={system.id}
                          systemName={system.name}
                          emulator={emulator}
                          pinnedVersion={pinnedVersion}
                          installedVersion={installedVersion}
                          enabled={isEmulatorEnabled}
                          emulatorSharedWith={sharedWith}
                          emulatorInstalledFor={installedFor}
                          provisions={provisions[emulator.id] ?? []}
                          userStore={userStore}
                          onToggle={onEmulatorToggle}
                          onVersionChange={onVersionChange}
                        />
                      )
                    })}
                  </div>
                )
              })}
            </div>
          </section>
        )
      })}
    </div>
  )
}

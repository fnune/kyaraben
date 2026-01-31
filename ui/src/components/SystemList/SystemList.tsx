import { SystemRow } from '@/components/SystemRow/SystemRow'
import type { DoctorResponse, EmulatorID, System, SystemID } from '@/types/daemon'
import type { Manufacturer } from '@/types/model.gen'
import { MANUFACTURER_ORDER } from '@/types/ui'

const SYSTEM_RELEASE_YEAR: Record<SystemID, number> = {
  // Nintendo
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
  // Sony
  psx: 1994,
  ps2: 2000,
  ps3: 2006,
  psp: 2004,
  psvita: 2011,
  // Sega
  genesis: 1988,
  saturn: 1994,
  dreamcast: 1998,
}

export interface SystemListProps {
  readonly systems: readonly System[]
  readonly selections: ReadonlyMap<SystemID, EmulatorID>
  readonly versionSelections: ReadonlyMap<SystemID, string | null>
  readonly installedVersions: ReadonlyMap<EmulatorID, string>
  readonly provisions: DoctorResponse
  readonly userStore: string
  readonly onToggle: (systemId: SystemID, enabled: boolean) => void
  readonly onEmulatorChange: (systemId: SystemID, emulatorId: EmulatorID) => void
  readonly onVersionChange: (systemId: SystemID, version: string | null) => void
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
  selections: ReadonlyMap<SystemID, EmulatorID>,
  installedVersions: ReadonlyMap<EmulatorID, string>,
  currentSystemId: SystemID,
  currentEmulatorId: EmulatorID | null,
): { sharedWith: string[]; installedFor: string[] } {
  if (!currentEmulatorId) {
    return { sharedWith: [], installedFor: [] }
  }

  const currentPackage = getPackageName(currentEmulatorId)
  const sharedWith: string[] = []
  const installedFor: string[] = []

  for (const system of systems) {
    if (system.id === currentSystemId) continue

    const selectedEmulator = selections.get(system.id)
    if (!selectedEmulator) continue

    const selectedPackage = getPackageName(selectedEmulator)
    if (selectedPackage === currentPackage) {
      sharedWith.push(system.label)
      for (const [installedId] of installedVersions) {
        if (getPackageName(installedId) === currentPackage) {
          installedFor.push(system.label)
          break
        }
      }
    }
  }

  return { sharedWith, installedFor }
}

export function SystemList({
  systems,
  selections,
  versionSelections,
  installedVersions,
  provisions,
  userStore,
  onToggle,
  onEmulatorChange,
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
            <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
              {manufacturerSystems.map((system) => {
                const selectedEmulator = selections.get(system.id) ?? null
                const effectiveEmulator = selectedEmulator ?? system.emulators[0]?.id ?? null
                const installedEmulator = system.emulators.find((e) => installedVersions.has(e.id))
                const installedVersion = installedEmulator
                  ? (installedVersions.get(installedEmulator.id) ?? null)
                  : null

                const { sharedWith, installedFor } = getEmulatorSharingInfo(
                  systems,
                  selections,
                  installedVersions,
                  system.id,
                  effectiveEmulator,
                )

                return (
                  <SystemRow
                    key={system.id}
                    system={system}
                    selectedEmulator={selectedEmulator}
                    pinnedVersion={versionSelections.get(system.id) ?? null}
                    installedVersion={installedVersion}
                    provisions={provisions[system.id] ?? []}
                    enabled={selections.has(system.id)}
                    userStore={userStore}
                    emulatorSharedWith={sharedWith}
                    emulatorInstalledFor={installedFor}
                    onToggle={onToggle}
                    onEmulatorChange={onEmulatorChange}
                    onVersionChange={onVersionChange}
                  />
                )
              })}
            </div>
          </section>
        )
      })}
    </div>
  )
}

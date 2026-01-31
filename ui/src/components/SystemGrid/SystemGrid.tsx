import { SystemCard } from '@/components/SystemCard/SystemCard'
import type { DoctorResponse, EmulatorID, System, SystemID } from '@/types/daemon'
import type { Manufacturer } from '@/types/model.gen'
import { MANUFACTURER_ORDER } from '@/types/ui'

export interface SystemGridProps {
  readonly systems: readonly System[]
  readonly selections: ReadonlyMap<SystemID, EmulatorID>
  readonly versionSelections: ReadonlyMap<SystemID, string | null>
  readonly installedVersions: ReadonlyMap<EmulatorID, string>
  readonly provisions: DoctorResponse
  readonly onToggle: (systemId: SystemID, enabled: boolean) => void
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

  return groups
}

export function SystemGrid({
  systems,
  selections,
  versionSelections,
  installedVersions,
  provisions,
  onToggle,
  onVersionChange,
}: SystemGridProps) {
  const grouped = groupByManufacturer(systems)

  return (
    <div className="space-y-8">
      {MANUFACTURER_ORDER.map((manufacturer) => {
        const manufacturerSystems = grouped.get(manufacturer)
        if (!manufacturerSystems || manufacturerSystems.length === 0) {
          return null
        }

        return (
          <section key={manufacturer}>
            <h2 className="text-lg font-semibold text-gray-700 mb-4">{manufacturer}</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {manufacturerSystems.map((system) => {
                const selectedEmulator = selections.get(system.id) ?? null
                const installedVersion = selectedEmulator
                  ? (installedVersions.get(selectedEmulator) ?? null)
                  : null

                return (
                  <SystemCard
                    key={system.id}
                    system={system}
                    selectedEmulator={selectedEmulator}
                    pinnedVersion={versionSelections.get(system.id) ?? null}
                    installedVersion={installedVersion}
                    provisions={provisions[system.id] ?? []}
                    enabled={selections.has(system.id)}
                    onToggle={onToggle}
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

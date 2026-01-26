import type { DoctorResponse, EmulatorID, System, SystemID } from "../../types";
import { MANUFACTURER_ORDER, SYSTEM_MANUFACTURERS, type Manufacturer } from "../../types/ui";
import { SystemCard } from "../SystemCard";
import styles from "./SystemGrid.module.css";

export interface SystemGridProps {
  readonly systems: readonly System[];
  readonly selections: ReadonlyMap<SystemID, EmulatorID>;
  readonly provisions: DoctorResponse;
  readonly onToggle: (systemId: SystemID, enabled: boolean) => void;
}

function groupByManufacturer(systems: readonly System[]): Map<Manufacturer, System[]> {
  const groups = new Map<Manufacturer, System[]>();

  for (const manufacturer of MANUFACTURER_ORDER) {
    groups.set(manufacturer, []);
  }

  for (const system of systems) {
    const manufacturer = SYSTEM_MANUFACTURERS[system.id];
    const group = groups.get(manufacturer);
    if (group) {
      group.push(system);
    }
  }

  return groups;
}

export function SystemGrid({
  systems,
  selections,
  provisions,
  onToggle,
}: SystemGridProps) {
  const grouped = groupByManufacturer(systems);

  return (
    <div className={styles.grid}>
      {MANUFACTURER_ORDER.map((manufacturer) => {
        const manufacturerSystems = grouped.get(manufacturer);
        if (!manufacturerSystems || manufacturerSystems.length === 0) {
          return null;
        }

        return (
          <section key={manufacturer} className={styles.group}>
            <h2 className={styles.groupTitle}>{manufacturer}</h2>
            <div className={styles.cards}>
              {manufacturerSystems.map((system) => (
                <SystemCard
                  key={system.id}
                  system={system}
                  selectedEmulator={selections.get(system.id) ?? null}
                  provisions={provisions[system.id] ?? []}
                  enabled={selections.has(system.id)}
                  onToggle={onToggle}
                />
              ))}
            </div>
          </section>
        );
      })}
    </div>
  );
}

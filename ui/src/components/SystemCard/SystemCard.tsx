import type { EmulatorID, ProvisionResult, System, SystemID } from "../../types";
import { SystemIcon } from "../SystemIcon";
import styles from "./SystemCard.module.css";

export interface SystemCardProps {
  readonly system: System;
  readonly selectedEmulator: EmulatorID | null;
  readonly provisions: readonly ProvisionResult[];
  readonly enabled: boolean;
  readonly onToggle: (systemId: SystemID, enabled: boolean) => void;
}

function ProvisionBadge({ provision }: { readonly provision: ProvisionResult }) {
  const isOk = provision.status === "found";
  const isOptional = !provision.required;
  const statusClass = isOk ? styles.badgeOk : isOptional ? styles.badgeOptional : styles.badgeMissing;
  const statusText = isOk ? "OK" : isOptional ? "optional" : "missing";

  return (
    <span className={`${styles.badge} ${statusClass}`} title={provision.description}>
      {provision.filename} ({statusText})
    </span>
  );
}

export function SystemCard({
  system,
  selectedEmulator,
  provisions,
  enabled,
  onToggle,
}: SystemCardProps) {
  const emulator = system.emulators.find((e) => e.id === selectedEmulator) ?? system.emulators[0];
  const hasRequiredMissing = provisions.some((p) => p.required && p.status !== "found");
  const hasProvisions = provisions.length > 0;

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onToggle(system.id, e.target.checked);
  };

  return (
    <article className={styles.card}>
      <label className={styles.header}>
        <input
          type="checkbox"
          checked={enabled}
          onChange={handleChange}
          className={styles.checkbox}
        />
        <SystemIcon systemId={system.id} size="medium" />
        <div className={styles.info}>
          <h3 className={styles.name}>{system.name}</h3>
          {emulator && <span className={styles.emulator}>{emulator.name}</span>}
        </div>
      </label>

      {hasProvisions && (
        <div className={styles.provisions}>
          {hasRequiredMissing && (
            <span className={styles.warning}>Requires files</span>
          )}
          <div className={styles.badgeList}>
            {provisions.map((p) => (
              <ProvisionBadge key={p.filename} provision={p} />
            ))}
          </div>
        </div>
      )}
    </article>
  );
}

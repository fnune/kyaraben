import type { SystemID } from "../../types";
import { SYSTEM_MANUFACTURERS, type Manufacturer } from "../../types/ui";
import styles from "./SystemIcon.module.css";

export interface SystemIconProps {
  readonly systemId: SystemID;
  readonly size?: "small" | "medium" | "large";
  readonly className?: string;
}

const SYSTEM_LABELS: Readonly<Record<SystemID, string>> = {
  snes: "SNES",
  gba: "GBA",
  nds: "NDS",
  switch: "NSW",
  psx: "PSX",
  psp: "PSP",
  tic80: "TIC",
  "e2e-test": "TST",
};

const MANUFACTURER_COLORS: Readonly<Record<Manufacturer, string>> = {
  Nintendo: "#e60012",
  Sony: "#003087",
  Other: "#6b7280",
};

export function SystemIcon({
  systemId,
  size = "medium",
  className = "",
}: SystemIconProps) {
  const manufacturer = SYSTEM_MANUFACTURERS[systemId];
  const label = SYSTEM_LABELS[systemId];
  const color = MANUFACTURER_COLORS[manufacturer];

  return (
    <div
      className={`${styles.icon} ${styles[size]} ${className}`}
      style={{ backgroundColor: color }}
      role="img"
      aria-label={`${systemId} icon`}
    >
      <span className={styles.label}>{label}</span>
    </div>
  );
}

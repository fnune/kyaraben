import styles from "./Settings.module.css";

export interface SettingsProps {
  readonly userStore: string;
  readonly onUserStoreChange: (value: string) => void;
}

export function Settings({ userStore, onUserStoreChange }: SettingsProps) {
  return (
    <div className={styles.settings}>
      <label className={styles.field}>
        <span className={styles.label}>Emulation folder</span>
        <input
          type="text"
          value={userStore}
          onChange={(e) => onUserStoreChange(e.target.value)}
          className={styles.input}
          placeholder="~/Emulation"
        />
      </label>
    </div>
  );
}

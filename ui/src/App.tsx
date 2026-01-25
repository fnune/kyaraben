import { useCallback, useEffect, useState } from "react";
import { ProgressDisplay, Settings, SystemGrid } from "./components";
import { useDaemon } from "./hooks";
import type {
  DoctorResponse,
  EmulatorID,
  System,
  SystemID,
} from "./types";
import type { ApplyStatus, ProgressStep } from "./types/ui";
import styles from "./App.module.css";

const PROGRESS_STEP_LABELS: Readonly<Record<string, string>> = {
  start: "Starting",
  build_flake: "Building Nix flake",
  build_nix: "Building with Nix",
  write_configs: "Writing emulator configs",
  save_manifest: "Saving manifest",
  done: "Complete",
};

function parseProgressStep(step: string, message: string): ProgressStep {
  return {
    id: step,
    label: PROGRESS_STEP_LABELS[step] ?? step,
    status: "completed",
    message: message,
  };
}

export function App() {
  const daemon = useDaemon();

  const [systems, setSystems] = useState<readonly System[]>([]);
  const [selections, setSelections] = useState<Map<SystemID, EmulatorID>>(new Map());
  const [provisions, setProvisions] = useState<DoctorResponse>({});
  const [userStore, setUserStore] = useState("~/Emulation");
  const [applyStatus, setApplyStatus] = useState<ApplyStatus>("idle");
  const [progressSteps, setProgressSteps] = useState<readonly ProgressStep[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function init() {
      const [systemsResult, configResult] = await Promise.all([
        daemon.getSystems(),
        daemon.getConfig(),
      ]);

      if (systemsResult.ok) {
        setSystems(systemsResult.data);
      }

      if (configResult.ok) {
        setUserStore(configResult.data.userStore);
        const newSelections = new Map<SystemID, EmulatorID>();
        for (const [sysId, emuId] of Object.entries(configResult.data.systems)) {
          newSelections.set(sysId as SystemID, emuId as EmulatorID);
        }
        setSelections(newSelections);
      }

      const doctorResult = await daemon.runDoctor();
      if (doctorResult.ok) {
        setProvisions(doctorResult.data);
      }
    }

    init();
  }, [daemon]);

  const handleToggle = useCallback((systemId: SystemID, enabled: boolean) => {
    setSelections((prev) => {
      const next = new Map(prev);
      if (enabled) {
        const system = systems.find((s) => s.id === systemId);
        const defaultEmulator = system?.emulators[0];
        if (defaultEmulator) {
          next.set(systemId, defaultEmulator.id);
        }
      } else {
        next.delete(systemId);
      }
      return next;
    });
  }, [systems]);

  const handleApply = useCallback(async () => {
    setApplyStatus("applying");
    setProgressSteps([]);
    setError(null);

    const systemsConfig: Partial<Record<SystemID, EmulatorID>> = {};
    for (const [sysId, emuId] of selections) {
      systemsConfig[sysId] = emuId;
    }

    const configResult = await daemon.setConfig({
      userStore,
      systems: systemsConfig,
    });

    if (!configResult.ok) {
      setError(configResult.error.message);
      setApplyStatus("error");
      return;
    }

    const applyResult = await daemon.apply();

    if (!applyResult.ok) {
      setError(applyResult.error.message);
      setApplyStatus("error");
      return;
    }

    const steps = applyResult.data.map((msg, i) => {
      const parts = msg.split(": ");
      const step = parts[0] ?? `step_${i}`;
      const message = parts.slice(1).join(": ") || msg;
      return parseProgressStep(step, message);
    });

    setProgressSteps(steps);
    setApplyStatus("success");

    const doctorResult = await daemon.runDoctor();
    if (doctorResult.ok) {
      setProvisions(doctorResult.data);
    }
  }, [daemon, selections, userStore]);

  const handleCheckProvisions = useCallback(async () => {
    const result = await daemon.runDoctor();
    if (result.ok) {
      setProvisions(result.data);
    }
  }, [daemon]);

  const isApplying = applyStatus === "applying";

  return (
    <div className={styles.app}>
      <header className={styles.header}>
        <h1 className={styles.title}>Kyaraben</h1>
        <p className={styles.subtitle}>Declarative emulation manager</p>
      </header>

      <main className={styles.main}>
        <Settings userStore={userStore} onUserStoreChange={setUserStore} />

        <SystemGrid
          systems={systems}
          selections={selections}
          provisions={provisions}
          onToggle={handleToggle}
        />

        <div className={styles.actions}>
          <button
            onClick={handleApply}
            disabled={isApplying || selections.size === 0}
            className={styles.primaryButton}
          >
            {isApplying ? "Applying..." : "Apply"}
          </button>
          <button
            onClick={handleCheckProvisions}
            disabled={isApplying}
            className={styles.secondaryButton}
          >
            Check provisions
          </button>
        </div>

        <ProgressDisplay steps={progressSteps} error={error ?? ""} />
      </main>
    </div>
  );
}

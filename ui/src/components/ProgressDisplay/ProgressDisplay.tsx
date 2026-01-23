import type { ProgressStep, ProgressStepStatus } from "../../types/ui";
import styles from "./ProgressDisplay.module.css";

export interface ProgressDisplayProps {
  readonly steps: readonly ProgressStep[];
  readonly error?: string;
}

function StepIcon({ status }: { readonly status: ProgressStepStatus }) {
  switch (status) {
    case "completed":
      return <span className={styles.iconCompleted}>✓</span>;
    case "in_progress":
      return <span className={styles.iconProgress}>●</span>;
    case "error":
      return <span className={styles.iconError}>✗</span>;
    default:
      return <span className={styles.iconPending}>○</span>;
  }
}

export function ProgressDisplay({ steps, error }: ProgressDisplayProps) {
  if (steps.length === 0 && !error) {
    return <></>;
  }

  return (
    <div className={styles.container}>
      {steps.length > 0 && (
        <ol className={styles.steps}>
          {steps.map((step) => (
            <li key={step.id} className={`${styles.step} ${styles[step.status]}`}>
              <StepIcon status={step.status} />
              <span className={styles.label}>{step.label}</span>
              {step.message && <span className={styles.message}>{step.message}</span>}
            </li>
          ))}
        </ol>
      )}

      {error && (
        <div className={styles.error}>
          <strong>Error:</strong> {error}
        </div>
      )}
    </div>
  );
}

import type { ProgressStep } from '@/types/ui'

export function getBuildPhaseSubtitle(step: ProgressStep): string | null {
  switch (step.buildPhase) {
    case 'evaluating':
      return 'Preparing...'
    case 'installing':
      return step.packageName ? `Installing ${step.packageName}...` : null
    case 'downloading':
      return step.packageName ? `Downloading ${step.packageName}` : 'Downloading'
    case 'extracting':
      return step.packageName ? `Extracting ${step.packageName}...` : 'Extracting...'
    case 'installed':
      return step.packageName ? `Installed ${step.packageName}` : null
    case 'skipped':
      return step.packageName ? `${step.packageName} already installed` : null
    default:
      return null
  }
}

export function getStepSubtitle(step: ProgressStep): string | null {
  if (step.id === 'build') {
    if (step.status === 'completed') {
      return 'Done'
    }
    if (step.status === 'in_progress') {
      const buildProgress = getBuildPhaseSubtitle(step)
      if (buildProgress) return buildProgress
      return step.message ?? null
    }
    return null
  }
  return step.message ?? null
}

export function getDownloadSpeedBytes(step: ProgressStep): number {
  if (step.id !== 'build' || step.status !== 'in_progress') return 0
  if (step.buildPhase !== 'downloading') return 0
  return step.bytesPerSecond ?? 0
}

import * as semver from 'semver'

export type ChangeType = 'install' | 'remove' | 'upgrade' | 'downgrade' | null

function compareVersions(installed: string, declared: string): 'upgrade' | 'downgrade' {
  // Try semver comparison (handles "1.2.3", "1.2.3-beta.1", etc.)
  const semverInstalled = semver.valid(semver.coerce(installed))
  const semverDeclared = semver.valid(semver.coerce(declared))
  if (semverInstalled && semverDeclared) {
    return semver.gt(semverDeclared, semverInstalled) ? 'upgrade' : 'downgrade'
  }

  // Try integer comparison (e.g., "24" vs "25")
  const numInstalled = parseInt(installed, 10)
  const numDeclared = parseInt(declared, 10)
  if (!Number.isNaN(numInstalled) && !Number.isNaN(numDeclared)) {
    return numDeclared > numInstalled ? 'upgrade' : 'downgrade'
  }

  // Fall back to string comparison
  return declared > installed ? 'upgrade' : 'downgrade'
}

export interface ChangeTypeConfig {
  readonly label: string
  readonly icon: string
  readonly bgColor: string
  readonly textColor: string
  readonly ringColor: string
}

export const CHANGE_CONFIG: Record<NonNullable<ChangeType>, ChangeTypeConfig> = {
  install: {
    label: 'Install',
    icon: '+',
    bgColor: 'bg-status-ok',
    textColor: 'text-status-ok',
    ringColor: 'ring-status-ok/30',
  },
  remove: {
    label: 'Remove',
    icon: '−',
    bgColor: 'bg-status-error',
    textColor: 'text-status-error',
    ringColor: 'ring-status-error/30',
  },
  upgrade: {
    label: 'Upgrade',
    icon: '↑',
    bgColor: 'bg-accent',
    textColor: 'text-accent',
    ringColor: 'ring-accent/30',
  },
  downgrade: {
    label: 'Downgrade',
    icon: '↓',
    bgColor: 'bg-status-warning',
    textColor: 'text-status-warning',
    ringColor: 'ring-status-warning/30',
  },
}

export function getChangeType(
  enabled: boolean,
  installedVersion: string | null,
  declaredVersion: string | null,
  availableVersions?: readonly string[],
): ChangeType {
  // Not enabled: will remove if currently installed
  if (!enabled) {
    return installedVersion ? 'remove' : null
  }

  // Enabled but not installed: will install
  if (!installedVersion) {
    return 'install'
  }

  // Enabled and installed: check for version change
  if (declaredVersion && installedVersion !== declaredVersion) {
    // Determine if upgrade or downgrade based on version order
    // availableVersions is sorted ascending (oldest first via sort.Strings in Go)
    if (availableVersions && availableVersions.length > 0) {
      const installedIdx = availableVersions.indexOf(installedVersion)
      const declaredIdx = availableVersions.indexOf(declaredVersion)

      // If both versions are in the list, compare indices
      // Higher index = newer version (ascending sort)
      if (installedIdx >= 0 && declaredIdx >= 0) {
        return declaredIdx > installedIdx ? 'upgrade' : 'downgrade'
      }
    }

    // Fallback: compare versions directly
    return compareVersions(installedVersion, declaredVersion)
  }

  // No change
  return null
}

export interface ChangeSummary {
  readonly installs: number
  readonly removes: number
  readonly upgrades: number
  readonly downgrades: number
  readonly total: number
  readonly downloadBytes: number
  readonly freeBytes: number
  readonly hasConfigChanges: boolean
}

export function emptyChangeSummary(): ChangeSummary {
  return {
    installs: 0,
    removes: 0,
    upgrades: 0,
    downgrades: 0,
    total: 0,
    downloadBytes: 0,
    freeBytes: 0,
    hasConfigChanges: false,
  }
}

export function formatBytes(bytes: number): string {
  const absBytes = Math.abs(bytes)
  if (absBytes >= 1024 * 1024 * 1024) {
    return `${(absBytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
  }
  if (absBytes >= 1024 * 1024) {
    return `${Math.round(absBytes / (1024 * 1024))} MB`
  }
  return `${Math.round(absBytes / 1024)} KB`
}

export function addChange(
  summary: ChangeSummary,
  changeType: ChangeType,
  sizeBytes?: number,
): ChangeSummary {
  if (!changeType) return summary

  const updates = {
    install: { installs: summary.installs + 1 },
    remove: { removes: summary.removes + 1 },
    upgrade: { upgrades: summary.upgrades + 1 },
    downgrade: { downgrades: summary.downgrades + 1 },
  }

  const download = changeType === 'install' || changeType === 'upgrade' ? (sizeBytes ?? 0) : 0
  const free = changeType === 'remove' ? (sizeBytes ?? 0) : 0

  return {
    ...summary,
    ...updates[changeType],
    total: summary.total + 1,
    downloadBytes: summary.downloadBytes + download,
    freeBytes: summary.freeBytes + free,
  }
}

export function withConfigChanges(
  summary: ChangeSummary,
  hasConfigChanges: boolean,
): ChangeSummary {
  return { ...summary, hasConfigChanges }
}

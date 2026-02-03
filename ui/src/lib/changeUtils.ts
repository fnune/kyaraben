export type ChangeType = 'install' | 'remove' | 'upgrade' | 'downgrade' | null

function compareSemver(a: string, b: string): number {
  const partsA = a.split('.').map((p) => parseInt(p, 10))
  const partsB = b.split('.').map((p) => parseInt(p, 10))

  if (partsA.some(Number.isNaN) || partsB.some(Number.isNaN)) {
    return NaN
  }

  const maxLen = Math.max(partsA.length, partsB.length)
  for (let i = 0; i < maxLen; i++) {
    const numA = partsA[i] ?? 0
    const numB = partsB[i] ?? 0
    if (numA !== numB) {
      return numA - numB
    }
  }
  return 0
}

function compareInt(a: string, b: string): number {
  const numA = parseInt(a, 10)
  const numB = parseInt(b, 10)
  if (Number.isNaN(numA) || Number.isNaN(numB)) {
    return NaN
  }
  return numA - numB
}

function compareVersions(installed: string, declared: string): 'upgrade' | 'downgrade' {
  // Try semver comparison (e.g., "1.2.3" vs "1.3.0")
  const semverResult = compareSemver(declared, installed)
  if (!Number.isNaN(semverResult)) {
    return semverResult > 0 ? 'upgrade' : 'downgrade'
  }

  // Try integer comparison (e.g., "24" vs "25")
  const intResult = compareInt(declared, installed)
  if (!Number.isNaN(intResult)) {
    return intResult > 0 ? 'upgrade' : 'downgrade'
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
    bgColor: 'bg-emerald-500',
    textColor: 'text-emerald-500',
    ringColor: 'ring-emerald-500/30',
  },
  remove: {
    label: 'Remove',
    icon: '−',
    bgColor: 'bg-red-500',
    textColor: 'text-red-500',
    ringColor: 'ring-red-500/30',
  },
  upgrade: {
    label: 'Upgrade',
    icon: '↑',
    bgColor: 'bg-sky-500',
    textColor: 'text-sky-500',
    ringColor: 'ring-sky-500/30',
  },
  downgrade: {
    label: 'Downgrade',
    icon: '↓',
    bgColor: 'bg-amber-500',
    textColor: 'text-amber-500',
    ringColor: 'ring-amber-500/30',
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
    // availableVersions is ordered newest first, so lower index = newer
    if (availableVersions && availableVersions.length > 0) {
      const installedIdx = availableVersions.indexOf(installedVersion)
      const declaredIdx = availableVersions.indexOf(declaredVersion)

      // If both versions are in the list, compare indices
      if (installedIdx >= 0 && declaredIdx >= 0) {
        return declaredIdx < installedIdx ? 'upgrade' : 'downgrade'
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

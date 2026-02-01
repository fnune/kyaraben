export type ChangeType = 'install' | 'remove' | 'upgrade' | 'downgrade' | null

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

    // Fallback: assume upgrade if we can't determine
    return 'upgrade'
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
}

export function emptyChangeSummary(): ChangeSummary {
  return { installs: 0, removes: 0, upgrades: 0, downgrades: 0, total: 0 }
}

export function addChange(summary: ChangeSummary, changeType: ChangeType): ChangeSummary {
  if (!changeType) return summary

  const updates = {
    install: { installs: summary.installs + 1 },
    remove: { removes: summary.removes + 1 },
    upgrade: { upgrades: summary.upgrades + 1 },
    downgrade: { downgrades: summary.downgrades + 1 },
  }

  return {
    ...summary,
    ...updates[changeType],
    total: summary.total + 1,
  }
}

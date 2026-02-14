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

export interface ChangeItem {
  readonly id: string
  readonly name: string
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
  readonly installItems: readonly ChangeItem[]
  readonly removeItems: readonly ChangeItem[]
  readonly upgradeItems: readonly ChangeItem[]
  readonly downgradeItems: readonly ChangeItem[]
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
    installItems: [],
    removeItems: [],
    upgradeItems: [],
    downgradeItems: [],
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

export interface ChangeGroup {
  readonly type: NonNullable<ChangeType>
  readonly items: readonly ChangeItem[]
}

export function getChangeGroups(summary: ChangeSummary): ChangeGroup[] {
  const groups: ChangeGroup[] = []
  if (summary.installItems.length > 0) {
    groups.push({ type: 'install', items: summary.installItems })
  }
  if (summary.removeItems.length > 0) {
    groups.push({ type: 'remove', items: summary.removeItems })
  }
  if (summary.upgradeItems.length > 0) {
    groups.push({ type: 'upgrade', items: summary.upgradeItems })
  }
  if (summary.downgradeItems.length > 0) {
    groups.push({ type: 'downgrade', items: summary.downgradeItems })
  }
  return groups
}

export function formatChangeSummary(summary: ChangeSummary): string | null {
  const parts: string[] = []

  if (summary.installItems.length > 0) {
    const names = summary.installItems.map((i) => i.name).join(', ')
    parts.push(`Installing ${names}`)
  }
  if (summary.removeItems.length > 0) {
    const names = summary.removeItems.map((i) => i.name).join(', ')
    parts.push(`Removing ${names}`)
  }
  if (summary.upgradeItems.length > 0) {
    const names = summary.upgradeItems.map((i) => i.name).join(', ')
    parts.push(`Upgrading ${names}`)
  }
  if (summary.downgradeItems.length > 0) {
    const names = summary.downgradeItems.map((i) => i.name).join(', ')
    parts.push(`Downgrading ${names}`)
  }

  return parts.length > 0 ? parts.join('; ') : null
}

export function addChange(
  summary: ChangeSummary,
  changeType: ChangeType,
  sizeBytes?: number,
  item?: ChangeItem,
): ChangeSummary {
  if (!changeType) return summary

  const updates = {
    install: {
      installs: summary.installs + 1,
      installItems: item ? [...summary.installItems, item] : summary.installItems,
    },
    remove: {
      removes: summary.removes + 1,
      removeItems: item ? [...summary.removeItems, item] : summary.removeItems,
    },
    upgrade: {
      upgrades: summary.upgrades + 1,
      upgradeItems: item ? [...summary.upgradeItems, item] : summary.upgradeItems,
    },
    downgrade: {
      downgrades: summary.downgrades + 1,
      downgradeItems: item ? [...summary.downgradeItems, item] : summary.downgradeItems,
    },
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

export interface EmulatorChangeInput {
  readonly id: string
  readonly packageName: string
  readonly downloadBytes: number
  readonly coreBytes: number
  readonly changeType: ChangeType
  readonly isInstalled: boolean
}

export function calculateEmulatorSizes(
  emulators: readonly EmulatorChangeInput[],
): Map<string, number> {
  const packagesRemainingInstalled = new Set<string>()
  for (const emu of emulators) {
    const willBeInstalled =
      emu.changeType === 'install' ||
      emu.changeType === 'upgrade' ||
      (emu.isInstalled && emu.changeType !== 'remove')
    if (willBeInstalled) {
      packagesRemainingInstalled.add(emu.packageName)
    }
  }

  const packagesCurrentlyInstalled = new Set<string>()
  for (const emu of emulators) {
    if (emu.isInstalled) {
      packagesCurrentlyInstalled.add(emu.packageName)
    }
  }

  const packagesBeingInstalled = new Set<string>()
  const packagesBeingRemoved = new Set<string>()
  const sizes = new Map<string, number>()

  for (const emu of emulators) {
    let size = emu.coreBytes

    if (emu.changeType === 'install' || emu.changeType === 'upgrade') {
      const isFirstInstall = !packagesBeingInstalled.has(emu.packageName)
      const noCurrentInstall = !packagesCurrentlyInstalled.has(emu.packageName)
      if (isFirstInstall && noCurrentInstall) {
        size += emu.downloadBytes
      }
      packagesBeingInstalled.add(emu.packageName)
    } else if (emu.changeType === 'remove') {
      const isFirstRemoval = !packagesBeingRemoved.has(emu.packageName)
      const noneRemaining = !packagesRemainingInstalled.has(emu.packageName)
      if (isFirstRemoval && noneRemaining) {
        size += emu.downloadBytes
      }
      packagesBeingRemoved.add(emu.packageName)
    }

    sizes.set(emu.id, size)
  }

  return sizes
}

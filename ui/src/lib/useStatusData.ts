import { useMemo, useState } from 'react'
import type {
  EmulatorID,
  EmulatorPaths,
  FrontendID,
  ManagedConfigInfo,
  StatusResponse,
} from '@/types/daemon'

export function useStatusData() {
  const [statusResponse, setStatusResponse] = useState<StatusResponse | null>(null)

  const installedVersions = useMemo(() => {
    const map = new Map<EmulatorID, string>()
    if (!statusResponse) return map
    for (const emu of statusResponse.installedEmulators) {
      map.set(emu.id, emu.version)
    }
    return map
  }, [statusResponse])

  const installedExecLines = useMemo(() => {
    const map = new Map<EmulatorID, string>()
    if (!statusResponse) return map
    for (const emu of statusResponse.installedEmulators) {
      map.set(emu.id, emu.execLine)
    }
    return map
  }, [statusResponse])

  const managedConfigs = useMemo(() => {
    const map = new Map<EmulatorID, ManagedConfigInfo[]>()
    if (!statusResponse) return map
    for (const emu of statusResponse.installedEmulators) {
      if (emu.managedConfigs) {
        map.set(emu.id, emu.managedConfigs)
      }
    }
    return map
  }, [statusResponse])

  const installedPaths = useMemo(() => {
    const map = new Map<EmulatorID, Record<string, EmulatorPaths>>()
    if (!statusResponse) return map
    for (const emu of statusResponse.installedEmulators) {
      if (emu.paths) {
        map.set(emu.id, emu.paths)
      }
    }
    return map
  }, [statusResponse])

  const installedFrontendVersions = useMemo(() => {
    const map = new Map<FrontendID, string>()
    if (!statusResponse) return map
    for (const fe of statusResponse.installedFrontends) {
      map.set(fe.id, fe.version)
    }
    return map
  }, [statusResponse])

  const installedFrontendExecLines = useMemo(() => {
    const map = new Map<FrontendID, string>()
    if (!statusResponse) return map
    for (const fe of statusResponse.installedFrontends) {
      map.set(fe.id, fe.execLine)
    }
    return map
  }, [statusResponse])

  const kyarabenVersion = statusResponse?.kyarabenVersion ?? null
  const manifestKyarabenVersion = statusResponse?.manifestKyarabenVersion ?? null
  const healthWarning = statusResponse?.healthWarning ?? null
  const configWarnings = statusResponse?.configWarnings ?? null

  return {
    setStatusResponse,
    installedVersions,
    installedExecLines,
    managedConfigs,
    installedPaths,
    installedFrontendVersions,
    installedFrontendExecLines,
    kyarabenVersion,
    manifestKyarabenVersion,
    healthWarning,
    configWarnings,
  }
}

import { createContext, type ReactNode, useCallback, useContext, useMemo, useState } from 'react'
import * as daemon from '@/lib/daemon'
import type { HotkeyActionKey } from '@/types/controller'
import type {
  ConfigResponse,
  EmulatorID,
  FrontendID,
  FrontendRef,
  System,
  SystemID,
} from '@/types/daemon'
import { VERSION_DEFAULT } from '@/types/ui'
import { useApply } from './ApplyContext'
import {
  addChange,
  type ChangeSummary,
  calculateEmulatorSizes,
  type EmulatorChangeInput,
  emptyChangeSummary,
  formatChangeSummary,
  getChangeType,
  withConfigChanges,
} from './changeUtils'

interface HotkeyConfig {
  modifier: string
  saveState: string
  loadState: string
  nextSlot: string
  prevSlot: string
  fastForward: string
  rewind: string
  pause: string
  screenshot: string
  quit: string
  toggleFullscreen: string
  openMenu: string
}

interface ConfigState {
  collection: string
  graphicsShaders: string
  savestateResume: string
  controllerNintendoConfirm: string
  hotkeys: HotkeyConfig
  systemEmulators: Map<SystemID, EmulatorID[]>
  emulatorVersions: Map<EmulatorID, string>
  emulatorShaders: Map<EmulatorID, string | null>
  emulatorResume: Map<EmulatorID, string | null>
  enabledFrontends: Map<FrontendID, boolean>
  frontendVersions: Map<FrontendID, string>
}

function defaultHotkeys(): HotkeyConfig {
  return {
    modifier: 'Back',
    saveState: 'RightShoulder',
    loadState: 'LeftShoulder',
    nextSlot: 'DPadRight',
    prevSlot: 'DPadLeft',
    fastForward: 'Y',
    rewind: 'X',
    pause: 'A',
    screenshot: 'B',
    quit: 'Start',
    toggleFullscreen: 'LeftStick',
    openMenu: 'RightStick',
  }
}

function emptyConfigState(): ConfigState {
  return {
    collection: '',
    graphicsShaders: '',
    savestateResume: '',
    controllerNintendoConfirm: '',
    hotkeys: defaultHotkeys(),
    systemEmulators: new Map(),
    emulatorVersions: new Map(),
    emulatorShaders: new Map(),
    emulatorResume: new Map(),
    enabledFrontends: new Map(),
    frontendVersions: new Map(),
  }
}

function cloneConfigState(state: ConfigState): ConfigState {
  return {
    collection: state.collection,
    graphicsShaders: state.graphicsShaders,
    savestateResume: state.savestateResume,
    controllerNintendoConfirm: state.controllerNintendoConfirm,
    hotkeys: { ...state.hotkeys },
    systemEmulators: new Map(state.systemEmulators),
    emulatorVersions: new Map(state.emulatorVersions),
    emulatorShaders: new Map(state.emulatorShaders),
    emulatorResume: new Map(state.emulatorResume),
    enabledFrontends: new Map(state.enabledFrontends),
    frontendVersions: new Map(state.frontendVersions),
  }
}

function parseConfigResponse(data: ConfigResponse): ConfigState {
  const systemEmulators = new Map<SystemID, EmulatorID[]>()
  const emulatorVersions = new Map<EmulatorID, string>()
  const emulatorShaders = new Map<EmulatorID, string | null>()
  const emulatorResume = new Map<EmulatorID, string | null>()
  const enabledFrontends = new Map<FrontendID, boolean>()
  const frontendVersions = new Map<FrontendID, string>()

  for (const [sysId, emulatorIds] of Object.entries(data.systems)) {
    if (emulatorIds && emulatorIds.length > 0) {
      systemEmulators.set(sysId as SystemID, emulatorIds as EmulatorID[])
    }
  }

  if (data.emulators) {
    for (const [emuId, conf] of Object.entries(data.emulators)) {
      if (conf.version) {
        emulatorVersions.set(emuId as EmulatorID, conf.version)
      }
      if (conf.shaders !== undefined) {
        emulatorShaders.set(emuId as EmulatorID, conf.shaders ?? null)
      }
      if (conf.resume !== undefined) {
        emulatorResume.set(emuId as EmulatorID, conf.resume ?? null)
      }
    }
  }

  if (data.frontends) {
    for (const [feId, conf] of Object.entries(data.frontends)) {
      enabledFrontends.set(feId as FrontendID, conf.enabled)
      if (conf.version) {
        frontendVersions.set(feId as FrontendID, conf.version)
      }
    }
  }

  return {
    collection: data.collection,
    graphicsShaders: data.graphics?.shaders ?? '',
    savestateResume: data.savestate?.resume ?? '',
    controllerNintendoConfirm: data.controller?.nintendoConfirm ?? 'east',
    hotkeys: data.controller?.hotkeys ?? defaultHotkeys(),
    systemEmulators,
    emulatorVersions,
    emulatorShaders,
    emulatorResume,
    enabledFrontends,
    frontendVersions,
  }
}

interface ConfigContextValue {
  configState: ConfigState
  configReady: boolean
  systems: readonly System[]
  frontends: readonly FrontendRef[]
  installedVersions: Map<EmulatorID, string>
  installedFrontendVersions: Map<FrontendID, string>
  enabledEmulators: ReadonlySet<EmulatorID>
  changes: ChangeSummary
  configChanges: readonly string[]
  upgradeAvailable: boolean

  setCollection: (value: string) => void
  setGraphicsShaders: (value: string) => void
  setSavestateResume: (value: string) => void
  setControllerNintendoConfirm: (value: string) => void
  setHotkeyModifier: (value: string) => void
  setHotkeyAction: (key: HotkeyActionKey, value: string) => void
  resetHotkeys: () => void
  toggleEmulator: (systemId: SystemID, emulatorId: EmulatorID, enabled: boolean) => void
  setEmulatorVersion: (emulatorId: EmulatorID, version: string) => void
  setEmulatorShaders: (emulatorId: EmulatorID, shaders: string | null) => void
  setEmulatorResume: (emulatorId: EmulatorID, resume: string | null) => void
  toggleFrontend: (frontendId: FrontendID, enabled: boolean) => void
  setFrontendVersion: (frontendId: FrontendID, version: string) => void
  enableAllSystems: () => void

  apply: () => Promise<void>
  reapply: () => Promise<void>
  discard: () => Promise<void>

  setSystems: (systems: readonly System[]) => void
  setFrontends: (frontends: readonly FrontendRef[]) => void
  setInstalledVersions: (versions: Map<EmulatorID, string>) => void
  setInstalledFrontendVersions: (versions: Map<FrontendID, string>) => void
  setUpgradeAvailable: (available: boolean) => void
  initFromResponse: (config: ConfigResponse) => void
  refreshAfterApply: () => Promise<void>
}

const ConfigContext = createContext<ConfigContextValue | null>(null)

export function useConfig(): ConfigContextValue {
  const context = useContext(ConfigContext)
  if (!context) {
    throw new Error('useConfig must be used within a ConfigProvider')
  }
  return context
}

interface ConfigProviderProps {
  children: ReactNode
}

export function ConfigProvider({ children }: ConfigProviderProps) {
  const [configState, setConfigState] = useState<ConfigState>(emptyConfigState)
  const [configReady, setConfigReady] = useState(false)
  const [systems, setSystems] = useState<readonly System[]>([])
  const [frontends, setFrontends] = useState<readonly FrontendRef[]>([])
  const [installedVersions, setInstalledVersions] = useState<Map<EmulatorID, string>>(new Map())
  const [installedFrontendVersions, setInstalledFrontendVersions] = useState<
    Map<FrontendID, string>
  >(new Map())
  const [upgradeAvailable, setUpgradeAvailable] = useState(false)

  const [savedConfig, setSavedConfig] = useState<ConfigState>(emptyConfigState)
  const { apply: applyFromContext } = useApply()

  const enabledEmulators = useMemo(
    () => new Set(Array.from(configState.systemEmulators.values()).flat()),
    [configState.systemEmulators],
  )

  const configChanges = useMemo(() => {
    if (!configReady) return []
    const changes: string[] = []

    if (configState.collection !== savedConfig.collection) {
      changes.push('Collection')
    }
    if (configState.graphicsShaders !== savedConfig.graphicsShaders) {
      changes.push('Shader settings')
    }
    if (configState.savestateResume !== savedConfig.savestateResume) {
      changes.push('Savestate settings')
    }
    if (configState.controllerNintendoConfirm !== savedConfig.controllerNintendoConfirm) {
      changes.push('Controller settings')
    }

    const hotkeysChanged = Object.keys(configState.hotkeys).some(
      (key) =>
        configState.hotkeys[key as keyof HotkeyConfig] !==
        savedConfig.hotkeys[key as keyof HotkeyConfig],
    )
    if (hotkeysChanged) {
      changes.push('Hotkey settings')
    }

    const systemEmulatorsChanged = (() => {
      if (configState.systemEmulators.size !== savedConfig.systemEmulators.size) return true
      for (const [sysId, emuIds] of configState.systemEmulators) {
        const savedIds = savedConfig.systemEmulators.get(sysId)
        if (!savedIds || emuIds.length !== savedIds.length) return true
        if (!emuIds.every((id, i) => savedIds[i] === id)) return true
      }
      for (const sysId of savedConfig.systemEmulators.keys()) {
        if (!configState.systemEmulators.has(sysId)) return true
      }
      return false
    })()
    if (systemEmulatorsChanged) {
      changes.push('Enabled emulators')
    }

    const emulatorVersionsChanged = (() => {
      for (const [emuId, version] of configState.emulatorVersions) {
        const saved = savedConfig.emulatorVersions.get(emuId) ?? VERSION_DEFAULT
        if (saved !== version) return true
      }
      for (const [emuId, saved] of savedConfig.emulatorVersions) {
        const current = configState.emulatorVersions.get(emuId) ?? VERSION_DEFAULT
        if (current !== saved) return true
      }
      return false
    })()
    if (emulatorVersionsChanged) {
      changes.push('Emulator versions')
    }

    const emulatorShadersChanged = (() => {
      for (const [emuId, shaders] of configState.emulatorShaders) {
        const saved = savedConfig.emulatorShaders.get(emuId) ?? null
        if (saved !== shaders) return true
      }
      for (const [emuId, saved] of savedConfig.emulatorShaders) {
        const current = configState.emulatorShaders.get(emuId) ?? null
        if (current !== saved) return true
      }
      return false
    })()
    if (emulatorShadersChanged) {
      changes.push('Emulator shaders')
    }

    const emulatorResumeChanged = (() => {
      for (const [emuId, resume] of configState.emulatorResume) {
        const saved = savedConfig.emulatorResume.get(emuId) ?? null
        if (saved !== resume) return true
      }
      for (const [emuId, saved] of savedConfig.emulatorResume) {
        const current = configState.emulatorResume.get(emuId) ?? null
        if (current !== saved) return true
      }
      return false
    })()
    if (emulatorResumeChanged) {
      changes.push('Emulator resume')
    }

    const enabledFrontendsChanged = (() => {
      if (configState.enabledFrontends.size !== savedConfig.enabledFrontends.size) return true
      for (const [feId, enabled] of configState.enabledFrontends) {
        if (savedConfig.enabledFrontends.get(feId) !== enabled) return true
      }
      for (const feId of savedConfig.enabledFrontends.keys()) {
        if (!configState.enabledFrontends.has(feId)) return true
      }
      return false
    })()
    if (enabledFrontendsChanged) {
      changes.push('Enabled frontends')
    }

    const frontendVersionsChanged = (() => {
      for (const [feId, version] of configState.frontendVersions) {
        const saved = savedConfig.frontendVersions.get(feId) ?? VERSION_DEFAULT
        if (saved !== version) return true
      }
      for (const [feId, saved] of savedConfig.frontendVersions) {
        const current = configState.frontendVersions.get(feId) ?? VERSION_DEFAULT
        if (current !== saved) return true
      }
      return false
    })()
    if (frontendVersionsChanged) {
      changes.push('Frontend versions')
    }

    return changes
  }, [configReady, configState, savedConfig])

  const changes = useMemo(() => {
    let summary = emptyChangeSummary()
    const seenEmulators = new Set<EmulatorID>()

    interface EmulatorInputWithName extends EmulatorChangeInput {
      name: string
    }

    const emulatorInputs: EmulatorInputWithName[] = []

    for (const system of systems) {
      for (const emulator of system.emulators) {
        if (seenEmulators.has(emulator.id)) continue
        seenEmulators.add(emulator.id)

        const enabled = enabledEmulators.has(emulator.id)
        const installedVersion = installedVersions.get(emulator.id) ?? null
        const selectedVersion = configState.emulatorVersions.get(emulator.id) ?? VERSION_DEFAULT
        const effectiveVersion =
          selectedVersion === VERSION_DEFAULT ? (emulator.defaultVersion ?? null) : selectedVersion

        const changeType = getChangeType(
          enabled,
          installedVersion,
          effectiveVersion,
          emulator.availableVersions,
        )

        emulatorInputs.push({
          id: emulator.id,
          name: emulator.name,
          packageName: emulator.packageName ?? emulator.id,
          downloadBytes: emulator.downloadBytes ?? 0,
          coreBytes: emulator.coreBytes ?? 0,
          changeType,
          isInstalled: installedVersion !== null,
        })
      }
    }

    const emulatorSizes = calculateEmulatorSizes(emulatorInputs)
    for (const input of emulatorInputs) {
      summary = addChange(summary, input.changeType, emulatorSizes.get(input.id) ?? 0, {
        id: input.id,
        name: input.name,
      })
    }

    for (const frontend of frontends) {
      const enabled = configState.enabledFrontends.get(frontend.id) ?? false
      const installedVersion = installedFrontendVersions.get(frontend.id) ?? null
      const selectedVersion = configState.frontendVersions.get(frontend.id) ?? VERSION_DEFAULT
      const effectiveVersion =
        selectedVersion === VERSION_DEFAULT ? (frontend.defaultVersion ?? null) : selectedVersion

      const changeType = getChangeType(
        enabled,
        installedVersion,
        effectiveVersion,
        frontend.availableVersions,
      )

      const downloadBytes = frontend.downloadBytes ?? 0
      summary = addChange(summary, changeType, downloadBytes, {
        id: frontend.id,
        name: frontend.name,
      })
    }

    return withConfigChanges(summary, configChanges)
  }, [
    systems,
    frontends,
    enabledEmulators,
    configState.emulatorVersions,
    configState.enabledFrontends,
    configState.frontendVersions,
    installedVersions,
    installedFrontendVersions,
    configChanges,
  ])

  const setCollection = useCallback((value: string) => {
    setConfigState((prev) => ({ ...prev, collection: value }))
  }, [])

  const setGraphicsShaders = useCallback((value: string) => {
    setConfigState((prev) => ({ ...prev, graphicsShaders: value }))
  }, [])

  const setSavestateResume = useCallback((value: string) => {
    setConfigState((prev) => ({ ...prev, savestateResume: value }))
  }, [])

  const setControllerNintendoConfirm = useCallback((value: string) => {
    setConfigState((prev) => ({ ...prev, controllerNintendoConfirm: value }))
  }, [])

  const setHotkeyModifier = useCallback((value: string) => {
    setConfigState((prev) => ({
      ...prev,
      hotkeys: { ...prev.hotkeys, modifier: value },
    }))
  }, [])

  const setHotkeyAction = useCallback((key: HotkeyActionKey, value: string) => {
    setConfigState((prev) => ({
      ...prev,
      hotkeys: { ...prev.hotkeys, [key]: value },
    }))
  }, [])

  const resetHotkeys = useCallback(() => {
    setConfigState((prev) => ({
      ...prev,
      hotkeys: defaultHotkeys(),
    }))
  }, [])

  const toggleEmulator = useCallback(
    (systemId: SystemID, emulatorId: EmulatorID, enabled: boolean) => {
      setConfigState((prev) => {
        const next = new Map(prev.systemEmulators)
        const current = next.get(systemId) ?? []

        if (enabled) {
          if (!current.includes(emulatorId)) {
            next.set(systemId, [...current, emulatorId])
          }
        } else {
          const filtered = current.filter((id) => id !== emulatorId)
          if (filtered.length === 0) {
            next.delete(systemId)
          } else {
            next.set(systemId, filtered)
          }
        }
        return { ...prev, systemEmulators: next }
      })
    },
    [],
  )

  const setEmulatorVersion = useCallback((emulatorId: EmulatorID, version: string) => {
    setConfigState((prev) => {
      const next = new Map(prev.emulatorVersions)
      next.set(emulatorId, version)
      return { ...prev, emulatorVersions: next }
    })
  }, [])

  const setEmulatorShaders = useCallback((emulatorId: EmulatorID, shaders: string | null) => {
    setConfigState((prev) => {
      const next = new Map(prev.emulatorShaders)
      if (shaders === null) {
        next.delete(emulatorId)
      } else {
        next.set(emulatorId, shaders)
      }
      return { ...prev, emulatorShaders: next }
    })
  }, [])

  const setEmulatorResume = useCallback((emulatorId: EmulatorID, resume: string | null) => {
    setConfigState((prev) => {
      const next = new Map(prev.emulatorResume)
      if (resume === null) {
        next.delete(emulatorId)
      } else {
        next.set(emulatorId, resume)
      }
      return { ...prev, emulatorResume: next }
    })
  }, [])

  const toggleFrontend = useCallback((frontendId: FrontendID, enabled: boolean) => {
    setConfigState((prev) => {
      const next = new Map(prev.enabledFrontends)
      next.set(frontendId, enabled)
      return { ...prev, enabledFrontends: next }
    })
  }, [])

  const setFrontendVersion = useCallback((frontendId: FrontendID, version: string) => {
    setConfigState((prev) => {
      const next = new Map(prev.frontendVersions)
      next.set(frontendId, version)
      return { ...prev, frontendVersions: next }
    })
  }, [])

  const enableAllSystems = useCallback(() => {
    const newSystemEmulators = new Map<SystemID, EmulatorID[]>()
    for (const sys of systems) {
      newSystemEmulators.set(sys.id, [sys.defaultEmulatorId])
    }
    setConfigState((prev) => ({ ...prev, systemEmulators: newSystemEmulators }))
  }, [systems])

  const buildApplyPayload = useCallback((state: ConfigState, changeSummary: ChangeSummary) => {
    const systemsConfig: Record<string, string[]> = {}
    for (const [sysId, emuIds] of state.systemEmulators) {
      systemsConfig[sysId] = emuIds
    }

    const emulatorsConfig: Record<
      string,
      { version?: string; shaders?: string | null; resume?: string | null }
    > = {}
    for (const [emuId, version] of state.emulatorVersions) {
      if (version === VERSION_DEFAULT) {
        emulatorsConfig[emuId] = { version: '' }
      } else {
        emulatorsConfig[emuId] = { version }
      }
    }
    for (const [emuId, shaders] of state.emulatorShaders) {
      emulatorsConfig[emuId] = { ...emulatorsConfig[emuId], shaders }
    }
    for (const [emuId, resume] of state.emulatorResume) {
      emulatorsConfig[emuId] = { ...emulatorsConfig[emuId], resume }
    }

    const frontendsConfig: Record<string, { enabled: boolean; version?: string }> = {}
    for (const [feId, enabled] of state.enabledFrontends) {
      const version = state.frontendVersions.get(feId)
      if (version === VERSION_DEFAULT) {
        frontendsConfig[feId] = { enabled, version: '' }
      } else if (version) {
        frontendsConfig[feId] = { enabled, version }
      } else {
        frontendsConfig[feId] = { enabled }
      }
    }

    const summaryMessage = formatChangeSummary(changeSummary)
    return {
      collection: state.collection,
      systems: systemsConfig,
      emulators: emulatorsConfig,
      frontends: frontendsConfig,
      ...(state.graphicsShaders && { graphics: { shaders: state.graphicsShaders } }),
      ...(state.savestateResume && { savestate: { resume: state.savestateResume } }),
      controller: {
        nintendoConfirm: state.controllerNintendoConfirm,
        hotkeys: state.hotkeys,
      },
      ...(summaryMessage && { summaryMessage }),
    }
  }, [])

  const apply = useCallback(async () => {
    const payload = buildApplyPayload(configState, changes)
    await applyFromContext(payload)
  }, [applyFromContext, buildApplyPayload, configState, changes])

  const reapply = useCallback(async () => {
    const payload = buildApplyPayload(savedConfig, emptyChangeSummary())
    await applyFromContext(payload)
  }, [applyFromContext, buildApplyPayload, savedConfig])

  const discard = useCallback(async () => {
    const configResult = await daemon.getConfig()
    if (configResult.ok) {
      setConfigState(parseConfigResponse(configResult.data))
    }
  }, [])

  const initFromResponse = useCallback((config: ConfigResponse) => {
    const parsed = parseConfigResponse(config)
    setConfigState(parsed)
    setSavedConfig(cloneConfigState(parsed))
    setConfigReady(true)
  }, [])

  const refreshAfterApply = useCallback(async () => {
    const configResult = await daemon.getConfig()
    if (configResult.ok) {
      const parsed = parseConfigResponse(configResult.data)
      setSavedConfig(cloneConfigState(parsed))
    }
  }, [])

  const value = useMemo<ConfigContextValue>(
    () => ({
      configState,
      configReady,
      systems,
      frontends,
      installedVersions,
      installedFrontendVersions,
      enabledEmulators,
      changes,
      configChanges,
      upgradeAvailable,

      setCollection,
      setGraphicsShaders,
      setSavestateResume,
      setControllerNintendoConfirm,
      setHotkeyModifier,
      setHotkeyAction,
      resetHotkeys,
      toggleEmulator,
      setEmulatorVersion,
      setEmulatorShaders,
      setEmulatorResume,
      toggleFrontend,
      setFrontendVersion,
      enableAllSystems,

      apply,
      reapply,
      discard,

      setSystems,
      setFrontends,
      setInstalledVersions,
      setInstalledFrontendVersions,
      setUpgradeAvailable,
      initFromResponse,
      refreshAfterApply,
    }),
    [
      configState,
      configReady,
      systems,
      frontends,
      installedVersions,
      installedFrontendVersions,
      enabledEmulators,
      changes,
      configChanges,
      upgradeAvailable,
      setCollection,
      setGraphicsShaders,
      setSavestateResume,
      setControllerNintendoConfirm,
      setHotkeyModifier,
      setHotkeyAction,
      resetHotkeys,
      toggleEmulator,
      setEmulatorVersion,
      setEmulatorShaders,
      setEmulatorResume,
      toggleFrontend,
      setFrontendVersion,
      enableAllSystems,
      apply,
      reapply,
      discard,
      initFromResponse,
      refreshAfterApply,
    ],
  )

  return <ConfigContext.Provider value={value}>{children}</ConfigContext.Provider>
}

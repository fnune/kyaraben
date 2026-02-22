import { useCallback, useEffect, useRef, useState } from 'react'
import * as daemon from '@/lib/daemon'
import type { SyncDiscoveredDevice, SyncStatusResponse } from '@/types/daemon'
import { SyncStateSyncing } from '@/types/daemon'

type ShowToast = (message: string, type?: 'error' | 'success' | 'info') => void

export interface UseSyncPairingResult {
  syncStatus: SyncStatusResponse | null
  discoveredDevices: SyncDiscoveredDevice[]
  connectionProgress: string | null
  connectionError: string | null
  enableError: string | null
  isEnabling: boolean
  isDiscovering: boolean
  isConnecting: boolean
  isPairing: boolean
  pairingDeviceId: string | null
  pairingCode: string | null
  lastSyncedAt: Date | null
  handleRemoveDevice: (deviceId: string) => Promise<void>
  handleConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
  handleEnableSync: () => Promise<void>
  handleResetSync: () => Promise<void>
  handleStartPairing: () => Promise<void>
  handleStopPairing: () => Promise<void>
  handleToggleGlobalDiscovery: (enabled: boolean) => Promise<void>
  clearConnectionError: () => void
  refreshSyncStatus: () => Promise<void>
  refreshDiscoveredDevices: () => Promise<void>
}

const POLL_INTERVAL_ACTIVE = 1000
const POLL_INTERVAL_SYNC_VIEW = 2000
const POLL_INTERVAL_NORMAL = 10000
const DISCOVERY_POLL_INTERVAL = 3000

export function useSyncPairing(showToast: ShowToast, isViewingSync: boolean): UseSyncPairingResult {
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null)
  const [discoveredDevices, setDiscoveredDevices] = useState<SyncDiscoveredDevice[]>([])
  const [connectionProgress, setConnectionProgress] = useState<string | null>(null)
  const [connectionError, setConnectionError] = useState<string | null>(null)
  const [enableError, setEnableError] = useState<string | null>(null)
  const [isEnabling, setIsEnabling] = useState(false)
  const [isDiscovering, setIsDiscovering] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)
  const [isPairing, setIsPairing] = useState(false)
  const [pairingDeviceId, setPairingDeviceId] = useState<string | null>(null)
  const [pairingCode, setPairingCode] = useState<string | null>(null)
  const [lastSyncedAt, setLastSyncedAt] = useState<Date | null>(null)
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const discoveryIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const previousStateRef = useRef<string | undefined>(undefined)
  const previousDeviceCountRef = useRef<number>(0)

  const refreshSyncStatus = useCallback(async () => {
    const result = await daemon.getSyncStatus()
    if (result.ok) {
      const newState = result.data.state
      const wasSyncing = previousStateRef.current === SyncStateSyncing
      const nowSynced = newState === 'synced'
      if (wasSyncing && nowSynced) {
        setLastSyncedAt(new Date())
      }
      previousStateRef.current = newState
      setSyncStatus(result.data)
    }
  }, [])

  const refreshDiscoveredDevices = useCallback(async () => {
    const result = await daemon.getDiscoveredDevices()
    if (result.ok) {
      setDiscoveredDevices(result.data.devices)
    }
  }, [])

  useEffect(() => {
    refreshSyncStatus()
  }, [refreshSyncStatus])

  useEffect(() => {
    const isSyncing = syncStatus?.state === SyncStateSyncing
    const isNotRunning = syncStatus?.enabled && !syncStatus?.running
    const isUnknown = syncStatus === null
    const isActive = isSyncing || isNotRunning || isUnknown

    let interval: number
    if (isActive) {
      interval = POLL_INTERVAL_ACTIVE
    } else if (isViewingSync) {
      interval = POLL_INTERVAL_SYNC_VIEW
    } else {
      interval = POLL_INTERVAL_NORMAL
    }

    if (pollIntervalRef.current) {
      clearInterval(pollIntervalRef.current)
    }

    pollIntervalRef.current = setInterval(refreshSyncStatus, interval)

    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
      }
    }
  }, [
    syncStatus?.enabled,
    syncStatus?.running,
    syncStatus?.state,
    syncStatus,
    isViewingSync,
    refreshSyncStatus,
  ])

  useEffect(() => {
    const hasNoDevices = (syncStatus?.devices?.length ?? 0) === 0
    const shouldDiscover = syncStatus?.enabled && syncStatus?.running && hasNoDevices

    if (discoveryIntervalRef.current) {
      clearInterval(discoveryIntervalRef.current)
      discoveryIntervalRef.current = null
    }

    if (shouldDiscover) {
      setIsDiscovering(true)
      refreshDiscoveredDevices()
      discoveryIntervalRef.current = setInterval(refreshDiscoveredDevices, DISCOVERY_POLL_INTERVAL)
    } else {
      setIsDiscovering(false)
      setDiscoveredDevices([])
    }

    return () => {
      if (discoveryIntervalRef.current) {
        clearInterval(discoveryIntervalRef.current)
      }
    }
  }, [
    syncStatus?.enabled,
    syncStatus?.running,
    syncStatus?.devices?.length,
    refreshDiscoveredDevices,
  ])

  useEffect(() => {
    return window.electron.on('pairing:progress', (data) => {
      if (data.message) {
        setConnectionProgress(data.message)
      }
    })
  }, [])

  useEffect(() => {
    const currentDeviceCount = syncStatus?.devices?.length ?? 0
    const previousDeviceCount = previousDeviceCountRef.current

    if (isPairing && currentDeviceCount > previousDeviceCount && currentDeviceCount > 0) {
      const newDevice = syncStatus?.devices?.[syncStatus.devices.length - 1]
      const deviceName = newDevice?.name || 'New device'
      showToast(`${deviceName} connected.`, 'success')
      setIsPairing(false)
      setPairingDeviceId(null)
      setPairingCode(null)
      daemon.cancelSyncPairing()
    }

    previousDeviceCountRef.current = currentDeviceCount
  }, [syncStatus?.devices, isPairing, showToast])

  const handleRemoveDevice = useCallback(
    async (deviceId: string) => {
      const result = await daemon.removeSyncDevice({ deviceId })
      if (result.ok) {
        showToast('Device removed.', 'info')
        await refreshSyncStatus()
      } else {
        showToast('Failed to remove device.', 'error')
      }
    },
    [refreshSyncStatus, showToast],
  )

  const handleConnectToDevice = useCallback(
    async (targetDeviceId: string): Promise<{ ok: boolean; error?: string }> => {
      setConnectionError(null)
      setConnectionProgress('Connecting...')
      setIsConnecting(true)
      const result = await daemon.joinSyncPeer({ code: targetDeviceId, pairingAddr: '' })
      setConnectionProgress(null)
      setIsConnecting(false)
      if (result.ok) {
        const peerName = result.data.peerName || 'peer'
        showToast(`Connected to ${peerName}.`, 'success')
        await refreshSyncStatus()
        return { ok: true }
      }
      const errorMsg = result.error?.message ?? 'Failed to connect to device'
      setConnectionError(errorMsg)
      showToast(errorMsg, 'error')
      return { ok: false, error: errorMsg }
    },
    [refreshSyncStatus, showToast],
  )

  const handleEnableSync = useCallback(async () => {
    setIsEnabling(true)
    setEnableError(null)
    try {
      const result = await daemon.enableSync({})
      if (result.ok) {
        showToast('Sync enabled.', 'success')
        await refreshSyncStatus()
      } else {
        const errorMsg = result.error?.message ?? 'Failed to enable sync'
        setEnableError(errorMsg)
        showToast(errorMsg, 'error')
      }
    } finally {
      setIsEnabling(false)
    }
  }, [refreshSyncStatus, showToast])

  const handleResetSync = useCallback(async () => {
    setEnableError(null)
    setConnectionError(null)
    setIsPairing(false)
    setPairingDeviceId(null)
    setPairingCode(null)
    const result = await daemon.resetSync()
    if (result.ok) {
      showToast('Sync reset.', 'info')
      await refreshSyncStatus()
    } else {
      showToast('Failed to reset sync.', 'error')
    }
  }, [refreshSyncStatus, showToast])

  const handleStartPairing = useCallback(async () => {
    const result = await daemon.startSyncPairing()
    if (result.ok) {
      setIsPairing(true)
      setPairingDeviceId(result.data.deviceId)
      setPairingCode(result.data.code ?? null)
      showToast('Pairing mode started.', 'info')
    } else {
      const errorMsg = result.error?.message ?? 'Failed to start pairing'
      setEnableError(errorMsg)
      showToast(errorMsg, 'error')
    }
  }, [showToast])

  const handleStopPairing = useCallback(async () => {
    await daemon.cancelSyncPairing()
    setIsPairing(false)
    setPairingDeviceId(null)
    setPairingCode(null)
  }, [])

  const handleToggleGlobalDiscovery = useCallback(
    async (enabled: boolean) => {
      const result = await daemon.setSyncSettings({ globalDiscoveryEnabled: enabled })
      if (result.ok) {
        await refreshSyncStatus()
      } else {
        showToast('Failed to update global discovery setting.', 'error')
      }
    },
    [refreshSyncStatus, showToast],
  )

  const clearConnectionError = useCallback(() => {
    setConnectionError(null)
  }, [])

  return {
    syncStatus,
    discoveredDevices,
    connectionProgress,
    connectionError,
    enableError,
    isEnabling,
    isDiscovering,
    isConnecting,
    isPairing,
    pairingDeviceId,
    pairingCode,
    lastSyncedAt,
    handleRemoveDevice,
    handleConnectToDevice,
    handleEnableSync,
    handleResetSync,
    handleStartPairing,
    handleStopPairing,
    handleToggleGlobalDiscovery,
    clearConnectionError,
    refreshSyncStatus,
    refreshDiscoveredDevices,
  }
}

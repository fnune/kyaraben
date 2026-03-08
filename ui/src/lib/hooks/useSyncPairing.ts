import type { SyncDiscoveredDevice, SyncStatusResponse } from '@shared/daemon'
import { SyncStateSyncing } from '@shared/daemon'
import { useCallback, useEffect, useRef, useState } from 'react'
import * as daemon from '@/lib/daemon'

type ShowToast = (message: string, type?: 'error' | 'success' | 'info') => void

export interface PendingDevice {
  deviceId: string
  name: string
}

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
  pendingDevice: PendingDevice | null
  handleRemoveDevice: (deviceId: string) => Promise<void>
  handleConnectToDevice: (deviceId: string) => Promise<{ ok: boolean; error?: string }>
  handleEnableSync: () => Promise<void>
  handleResetSync: () => Promise<void>
  handleStartPairing: () => Promise<void>
  handleStopPairing: () => Promise<void>
  handleToggleGlobalDiscovery: (enabled: boolean) => Promise<void>
  handleToggleRunning: (running: boolean) => Promise<void>
  handleToggleAutostart: (enabled: boolean) => Promise<void>
  handleAcceptDevice: (accept: boolean) => Promise<void>
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
  const [pendingDevice, setPendingDevice] = useState<PendingDevice | null>(null)
  const [expectedPairedDevice, setExpectedPairedDevice] = useState<string | null>(null)
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const discoveryIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const previousStateRef = useRef<string | undefined>(undefined)
  const devicesAtPairingStartRef = useRef<Set<string>>(new Set())

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
      if (data.deviceId) {
        setExpectedPairedDevice(data.deviceId)
      }
    })
  }, [])

  useEffect(() => {
    return window.electron.on('pairing:pending_device', (data) => {
      const { deviceId, name } = data as { deviceId: string; name: string }
      setPendingDevice({ deviceId, name })
    })
  }, [])

  useEffect(() => {
    if (!isPairing) {
      return
    }

    const currentDevices = syncStatus?.devices ?? []
    let pairedDevice = null

    if (expectedPairedDevice) {
      pairedDevice = currentDevices.find((d) => d.id === expectedPairedDevice)
    } else {
      pairedDevice = currentDevices.find((d) => !devicesAtPairingStartRef.current.has(d.id))
    }

    if (pairedDevice) {
      const deviceName = pairedDevice.name || 'New device'
      showToast(`${deviceName} connected.`, 'success')
      setIsPairing(false)
      setPairingDeviceId(null)
      setPairingCode(null)
      setExpectedPairedDevice(null)
      devicesAtPairingStartRef.current = new Set()
      daemon.cancelSyncPairing()
    }
  }, [syncStatus?.devices, isPairing, expectedPairedDevice, showToast])

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
    devicesAtPairingStartRef.current = new Set(syncStatus?.devices?.map((d) => d.id) ?? [])
    const result = await daemon.startSyncPairing()
    if (result.ok) {
      setIsPairing(true)
      setPairingDeviceId(result.data.deviceId)
      setPairingCode(result.data.code ?? null)
      showToast('Pairing mode started.', 'info')
    } else {
      devicesAtPairingStartRef.current = new Set()
      const errorMsg = result.error?.message ?? 'Failed to start pairing'
      setEnableError(errorMsg)
      showToast(errorMsg, 'error')
    }
  }, [showToast, syncStatus?.devices])

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

  const handleToggleRunning = useCallback(
    async (running: boolean) => {
      const result = await daemon.setSyncSettings({ running })
      if (result.ok) {
        await refreshSyncStatus()
      } else {
        showToast(`Failed to ${running ? 'start' : 'stop'} syncthing.`, 'error')
      }
    },
    [refreshSyncStatus, showToast],
  )

  const handleToggleAutostart = useCallback(
    async (enabled: boolean) => {
      const result = await daemon.setSyncSettings({ autostartEnabled: enabled })
      if (result.ok) {
        await refreshSyncStatus()
      } else {
        showToast('Failed to update autostart setting.', 'error')
      }
    },
    [refreshSyncStatus, showToast],
  )

  const handleAcceptDevice = useCallback(
    async (accept: boolean) => {
      if (!pendingDevice) return
      const result = await daemon.acceptSyncDevice({
        deviceId: pendingDevice.deviceId,
        accept,
      })
      if (result.ok) {
        if (accept) {
          showToast(`${pendingDevice.name || 'Device'} accepted.`, 'success')
        } else {
          showToast(`${pendingDevice.name || 'Device'} rejected.`, 'info')
        }
        await refreshSyncStatus()
      } else {
        showToast(`Failed to ${accept ? 'accept' : 'reject'} device.`, 'error')
      }
      setPendingDevice(null)
    },
    [pendingDevice, refreshSyncStatus, showToast],
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
    pendingDevice,
    handleRemoveDevice,
    handleConnectToDevice,
    handleEnableSync,
    handleResetSync,
    handleStartPairing,
    handleStopPairing,
    handleToggleGlobalDiscovery,
    handleToggleRunning,
    handleToggleAutostart,
    handleAcceptDevice,
    clearConnectionError,
    refreshSyncStatus,
    refreshDiscoveredDevices,
  }
}

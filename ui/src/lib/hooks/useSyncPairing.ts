import { useCallback, useEffect, useRef, useState } from 'react'
import * as daemon from '@/lib/daemon'
import type { SyncMode, SyncStatusResponse } from '@/types/daemon'
import { SyncStateSyncing } from '@/types/daemon'

export interface UseSyncPairingResult {
  syncStatus: SyncStatusResponse | null
  pairingCode: string | null
  pairingProgress: string | null
  isEnabling: boolean
  handleAddDevice: (deviceId: string, name: string) => Promise<void>
  handleRemoveDevice: (deviceId: string) => Promise<void>
  handleStartPairing: () => Promise<void>
  handleCancelPairing: () => Promise<void>
  handleJoinPrimary: (code: string, pairingAddr: string) => Promise<void>
  handlePauseSync: () => Promise<void>
  handleResumeSync: () => Promise<void>
  handleEnableSync: (mode: SyncMode) => Promise<void>
  refreshSyncStatus: () => Promise<void>
}

const POLL_INTERVAL_SYNCING = 2000
const POLL_INTERVAL_NORMAL = 10000

export function useSyncPairing(): UseSyncPairingResult {
  const [syncStatus, setSyncStatus] = useState<SyncStatusResponse | null>(null)
  const [pairingCode, setPairingCode] = useState<string | null>(null)
  const [pairingProgress, setPairingProgress] = useState<string | null>(null)
  const [isEnabling, setIsEnabling] = useState(false)
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const refreshSyncStatus = useCallback(async () => {
    const result = await daemon.getSyncStatus()
    if (result.ok) {
      setSyncStatus(result.data)
    }
  }, [])

  useEffect(() => {
    const isSyncing = syncStatus?.state === SyncStateSyncing
    const interval = isSyncing ? POLL_INTERVAL_SYNCING : POLL_INTERVAL_NORMAL

    if (pollIntervalRef.current) {
      clearInterval(pollIntervalRef.current)
    }

    if (syncStatus?.enabled && syncStatus?.running) {
      pollIntervalRef.current = setInterval(refreshSyncStatus, interval)
    }

    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
      }
    }
  }, [syncStatus?.enabled, syncStatus?.running, syncStatus?.state, refreshSyncStatus])

  useEffect(() => {
    return window.electron.on('pairing:progress', (data) => {
      const msg = data.message
      if (msg) {
        if (msg.startsWith('Pairing code: ')) {
          setPairingCode(msg.replace('Pairing code: ', ''))
        }
        setPairingProgress(msg)
      }
    })
  }, [])

  const handleAddDevice = useCallback(
    async (deviceId: string, name: string) => {
      const result = await daemon.addSyncDevice({ deviceId, name })
      if (result.ok) {
        await refreshSyncStatus()
      }
    },
    [refreshSyncStatus],
  )

  const handleRemoveDevice = useCallback(
    async (deviceId: string) => {
      const result = await daemon.removeSyncDevice({ deviceId })
      if (result.ok) {
        await refreshSyncStatus()
      }
    },
    [refreshSyncStatus],
  )

  const handleStartPairing = useCallback(async () => {
    setPairingProgress('Starting pairing...')
    const result = await daemon.startSyncPairing()
    if (result.ok) {
      setPairingCode(null)
      setPairingProgress(null)
      await refreshSyncStatus()
    } else {
      setPairingProgress(null)
    }
  }, [refreshSyncStatus])

  const handleCancelPairing = useCallback(async () => {
    await daemon.cancelSyncPairing()
    setPairingCode(null)
    setPairingProgress(null)
  }, [])

  const handleJoinPrimary = useCallback(
    async (code: string, pairingAddr: string) => {
      const result = await daemon.joinSyncPrimary({ code, pairingAddr })
      if (result.ok) {
        await refreshSyncStatus()
      }
    },
    [refreshSyncStatus],
  )

  const handlePauseSync = useCallback(async () => {
    const result = await daemon.pauseSync()
    if (result.ok) {
      await refreshSyncStatus()
    }
  }, [refreshSyncStatus])

  const handleResumeSync = useCallback(async () => {
    const result = await daemon.resumeSync()
    if (result.ok) {
      await refreshSyncStatus()
    }
  }, [refreshSyncStatus])

  const handleEnableSync = useCallback(
    async (mode: SyncMode) => {
      setIsEnabling(true)
      try {
        const result = await daemon.enableSync({ mode })
        if (result.ok) {
          await refreshSyncStatus()
        }
      } finally {
        setIsEnabling(false)
      }
    },
    [refreshSyncStatus],
  )

  return {
    syncStatus,
    pairingCode,
    pairingProgress,
    isEnabling,
    handleAddDevice,
    handleRemoveDevice,
    handleStartPairing,
    handleCancelPairing,
    handleJoinPrimary,
    handlePauseSync,
    handleResumeSync,
    handleEnableSync,
    refreshSyncStatus,
  }
}

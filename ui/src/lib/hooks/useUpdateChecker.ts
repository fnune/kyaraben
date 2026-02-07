import { useCallback, useEffect, useState } from 'react'
import type { UpdateInfo } from '@/lib/daemon'
import * as daemon from '@/lib/daemon'

const UPDATE_DISMISSED_KEY = 'kyaraben:update-dismissed'
const APPLY_BANNER_DISMISSED_KEY = 'kyaraben:apply-banner-dismissed'

export interface UseUpdateCheckerResult {
  updateInfo: UpdateInfo | null
  isDownloading: boolean
  downloadProgress: number
  updateDismissed: boolean
  showApplyBanner: boolean
  applyBannerDismissed: boolean
  handleUpdate: () => Promise<void>
  handleDismissUpdate: () => void
  handleDismissApplyBanner: () => void
  handleApplyFromBanner: () => void
  setShowApplyBanner: (show: boolean) => void
  clearApplyBannerDismissal: () => void
}

export function useUpdateChecker(
  showToast: (message: string, variant: 'success' | 'error') => void,
  setCurrentView: (view: 'systems' | 'installation' | 'sync') => void,
): UseUpdateCheckerResult {
  const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null)
  const [isDownloading, setIsDownloading] = useState(false)
  const [downloadProgress, setDownloadProgress] = useState(0)
  const [updateDismissed, setUpdateDismissed] = useState(
    () => localStorage.getItem(UPDATE_DISMISSED_KEY) !== null,
  )
  const [showApplyBanner, setShowApplyBanner] = useState(false)
  const [applyBannerDismissed, setApplyBannerDismissed] = useState(
    () => localStorage.getItem(APPLY_BANNER_DISMISSED_KEY) !== null,
  )

  useEffect(() => {
    return window.electron.on('update:progress', (data) => {
      setDownloadProgress(data.percent)
    })
  }, [])

  useEffect(() => {
    const timer = setTimeout(async () => {
      const result = await daemon.checkForUpdates()
      if (result.ok) {
        setUpdateInfo(result.data)
        if (result.data.available) {
          localStorage.removeItem(UPDATE_DISMISSED_KEY)
          setUpdateDismissed(false)
        }
      }
    }, 5000)
    return () => clearTimeout(timer)
  }, [])

  const handleUpdate = useCallback(async () => {
    if (!updateInfo?.downloadUrl) return

    setIsDownloading(true)
    setDownloadProgress(0)

    const downloadResult = await daemon.downloadUpdate(updateInfo.downloadUrl)
    if (!downloadResult.ok) {
      showToast(downloadResult.error.message || 'Download failed', 'error')
      setIsDownloading(false)
      return
    }
    if (!downloadResult.data.success || !downloadResult.data.path) {
      showToast(downloadResult.data.error || 'Download failed', 'error')
      setIsDownloading(false)
      return
    }

    const applyResult = await daemon.applyUpdate(downloadResult.data.path)
    if (!applyResult.ok) {
      showToast(applyResult.error.message || 'Update failed', 'error')
      setIsDownloading(false)
      return
    }
    if (!applyResult.data.success) {
      showToast(applyResult.data.error || 'Update failed', 'error')
      setIsDownloading(false)
    }
  }, [updateInfo, showToast])

  const handleDismissUpdate = useCallback(() => {
    localStorage.setItem(UPDATE_DISMISSED_KEY, 'true')
    setUpdateDismissed(true)
  }, [])

  const handleDismissApplyBanner = useCallback(() => {
    localStorage.setItem(APPLY_BANNER_DISMISSED_KEY, 'true')
    setApplyBannerDismissed(true)
  }, [])

  const handleApplyFromBanner = useCallback(() => {
    setCurrentView('systems')
    setApplyBannerDismissed(true)
  }, [setCurrentView])

  const clearApplyBannerDismissal = useCallback(() => {
    setShowApplyBanner(false)
    localStorage.removeItem(APPLY_BANNER_DISMISSED_KEY)
    setApplyBannerDismissed(false)
  }, [])

  return {
    updateInfo,
    isDownloading,
    downloadProgress,
    updateDismissed,
    showApplyBanner,
    applyBannerDismissed,
    handleUpdate,
    handleDismissUpdate,
    handleDismissApplyBanner,
    handleApplyFromBanner,
    setShowApplyBanner,
    clearApplyBannerDismissal,
  }
}

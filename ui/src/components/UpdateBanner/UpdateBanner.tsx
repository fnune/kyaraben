import { Button } from '@/lib/Button'
import type { UpdateInfo } from '@/lib/daemon'

export interface UpdateBannerProps {
  readonly updateInfo: UpdateInfo
  readonly onUpdate: () => void
  readonly onDismiss: () => void
  readonly isDownloading: boolean
  readonly downloadProgress: number
}

export function UpdateBanner({
  updateInfo,
  onUpdate,
  onDismiss,
  isDownloading,
  downloadProgress,
}: UpdateBannerProps) {
  if (!updateInfo.available) return null

  return (
    <div className="bg-blue-900/50 border-b border-blue-700/50 px-4 py-3">
      <div className="flex items-center justify-between gap-4">
        <div className="flex-1 min-w-0">
          <p className="text-sm text-blue-100">
            A new version of Kyaraben is available: {updateInfo.latestVersion}. You can also update
            from the Installation tab.
          </p>
          {isDownloading && (
            <div className="mt-2">
              <div className="h-1.5 bg-blue-950 rounded-full overflow-hidden">
                <div
                  className="h-full bg-blue-400 transition-all duration-200"
                  style={{ width: `${downloadProgress}%` }}
                />
              </div>
              <p className="text-xs text-blue-300 mt-1">Downloading... {downloadProgress}%</p>
            </div>
          )}
        </div>
        <div className="flex items-center gap-2 shrink-0">
          {!isDownloading && (
            <>
              <Button variant="secondary" onClick={onDismiss}>
                Dismiss
              </Button>
              <Button onClick={onUpdate}>Update now</Button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

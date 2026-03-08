import { ActionBanner } from '@/lib/ActionBanner'
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
    <ActionBanner
      variant="accent"
      title={
        <>
          A new version of Kyaraben is available: {updateInfo.latestVersion}. You can also update
          from the Installation tab.
        </>
      }
      description={
        isDownloading ? (
          <div className="mt-2">
            <div className="h-1.5 bg-accent-muted rounded-full overflow-hidden">
              <div
                className="h-full bg-accent transition-all duration-200"
                style={{ width: `${downloadProgress}%` }}
              />
            </div>
            <p className="text-xs text-accent mt-1">Downloading... {downloadProgress}%</p>
          </div>
        ) : undefined
      }
      actions={
        !isDownloading ? (
          <>
            <Button variant="secondary" onClick={onDismiss}>
              Dismiss
            </Button>
            <Button onClick={onUpdate}>Update now</Button>
          </>
        ) : null
      }
    />
  )
}

import { useCallback, useState } from 'react'
import { Button } from '@/lib/Button'
import { openUrl } from '@/lib/daemon'

export interface SyncSettingsSectionProps {
  readonly guiURL: string | undefined
  readonly onReset: () => Promise<void>
}

export function SyncSettingsSection({ guiURL, onReset }: SyncSettingsSectionProps) {
  const [isCollapsed, setIsCollapsed] = useState(true)
  const [showResetConfirm, setShowResetConfirm] = useState(false)
  const [isResetting, setIsResetting] = useState(false)

  const handleReset = useCallback(async () => {
    setIsResetting(true)
    try {
      await onReset()
    } finally {
      setIsResetting(false)
      setShowResetConfirm(false)
    }
  }, [onReset])

  return (
    <div className="p-4 bg-surface-alt rounded-card">
      <button
        type="button"
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="flex items-center gap-2 w-full text-left"
      >
        <span
          className={`transition-transform text-xs text-on-surface-muted ${isCollapsed ? '' : 'rotate-90'}`}
        >
          ▶
        </span>
        <h3 className="text-sm font-medium text-on-surface">Settings</h3>
      </button>
      {!isCollapsed && (
        <div className="mt-3 space-y-4">
          {guiURL && (
            <button
              type="button"
              onClick={() => openUrl(guiURL)}
              className="text-sm text-accent hover:underline block text-left"
            >
              Open Syncthing web interface
            </button>
          )}
          <div>
            <h4 className="text-sm font-medium text-on-surface mb-2">Reset sync</h4>
            {showResetConfirm ? (
              <div className="space-y-3">
                <div className="p-3 bg-status-warning/10 border border-status-warning/30 rounded text-sm">
                  <p className="text-on-surface mb-2">This will:</p>
                  <ul className="list-disc list-inside text-on-surface-muted space-y-1">
                    <li>Stop and remove the syncthing service</li>
                    <li>Delete syncthing configuration and database</li>
                    <li>Remove all device pairings</li>
                    <li>Disable sync in your kyaraben config</li>
                  </ul>
                  <p className="mt-2 text-on-surface-muted">
                    Your ROMs, saves, and other emulation data will not be affected. You can
                    re-enable sync afterwards.
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button variant="danger" onClick={handleReset} disabled={isResetting}>
                    {isResetting ? 'Resetting...' : 'Confirm reset'}
                  </Button>
                  <Button variant="secondary" onClick={() => setShowResetConfirm(false)}>
                    Cancel
                  </Button>
                </div>
              </div>
            ) : (
              <div>
                <p className="text-sm text-on-surface-muted mb-2">
                  Remove all syncthing state and start fresh.
                </p>
                <Button variant="secondary" onClick={() => setShowResetConfirm(true)}>
                  Reset sync
                </Button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

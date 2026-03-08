import { useCallback, useState } from 'react'
import { Button } from '@/lib/Button'
import { useOpenUrl } from '@/lib/hooks/useOpenUrl'
import { ToggleSwitch } from '@/lib/ToggleSwitch'

export interface SyncSettingsSectionProps {
  readonly guiURL: string | undefined
  readonly globalDiscoveryEnabled: boolean
  readonly running: boolean
  readonly autostartEnabled: boolean
  readonly onToggleGlobalDiscovery: (enabled: boolean) => Promise<void>
  readonly onToggleRunning: (running: boolean) => Promise<void>
  readonly onToggleAutostart: (enabled: boolean) => Promise<void>
  readonly onReset: () => Promise<void>
}

export function SyncSettingsSection({
  guiURL,
  globalDiscoveryEnabled,
  running,
  autostartEnabled,
  onToggleGlobalDiscovery,
  onToggleRunning,
  onToggleAutostart,
  onReset,
}: SyncSettingsSectionProps) {
  const [isCollapsed, setIsCollapsed] = useState(true)
  const [showResetConfirm, setShowResetConfirm] = useState(false)
  const [isResetting, setIsResetting] = useState(false)
  const [isTogglingDiscovery, setIsTogglingDiscovery] = useState(false)
  const [isTogglingRunning, setIsTogglingRunning] = useState(false)
  const [isTogglingAutostart, setIsTogglingAutostart] = useState(false)
  const openUrl = useOpenUrl()

  const handleReset = useCallback(async () => {
    setIsResetting(true)
    try {
      await onReset()
    } finally {
      setIsResetting(false)
      setShowResetConfirm(false)
    }
  }, [onReset])

  const handleToggleGlobalDiscovery = useCallback(
    async (enabled: boolean) => {
      setIsTogglingDiscovery(true)
      try {
        await onToggleGlobalDiscovery(enabled)
      } finally {
        setIsTogglingDiscovery(false)
      }
    },
    [onToggleGlobalDiscovery],
  )

  const handleToggleRunning = useCallback(
    async (enabled: boolean) => {
      setIsTogglingRunning(true)
      try {
        await onToggleRunning(enabled)
      } finally {
        setIsTogglingRunning(false)
      }
    },
    [onToggleRunning],
  )

  const handleToggleAutostart = useCallback(
    async (enabled: boolean) => {
      setIsTogglingAutostart(true)
      try {
        await onToggleAutostart(enabled)
      } finally {
        setIsTogglingAutostart(false)
      }
    },
    [onToggleAutostart],
  )

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
          <div className="flex items-center justify-between">
            <div>
              <label htmlFor="syncing-toggle" className="text-sm font-medium text-on-surface">
                Syncing
              </label>
              <p className="text-xs text-on-surface-muted mt-0.5">
                Enable or disable synchronization.
              </p>
            </div>
            <ToggleSwitch
              enabled={running}
              onChange={handleToggleRunning}
              disabled={isTogglingRunning}
            />
          </div>
          <div className="flex items-center justify-between">
            <div>
              <label htmlFor="autostart-toggle" className="text-sm font-medium text-on-surface">
                Start on boot
              </label>
              <p className="text-xs text-on-surface-muted mt-0.5">
                Automatically start Syncthing when you log in. Works in Game Mode on Steam Deck.
              </p>
            </div>
            <ToggleSwitch
              enabled={autostartEnabled}
              onChange={handleToggleAutostart}
              disabled={isTogglingAutostart}
            />
          </div>
          <div className="flex items-center justify-between">
            <div>
              <label
                htmlFor="global-discovery-toggle"
                className="text-sm font-medium text-on-surface"
              >
                Global discovery
              </label>
              <p className="text-xs text-on-surface-muted mt-0.5">
                Announce this device to remote peers. Only needed for syncing across networks.
              </p>
            </div>
            <ToggleSwitch
              enabled={globalDiscoveryEnabled}
              onChange={handleToggleGlobalDiscovery}
              disabled={isTogglingDiscovery}
            />
          </div>
          <div>
            <h4 className="text-sm font-medium text-on-surface mb-2">Reset synchronization</h4>
            {showResetConfirm ? (
              <div className="space-y-3">
                <div className="p-3 bg-status-warning/10 border border-status-warning/30 rounded text-sm">
                  <p className="text-on-surface mb-2">This will:</p>
                  <ul className="list-disc list-inside text-on-surface-muted space-y-1">
                    <li>Stop and remove the Syncthing service</li>
                    <li>Delete Syncthing configuration and database</li>
                    <li>Remove all device pairings</li>
                    <li>Disable synchronization in your Kyaraben config</li>
                  </ul>
                  <p className="mt-2 text-on-surface-muted">
                    Your ROMs, saves, and other emulation data will not be affected. You can
                    re-enable synchronization afterwards.
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
                  Remove all Syncthing state and start fresh.
                </p>
                <Button variant="secondary" onClick={() => setShowResetConfirm(true)}>
                  Reset synchronization
                </Button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

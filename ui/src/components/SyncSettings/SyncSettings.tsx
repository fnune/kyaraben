import { useState } from 'react'
import type { SyncDevice, SyncStatusResponse } from '@/types/daemon'

export interface SyncSettingsProps {
  readonly status: SyncStatusResponse | null
  readonly onAddDevice: (deviceId: string, name: string) => Promise<void>
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
  readonly onClose: () => void
}

function truncateDeviceId(id: string): string {
  if (id.length > 20) {
    return `${id.slice(0, 7)}...${id.slice(-7)}`
  }
  return id
}

function DeviceRow({
  device,
  onRemove,
}: {
  readonly device: SyncDevice
  readonly onRemove: () => void
}) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0">
      <div className="flex items-center gap-2">
        <span
          className={`w-2 h-2 rounded-full ${device.connected ? 'bg-green-500' : 'bg-gray-300'}`}
        />
        <span className="font-medium text-gray-900">{device.name}</span>
        <span className="text-xs text-gray-400">{truncateDeviceId(device.id)}</span>
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="text-xs text-red-600 hover:text-red-800"
      >
        Remove
      </button>
    </div>
  )
}

export function SyncSettings({ status, onAddDevice, onRemoveDevice, onClose }: SyncSettingsProps) {
  const [newDeviceId, setNewDeviceId] = useState('')
  const [newDeviceName, setNewDeviceName] = useState('')
  const [isAdding, setIsAdding] = useState(false)

  const handleAddDevice = async () => {
    if (!newDeviceId.trim()) return
    setIsAdding(true)
    try {
      await onAddDevice(newDeviceId.trim(), newDeviceName.trim() || undefined!)
      setNewDeviceId('')
      setNewDeviceName('')
    } finally {
      setIsAdding(false)
    }
  }

  const handleOpenGui = () => {
    if (status?.guiURL) {
      window.open(status.guiURL, '_blank')
    }
  }

  if (!status?.enabled) {
    return (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Sync Settings</h2>
          <p className="text-gray-600 mb-4">
            Sync is not enabled. Enable it in your config.toml:
          </p>
          <pre className="bg-gray-100 p-3 rounded text-sm mb-4">
{`[sync]
enabled = true
mode = "primary"  # or "secondary"`}
          </pre>
          <button
            type="button"
            onClick={onClose}
            className="w-full py-2 bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
          >
            Close
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-lg w-full mx-4 p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">Sync Settings</h2>
          <button
            type="button"
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            &times;
          </button>
        </div>

        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-600">Mode</span>
            <span className="text-sm font-medium capitalize">{status.mode}</span>
          </div>

          {status.deviceId && (
            <div>
              <span className="text-sm text-gray-600 block mb-1">This device ID</span>
              <div className="flex items-center gap-2">
                <code className="flex-1 bg-gray-100 px-3 py-2 rounded text-xs break-all">
                  {status.deviceId}
                </code>
                <button
                  type="button"
                  onClick={() => navigator.clipboard.writeText(status.deviceId!)}
                  className="px-3 py-2 text-xs bg-gray-100 rounded hover:bg-gray-200"
                >
                  Copy
                </button>
              </div>
            </div>
          )}

          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-gray-600">Paired devices</span>
              {status.guiURL && (
                <button
                  type="button"
                  onClick={handleOpenGui}
                  className="text-xs text-blue-600 hover:text-blue-800"
                >
                  Open Syncthing UI
                </button>
              )}
            </div>

            {status.devices && status.devices.length > 0 ? (
              <div className="border rounded-lg px-3">
                {status.devices.map((device) => (
                  <DeviceRow
                    key={device.id}
                    device={device}
                    onRemove={() => onRemoveDevice(device.id)}
                  />
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500 italic">No devices paired</p>
            )}
          </div>

          <div className="border-t pt-4">
            <span className="text-sm text-gray-600 block mb-2">Add a device</span>
            <div className="space-y-2">
              <input
                type="text"
                value={newDeviceId}
                onChange={(e) => setNewDeviceId(e.target.value)}
                placeholder="Device ID"
                className="w-full px-3 py-2 border rounded text-sm"
              />
              <input
                type="text"
                value={newDeviceName}
                onChange={(e) => setNewDeviceName(e.target.value)}
                placeholder="Friendly name (optional)"
                className="w-full px-3 py-2 border rounded text-sm"
              />
              <button
                type="button"
                onClick={handleAddDevice}
                disabled={!newDeviceId.trim() || isAdding}
                className="w-full py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
              >
                {isAdding ? 'Adding...' : 'Add Device'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

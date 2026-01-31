import { useState } from 'react'
import { Button } from '@/lib/Button'
import { Input } from '@/lib/Input'
import { Modal } from '@/lib/Modal'
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
      <button type="button" onClick={onRemove} className="text-xs text-red-600 hover:text-red-800">
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
      await onAddDevice(newDeviceId.trim(), newDeviceName.trim())
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
      <Modal open={true} onClose={onClose} title="Sync settings">
        <p className="text-gray-600 mb-4">Sync is not enabled. Enable it in your config.toml:</p>
        <pre className="bg-gray-100 p-3 rounded text-sm mb-4">
          {`[sync]
enabled = true
mode = "primary"  # or "secondary"`}
        </pre>
        <Button variant="secondary" onClick={onClose}>
          Close
        </Button>
      </Modal>
    )
  }

  return (
    <Modal open={true} onClose={onClose} title="Sync settings">
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
                onClick={() => {
                  if (status.deviceId) {
                    navigator.clipboard.writeText(status.deviceId)
                  }
                }}
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
            <Input value={newDeviceId} onChange={setNewDeviceId} placeholder="Device ID" />
            <Input
              value={newDeviceName}
              onChange={setNewDeviceName}
              placeholder="Friendly name (optional)"
            />
            <Button onClick={handleAddDevice} disabled={!newDeviceId.trim() || isAdding}>
              {isAdding ? 'Adding...' : 'Add device'}
            </Button>
          </div>
        </div>
      </div>
    </Modal>
  )
}

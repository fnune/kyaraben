import { useState } from 'react'
import { Button } from '@/lib/Button'
import { Input } from '@/lib/Input'
import type { SyncDevice, SyncStatusResponse } from '@/types/daemon'

export interface SyncViewProps {
  readonly status: SyncStatusResponse | null
  readonly onAddDevice: (deviceId: string, name: string) => Promise<void>
  readonly onRemoveDevice: (deviceId: string) => Promise<void>
}

function truncateDeviceId(id: string): string {
  if (id.length > 20) {
    return `${id.slice(0, 7)}...${id.slice(-7)}`
  }
  return id
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="p-4 bg-surface-alt rounded-card">
      <h3 className="text-sm font-medium text-on-surface mb-3">{title}</h3>
      {children}
    </div>
  )
}

function DeviceRow({
  device,
  onRemove,
}: {
  readonly device: SyncDevice
  readonly onRemove: () => void
}) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-outline last:border-0">
      <div className="flex items-center gap-2">
        <span
          className={`w-2 h-2 rounded-full ${device.connected ? 'bg-status-ok' : 'bg-outline'}`}
        />
        <span className="font-medium text-on-surface">{device.name}</span>
        <span className="text-xs text-on-surface-dim">{truncateDeviceId(device.id)}</span>
      </div>
      <button
        type="button"
        onClick={onRemove}
        className="text-xs text-status-error hover:text-status-error"
      >
        Remove
      </button>
    </div>
  )
}

function DisabledState() {
  return (
    <div className="p-6">
      <Section title="Sync is not enabled">
        <p className="text-sm text-on-surface-muted mb-4">Enable sync in your config.toml:</p>
        <pre className="bg-surface-raised text-on-surface-secondary p-3 rounded-sm text-sm font-mono">
          {`[sync]
enabled = true
mode = "primary"  # or "secondary"`}
        </pre>
      </Section>
    </div>
  )
}

export function SyncView({ status, onAddDevice, onRemoveDevice }: SyncViewProps) {
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
    return <DisabledState />
  }

  const connectedCount = status.devices?.filter((d) => d.connected).length ?? 0
  const totalDevices = status.devices?.length ?? 0

  return (
    <div className="p-6 space-y-6">
      <Section title="Status">
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-on-surface-muted">Mode</span>
            <span className="font-medium text-on-surface capitalize">{status.mode}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-on-surface-muted">Running</span>
            <span
              className={`font-medium ${status.running ? 'text-status-ok' : 'text-status-error'}`}
            >
              {status.running ? 'Yes' : 'No'}
            </span>
          </div>
          {totalDevices > 0 && (
            <div className="flex justify-between">
              <span className="text-on-surface-muted">Connected devices</span>
              <span className="font-medium text-on-surface">
                {connectedCount}/{totalDevices}
              </span>
            </div>
          )}
        </div>
      </Section>

      {status.deviceId && (
        <Section title="This device ID">
          <div className="flex items-center gap-2">
            <code className="flex-1 bg-surface-raised text-on-surface-secondary px-3 py-2 rounded-sm text-xs break-all font-mono">
              {status.deviceId}
            </code>
            <button
              type="button"
              onClick={() => {
                if (status.deviceId) {
                  navigator.clipboard.writeText(status.deviceId)
                }
              }}
              className="px-3 py-2 text-xs bg-surface-raised text-on-surface-secondary rounded-sm hover:bg-outline"
            >
              Copy
            </button>
          </div>
        </Section>
      )}

      <Section title="Paired devices">
        <div className="flex items-center justify-between mb-3">
          <span className="text-sm text-on-surface-muted">
            {totalDevices === 0
              ? 'No devices paired'
              : `${totalDevices} device${totalDevices === 1 ? '' : 's'}`}
          </span>
          {status.guiURL && (
            <button
              type="button"
              onClick={handleOpenGui}
              className="text-xs text-accent hover:text-accent-hover"
            >
              Open Syncthing UI
            </button>
          )}
        </div>

        {status.devices && status.devices.length > 0 && (
          <div className="border border-outline rounded-card px-3 bg-surface">
            {status.devices.map((device) => (
              <DeviceRow
                key={device.id}
                device={device}
                onRemove={() => onRemoveDevice(device.id)}
              />
            ))}
          </div>
        )}
      </Section>

      <Section title="Add a device">
        <div className="space-y-3">
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
      </Section>
    </div>
  )
}

import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { SyncFolder } from '@/types/daemon'
import { ActivityCard } from './ActivityCard'

function createFolder(overrides: Partial<SyncFolder> = {}): SyncFolder {
  return {
    id: 'folder-1',
    path: '/path/to/folder',
    label: 'Test Folder',
    state: 'idle',
    type: 'sendreceive',
    globalSize: 0,
    localSize: 0,
    needSize: 0,
    receiveOnlyChanges: 0,
    ...overrides,
  }
}

describe('ActivityCard', () => {
  describe('when no paired devices', () => {
    it('shows waiting message', () => {
      render(
        <ActivityCard
          state="disconnected"
          folders={[]}
          lastSyncedAt={null}
          hasPairedDevices={false}
        />,
      )

      expect(screen.getByText('Waiting for device connection')).toBeInTheDocument()
    })
  })

  describe('when all synced', () => {
    it('shows all synced message', () => {
      render(
        <ActivityCard
          state="synced"
          folders={[createFolder()]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('All synced')).toBeInTheDocument()
    })

    it('shows last sync time when available', () => {
      const tenMinutesAgo = new Date(Date.now() - 10 * 60 * 1000)
      render(
        <ActivityCard
          state="synced"
          folders={[createFolder()]}
          lastSyncedAt={tenMinutesAgo}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('Last sync: 10m ago')).toBeInTheDocument()
    })
  })

  describe('scanning progress', () => {
    it('shows shimmer when no progress data available', () => {
      render(
        <ActivityCard
          state="syncing"
          folders={[createFolder({ state: 'scanning', globalSize: 0, needSize: 0 })]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('Test Folder')).toBeInTheDocument()
      expect(screen.getByText('Scanning...')).toBeInTheDocument()
    })

    it('shows progress bar when progress data is available', () => {
      render(
        <ActivityCard
          state="syncing"
          folders={[
            createFolder({
              state: 'scanning',
              globalSize: 100 * 1024 * 1024,
              needSize: 25 * 1024 * 1024,
            }),
          ]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('Test Folder')).toBeInTheDocument()
      expect(screen.getByText('Scanning 25 MB remaining')).toBeInTheDocument()
    })

    it('shows queue count when multiple folders are scanning', () => {
      render(
        <ActivityCard
          state="syncing"
          folders={[
            createFolder({ id: 'folder-1', label: 'Folder 1', state: 'scanning' }),
            createFolder({ id: 'folder-2', label: 'Folder 2', state: 'scanning' }),
            createFolder({ id: 'folder-3', label: 'Folder 3', state: 'scanning' }),
          ]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('+2 folders in queue')).toBeInTheDocument()
    })

    it('uses singular form for single folder in queue', () => {
      render(
        <ActivityCard
          state="syncing"
          folders={[
            createFolder({ id: 'folder-1', label: 'Folder 1', state: 'scanning' }),
            createFolder({ id: 'folder-2', label: 'Folder 2', state: 'scanning' }),
          ]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('+1 folder in queue')).toBeInTheDocument()
    })
  })

  describe('syncing progress', () => {
    it('shows progress with remaining bytes', () => {
      render(
        <ActivityCard
          state="syncing"
          folders={[
            createFolder({
              state: 'syncing',
              globalSize: 100 * 1024 * 1024,
              needSize: 40 * 1024 * 1024,
            }),
          ]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('Test Folder')).toBeInTheDocument()
      expect(screen.getByText('40 MB remaining')).toBeInTheDocument()
    })

    it('shows syncing for idle folders with needSize > 0', () => {
      render(
        <ActivityCard
          state="syncing"
          folders={[
            createFolder({
              state: 'idle',
              globalSize: 100 * 1024 * 1024,
              needSize: 50 * 1024 * 1024,
            }),
          ]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('50 MB remaining')).toBeInTheDocument()
    })

    it('shows queue count when multiple folders are syncing', () => {
      render(
        <ActivityCard
          state="syncing"
          folders={[
            createFolder({ id: 'folder-1', label: 'Folder 1', state: 'syncing', needSize: 100 }),
            createFolder({ id: 'folder-2', label: 'Folder 2', state: 'syncing', needSize: 200 }),
          ]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('+1 folder in queue')).toBeInTheDocument()
    })
  })

  describe('combined scanning and syncing', () => {
    it('shows both scanning and syncing progress when both states exist', () => {
      render(
        <ActivityCard
          state="syncing"
          folders={[
            createFolder({
              id: 'folder-1',
              label: 'Scanning Folder',
              state: 'scanning',
              globalSize: 1000,
              needSize: 800,
            }),
            createFolder({
              id: 'folder-2',
              label: 'Syncing Folder',
              state: 'syncing',
              globalSize: 2000,
              needSize: 500,
            }),
          ]}
          lastSyncedAt={null}
          hasPairedDevices={true}
        />,
      )

      expect(screen.getByText('Scanning Folder')).toBeInTheDocument()
      expect(screen.getByText('Syncing Folder')).toBeInTheDocument()
    })
  })
})

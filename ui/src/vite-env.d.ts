/// <reference types="vite/client" />

import type { EventChannel, InvokeChannel, UpdateProgressEvent } from '../electron/channels'
import type {
  ProgressEvent,
  SyncEnableProgressEvent,
  SyncPairingProgressEvent,
} from './types/daemon.gen'

type EventPayloadMap = {
  'apply:progress': ProgressEvent
  'pairing:progress': SyncPairingProgressEvent
  'sync_enable:progress': SyncEnableProgressEvent
  'update:progress': UpdateProgressEvent
}

declare global {
  interface Window {
    electron: {
      homeDir: string
      invoke<T>(command: InvokeChannel, data?: unknown): Promise<T>
      on<C extends EventChannel>(
        channel: C,
        callback: (data: EventPayloadMap[C]) => void,
      ): () => void
      off(channel: EventChannel): void
    }
  }
}

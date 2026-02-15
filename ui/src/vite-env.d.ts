/// <reference types="vite/client" />

import type { EventChannel, InvokeChannel, UpdateProgressEvent } from '../electron/channels'
import type { ProgressEvent, SyncPairingProgressEvent } from './types/daemon.gen'

type EventPayloadMap = {
  'apply:progress': ProgressEvent
  'pairing:progress': SyncPairingProgressEvent
  'update:progress': UpdateProgressEvent
}

declare global {
  interface Window {
    electron: {
      invoke<T>(command: InvokeChannel, data?: unknown): Promise<T>
      on<C extends EventChannel>(
        channel: C,
        callback: (data: EventPayloadMap[C]) => void,
      ): () => void
      off(channel: EventChannel): void
    }
  }
}

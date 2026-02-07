/// <reference types="vite/client" />

import type { EventChannel, InvokeChannel } from '../electron/channels'

declare global {
  interface Window {
    electron: {
      invoke<T>(command: InvokeChannel, data?: unknown): Promise<T>
      on(channel: EventChannel, callback: (...args: unknown[]) => void): void
      off(channel: EventChannel): void
    }
  }
}

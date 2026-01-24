interface Window {
  electron: {
    invoke: (channel: string, ...args: unknown[]) => Promise<unknown>
  }
}

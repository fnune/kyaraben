import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, vi } from 'vitest'

afterEach(() => {
  cleanup()
})

Object.defineProperty(window, 'electron', {
  value: {
    invoke: vi.fn((channel: string) => {
      if (channel === 'get_home_dir') return Promise.resolve('/home/testuser')
      return Promise.resolve(undefined)
    }),
    on: vi.fn(() => vi.fn()),
    off: vi.fn(),
  },
  writable: true,
})

import { describe, expect, it, vi } from 'vitest'
import { buildXdgDataDirs } from './xdg'

vi.mock('node:os', () => ({
  homedir: () => '/home/testuser',
}))

describe('buildXdgDataDirs', () => {
  it('prepends flatpak dirs when XDG_DATA_DIRS is undefined', () => {
    const result = buildXdgDataDirs(undefined)
    expect(result).toBe(
      '/home/testuser/.local/share/flatpak/exports/share:/var/lib/flatpak/exports/share:/usr/local/share:/usr/share',
    )
  })

  it('prepends flatpak dirs to existing XDG_DATA_DIRS', () => {
    const result = buildXdgDataDirs('/custom/share:/usr/share')
    expect(result).toBe(
      '/home/testuser/.local/share/flatpak/exports/share:/var/lib/flatpak/exports/share:/custom/share:/usr/share',
    )
  })

  it('deduplicates directories', () => {
    const result = buildXdgDataDirs(
      '/var/lib/flatpak/exports/share:/usr/share:/home/testuser/.local/share/flatpak/exports/share',
    )
    expect(result).toBe(
      '/home/testuser/.local/share/flatpak/exports/share:/var/lib/flatpak/exports/share:/usr/share',
    )
  })

  it('handles empty string', () => {
    const result = buildXdgDataDirs('')
    expect(result).toBe(
      '/home/testuser/.local/share/flatpak/exports/share:/var/lib/flatpak/exports/share:/usr/local/share:/usr/share',
    )
  })
})

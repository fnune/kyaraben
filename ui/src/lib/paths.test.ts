import { describe, expect, it } from 'vitest'
import { collapseTilde, expandTilde } from './paths'

describe('expandTilde', () => {
  const homeDir = '/home/user'

  it('expands ~/path to absolute path', () => {
    expect(expandTilde('~/Emulation', homeDir)).toBe('/home/user/Emulation')
  })

  it('expands ~/nested/path correctly', () => {
    expect(expandTilde('~/Documents/ROMs', homeDir)).toBe('/home/user/Documents/ROMs')
  })

  it('expands ~ alone to home directory', () => {
    expect(expandTilde('~', homeDir)).toBe('/home/user')
  })

  it('returns path unchanged if no tilde prefix', () => {
    expect(expandTilde('/run/media/user/sdcard', homeDir)).toBe('/run/media/user/sdcard')
  })

  it('returns path unchanged if homeDir is empty', () => {
    expect(expandTilde('~/Emulation', '')).toBe('~/Emulation')
  })

  it('does not expand tilde in middle of path', () => {
    expect(expandTilde('/some/~/path', homeDir)).toBe('/some/~/path')
  })
})

describe('collapseTilde', () => {
  const homeDir = '/home/user'

  it('collapses home directory path to ~/', () => {
    expect(collapseTilde('/home/user/Emulation', homeDir)).toBe('~/Emulation')
  })

  it('collapses nested home paths correctly', () => {
    expect(collapseTilde('/home/user/Documents/ROMs', homeDir)).toBe('~/Documents/ROMs')
  })

  it('collapses exact home directory to ~', () => {
    expect(collapseTilde('/home/user', homeDir)).toBe('~')
  })

  it('returns path unchanged if not under home directory', () => {
    expect(collapseTilde('/run/media/user/sdcard', homeDir)).toBe('/run/media/user/sdcard')
  })

  it('returns path unchanged if homeDir is empty', () => {
    expect(collapseTilde('/home/user/Emulation', '')).toBe('/home/user/Emulation')
  })

  it('does not collapse partial matches', () => {
    expect(collapseTilde('/home/username/Emulation', homeDir)).toBe('/home/username/Emulation')
  })
})

describe('expandTilde and collapseTilde roundtrip', () => {
  const homeDir = '/home/user'

  it('roundtrips ~/path correctly', () => {
    const original = '~/Emulation'
    const expanded = expandTilde(original, homeDir)
    const collapsed = collapseTilde(expanded, homeDir)
    expect(collapsed).toBe(original)
  })

  it('roundtrips absolute home path correctly', () => {
    const original = '/home/user/Emulation'
    const collapsed = collapseTilde(original, homeDir)
    const expanded = expandTilde(collapsed, homeDir)
    expect(expanded).toBe(original)
  })
})

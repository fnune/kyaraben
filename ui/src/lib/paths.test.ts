import { describe, expect, it } from 'vitest'
import { collapsePathsInText, collapseTilde, expandTilde } from './paths'

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

describe('collapsePathsInText', () => {
  const homeDir = '/home/user'

  it('collapses home directory in message text', () => {
    expect(
      collapsePathsInText('Using /home/user/Emulation (existing data preserved)', homeDir),
    ).toBe('Using ~/Emulation (existing data preserved)')
  })

  it('collapses multiple occurrences', () => {
    expect(collapsePathsInText('Copying /home/user/src to /home/user/dest', homeDir)).toBe(
      'Copying ~/src to ~/dest',
    )
  })

  it('returns text unchanged if no home directory present', () => {
    expect(collapsePathsInText('Installing package...', homeDir)).toBe('Installing package...')
  })

  it('returns text unchanged if homeDir is empty', () => {
    expect(collapsePathsInText('Using /home/user/Emulation', '')).toBe('Using /home/user/Emulation')
  })

  it('returns empty string unchanged', () => {
    expect(collapsePathsInText('', homeDir)).toBe('')
  })
})

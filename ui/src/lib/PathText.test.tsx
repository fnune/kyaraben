import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { PathText } from './PathText'

vi.stubGlobal('window', {
  electron: {
    homeDir: '/home/testuser',
  },
})

describe('PathText', () => {
  it('collapses home directory to tilde', () => {
    render(<PathText>/home/testuser/Emulation</PathText>)
    expect(screen.getByText('~/Emulation')).toBeInTheDocument()
  })

  it('shows full path in title attribute', () => {
    render(<PathText>/home/testuser/Emulation</PathText>)
    const element = screen.getByText('~/Emulation')
    expect(element).toHaveAttribute('title', '/home/testuser/Emulation')
  })

  it('leaves non-home paths unchanged', () => {
    render(<PathText>/run/media/user/sdcard</PathText>)
    expect(screen.getByText('/run/media/user/sdcard')).toBeInTheDocument()
  })

  it('expands tilde for title when path uses tilde', () => {
    render(<PathText>~/Emulation</PathText>)
    const element = screen.getByText('~/Emulation')
    expect(element).toHaveAttribute('title', '/home/testuser/Emulation')
  })

  it('applies monospace font class', () => {
    render(<PathText>/some/path</PathText>)
    const element = screen.getByText('/some/path')
    expect(element).toHaveClass('font-mono')
  })
})

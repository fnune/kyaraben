import { renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useOnWindowFocus } from './useOnWindowFocus'

describe('useOnWindowFocus', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('does not call callback on first focus event', () => {
    const callback = vi.fn()
    renderHook(() => useOnWindowFocus(callback))

    window.dispatchEvent(new Event('focus'))
    vi.runAllTimers()

    expect(callback).not.toHaveBeenCalled()
  })

  it('calls callback on second focus event after debounce', () => {
    const callback = vi.fn()
    renderHook(() => useOnWindowFocus(callback))

    window.dispatchEvent(new Event('focus'))
    window.dispatchEvent(new Event('focus'))
    vi.runAllTimers()

    expect(callback).toHaveBeenCalledTimes(1)
  })

  it('debounces rapid focus events', () => {
    const callback = vi.fn()
    renderHook(() => useOnWindowFocus(callback, 500))

    window.dispatchEvent(new Event('focus'))
    window.dispatchEvent(new Event('focus'))
    vi.advanceTimersByTime(200)
    window.dispatchEvent(new Event('focus'))
    vi.advanceTimersByTime(200)
    window.dispatchEvent(new Event('focus'))
    vi.runAllTimers()

    expect(callback).toHaveBeenCalledTimes(1)
  })

  it('uses custom debounce time', () => {
    const callback = vi.fn()
    renderHook(() => useOnWindowFocus(callback, 1000))

    window.dispatchEvent(new Event('focus'))
    window.dispatchEvent(new Event('focus'))
    vi.advanceTimersByTime(500)

    expect(callback).not.toHaveBeenCalled()

    vi.advanceTimersByTime(500)

    expect(callback).toHaveBeenCalledTimes(1)
  })

  it('cleans up event listener on unmount', () => {
    const callback = vi.fn()
    const { unmount } = renderHook(() => useOnWindowFocus(callback))

    window.dispatchEvent(new Event('focus'))
    unmount()
    window.dispatchEvent(new Event('focus'))
    vi.runAllTimers()

    expect(callback).not.toHaveBeenCalled()
  })

  it('uses latest callback reference', () => {
    const callback1 = vi.fn()
    const callback2 = vi.fn()
    const { rerender } = renderHook(({ cb }) => useOnWindowFocus(cb), {
      initialProps: { cb: callback1 },
    })

    window.dispatchEvent(new Event('focus'))
    rerender({ cb: callback2 })
    window.dispatchEvent(new Event('focus'))
    vi.runAllTimers()

    expect(callback1).not.toHaveBeenCalled()
    expect(callback2).toHaveBeenCalledTimes(1)
  })
})

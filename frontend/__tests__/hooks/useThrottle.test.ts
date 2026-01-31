import { renderHook } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

import { useThrottle } from '@/lib/hooks/useThrottle'

describe('useThrottle', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('throttles function calls', () => {
    const mockFn = vi.fn()
    const { result } = renderHook(() => useThrottle(mockFn, 300))

    // Call the throttled function multiple times rapidly
    result.current('arg1')
    result.current('arg2')
    result.current('arg3')

    // Should only call once immediately
    expect(mockFn).toHaveBeenCalledTimes(1)
    expect(mockFn).toHaveBeenCalledWith('arg1')

    // Advance time
    vi.advanceTimersByTime(300)
    
    // Call again after throttle period
    result.current('arg4')
    expect(mockFn).toHaveBeenCalledTimes(2)
    expect(mockFn).toHaveBeenCalledWith('arg4')
  })

  it('returns same function reference on re-render', () => {
    const mockFn = vi.fn()
    const { result, rerender } = renderHook(() => useThrottle(mockFn, 300))
    const firstResult = result.current

    rerender()

    // Function reference should be stable
    expect(result.current).toBe(firstResult)
  })
})

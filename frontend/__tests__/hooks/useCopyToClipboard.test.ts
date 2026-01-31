import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'

import { useCopyToClipboard } from '@/lib/hooks/useCopyToClipboard'

describe('useCopyToClipboard', () => {
  beforeEach(() => {
    // Mock clipboard API
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
    })
  })

  it('copies text to clipboard', async () => {
    const { result } = renderHook(() => useCopyToClipboard())
    
    await act(async () => {
      await result.current.copy('test text')
    })

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('test text')
    expect(result.current.copied).toBe(true)
  })

  it('resets copied state after timeout', async () => {
    vi.useFakeTimers()
    const { result } = renderHook(() => useCopyToClipboard())

    await act(async () => {
      await result.current.copy('test text')
    })

    expect(result.current.copied).toBe(true)

    await act(async () => {
      vi.advanceTimersByTime(2000)
    })

    expect(result.current.copied).toBe(false)
    vi.useRealTimers()
  })

  it('handles copy errors', async () => {
    const error = new Error('Copy failed')
    vi.mocked(navigator.clipboard.writeText).mockRejectedValueOnce(error)
    
    const { result } = renderHook(() => useCopyToClipboard())
    
    const success = await act(async () => {
      return await result.current.copy('test text')
    })

    // Should return false on error
    expect(success).toBe(false)
    expect(result.current.copied).toBe(false)
  })
})

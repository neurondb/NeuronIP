import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, beforeEach, afterEach } from 'vitest'

import { useLocalStorage } from '@/lib/hooks/useLocalStorage'

describe('useLocalStorage', () => {
  const key = 'test-key'
  
  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    localStorage.clear()
  })

  it('returns initial value when localStorage is empty', () => {
    const { result } = renderHook(() => useLocalStorage(key, 'default'))
    expect(result.current[0]).toBe('default')
  })

  it('reads from localStorage', () => {
    localStorage.setItem(key, JSON.stringify('stored-value'))
    const { result } = renderHook(() => useLocalStorage(key, 'default'))
    expect(result.current[0]).toBe('stored-value')
  })

  it('updates localStorage on setValue', () => {
    const { result } = renderHook(() => useLocalStorage(key, 'default'))
    
    act(() => {
      result.current[1]('new-value')
    })

    expect(result.current[0]).toBe('new-value')
    expect(JSON.parse(localStorage.getItem(key) || '')).toBe('new-value')
  })

  it('handles complex objects', () => {
    const obj = { name: 'test', count: 42 }
    const { result } = renderHook(() => useLocalStorage(key, obj))
    
    act(() => {
      result.current[1]({ name: 'updated', count: 100 })
    })

    expect(result.current[0]).toEqual({ name: 'updated', count: 100 })
  })
})

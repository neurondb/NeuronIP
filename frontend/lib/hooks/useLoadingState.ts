import { useState, useCallback } from 'react'

export interface LoadingState {
  isLoading: boolean
  error: Error | null
  startLoading: () => void
  stopLoading: () => void
  setError: (error: Error | null) => void
  clearError: () => void
  reset: () => void
}

/**
 * Hook for managing loading state with error handling
 */
export function useLoadingState(initialState = false): LoadingState {
  const [isLoading, setIsLoading] = useState(initialState)
  const [error, setError] = useState<Error | null>(null)

  const startLoading = useCallback(() => {
    setIsLoading(true)
    setError(null)
  }, [])

  const stopLoading = useCallback(() => {
    setIsLoading(false)
  }, [])

  const setErrorState = useCallback((err: Error | null) => {
    setError(err)
    setIsLoading(false)
  }, [])

  const clearError = useCallback(() => {
    setError(null)
  }, [])

  const reset = useCallback(() => {
    setIsLoading(false)
    setError(null)
  }, [])

  return {
    isLoading,
    error,
    startLoading,
    stopLoading,
    setError: setErrorState,
    clearError,
    reset,
  }
}

/**
 * Hook for managing async operation loading state
 */
export function useAsyncLoading<T>() {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [data, setData] = useState<T | null>(null)

  const execute = useCallback(async (asyncFn: () => Promise<T>) => {
    setIsLoading(true)
    setError(null)
    try {
      const result = await asyncFn()
      setData(result)
      return result
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err))
      setError(error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }, [])

  const reset = useCallback(() => {
    setIsLoading(false)
    setError(null)
    setData(null)
  }, [])

  return {
    isLoading,
    error,
    data,
    execute,
    reset,
  }
}

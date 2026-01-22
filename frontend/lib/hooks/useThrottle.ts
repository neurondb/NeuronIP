import { useRef, useCallback } from 'react'

export function useThrottle<T extends (...args: any[]) => any>(
  func: T,
  delay: number = 300
): T {
  const lastRan = useRef<number>(0)

  return useCallback(
    ((...args: Parameters<T>) => {
      const now = Date.now()
      if (now - lastRan.current >= delay) {
        lastRan.current = now
        return func(...args)
      }
    }) as T,
    [func, delay]
  )
}

import { create } from 'zustand'
import { persist, devtools } from 'zustand/middleware'

interface StoreConfig<T> {
  name: string
  persist?: boolean
  devtools?: boolean
}

export function createEnhancedStore<T extends object>(
  initialState: T,
  config: StoreConfig<T>
) {
  const { name, persist: enablePersist = false, devtools: enableDevtools = true } = config

  let store = (set: any, get: any) => ({
    ...initialState,
  })

  if (enablePersist) {
    store = persist(store as any, {
      name,
      partialize: (state: any) => {
        // Only persist specific keys if needed
        return state
      },
    }) as any
  }

  if (enableDevtools && typeof window !== 'undefined') {
    store = devtools(store as any, { name }) as any
  }

  return create<T & { reset: () => void }>(store as any)
}

// Optimistic update helper
export function createOptimisticStore<T extends object>(
  initialState: T,
  config: StoreConfig<T>
) {
  const store = createEnhancedStore(initialState, config)

  return {
    ...store,
    optimisticUpdate: (updates: Partial<T>, rollback: () => void) => {
      const currentState = store.getState()
      store.setState(updates as any)

      return {
        rollback: () => {
          store.setState(currentState as any)
          rollback()
        },
      }
    },
  }
}

import { ComponentType, lazy, LazyExoticComponent } from 'react'

export function createLazyComponent<T extends ComponentType<any>>(
  importFn: () => Promise<{ default: T }>
): LazyExoticComponent<T> {
  return lazy(importFn)
}

export function lazyRoute<T extends ComponentType<any>>(
  importFn: () => Promise<{ default: T }>
): LazyExoticComponent<T> {
  return lazy(() =>
    importFn().then((module) => ({
      default: module.default,
    }))
  )
}

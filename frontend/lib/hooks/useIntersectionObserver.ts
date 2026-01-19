import { RefObject, useEffect, useState } from 'react'

interface UseIntersectionObserverOptions {
  elementRef: RefObject<Element>
  threshold?: number | number[]
  rootMargin?: string
  root?: Element | null
}

export function useIntersectionObserver({
  elementRef,
  threshold = 0,
  rootMargin = '0px',
  root = null,
}: UseIntersectionObserverOptions) {
  const [isIntersecting, setIsIntersecting] = useState(false)
  const [hasIntersected, setHasIntersected] = useState(false)

  useEffect(() => {
    const element = elementRef.current
    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsIntersecting(entry.isIntersecting)
        if (entry.isIntersecting) {
          setHasIntersected(true)
        }
      },
      { threshold, rootMargin, root }
    )

    observer.observe(element)

    return () => {
      observer.disconnect()
    }
  }, [elementRef, threshold, rootMargin, root])

  return { isIntersecting, hasIntersected }
}

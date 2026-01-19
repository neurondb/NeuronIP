import { useEffect, useRef, useState } from 'react'

export function useInView(options?: IntersectionObserverInit) {
  const [isInView, setIsInView] = useState(false)
  const ref = useRef<HTMLElement>(null)

  useEffect(() => {
    const element = ref.current
    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsInView(entry.isIntersecting)
      },
      { threshold: 0.1, ...options }
    )

    observer.observe(element)

    return () => {
      observer.disconnect()
    }
  }, [options])

  return { ref, isInView }
}

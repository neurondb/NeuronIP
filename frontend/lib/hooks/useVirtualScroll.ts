import { useRef, useEffect, useState, RefObject } from 'react'

interface UseVirtualScrollOptions {
  itemHeight: number | ((index: number) => number)
  containerHeight: number
  itemCount: number
  overscan?: number
}

interface VirtualScrollResult {
  scrollElementRef: RefObject<HTMLDivElement>
  startIndex: number
  endIndex: number
  totalHeight: number
  offsetY: number
}

export function useVirtualScroll({
  itemHeight,
  containerHeight,
  itemCount,
  overscan = 5,
}: UseVirtualScrollOptions): VirtualScrollResult {
  const scrollElementRef = useRef<HTMLDivElement>(null)
  const [scrollTop, setScrollTop] = useState(0)

  const getItemHeight = (index: number): number => {
    return typeof itemHeight === 'function' ? itemHeight(index) : itemHeight
  }

  const totalHeight = Array.from({ length: itemCount }, (_, i) => getItemHeight(i)).reduce(
    (sum, height) => sum + height,
    0
  )

  let startIndex = 0
  let offsetY = 0
  let accumulatedHeight = 0

  for (let i = 0; i < itemCount; i++) {
    const height = getItemHeight(i)
    if (accumulatedHeight + height > scrollTop) {
      startIndex = Math.max(0, i - overscan)
      break
    }
    accumulatedHeight += height
    offsetY = accumulatedHeight
  }

  let endIndex = startIndex
  accumulatedHeight = offsetY

  for (let i = startIndex; i < itemCount; i++) {
    const height = getItemHeight(i)
    if (accumulatedHeight > scrollTop + containerHeight) {
      endIndex = i + overscan
      break
    }
    accumulatedHeight += height
    endIndex = i + 1
  }

  endIndex = Math.min(itemCount, endIndex + overscan)

  useEffect(() => {
    const element = scrollElementRef.current
    if (!element) return

    const handleScroll = () => {
      setScrollTop(element.scrollTop)
    }

    element.addEventListener('scroll', handleScroll, { passive: true })
    return () => element.removeEventListener('scroll', handleScroll)
  }, [])

  return {
    scrollElementRef,
    startIndex,
    endIndex,
    totalHeight,
    offsetY,
  }
}

'use client'

import { ReactNode, useEffect, useRef, useCallback } from 'react'
import { useIntersectionObserver } from '@/lib/hooks/useIntersectionObserver'

interface InfiniteScrollProps {
  children: ReactNode
  onLoadMore: () => void | Promise<void>
  hasMore: boolean
  isLoading?: boolean
  loader?: ReactNode
  endMessage?: ReactNode
  threshold?: number
  rootMargin?: string
  className?: string
}

export default function InfiniteScroll({
  children,
  onLoadMore,
  hasMore,
  isLoading = false,
  loader,
  endMessage,
  threshold = 0.1,
  rootMargin = '100px',
  className,
}: InfiniteScrollProps) {
  const loadMoreRef = useRef<HTMLDivElement>(null)

  const { isIntersecting } = useIntersectionObserver({
    elementRef: loadMoreRef,
    threshold,
    rootMargin,
  })

  const handleLoadMore = useCallback(async () => {
    if (hasMore && !isLoading && isIntersecting) {
      await onLoadMore()
    }
  }, [hasMore, isLoading, isIntersecting, onLoadMore])

  useEffect(() => {
    handleLoadMore()
  }, [handleLoadMore])

  return (
    <div className={className}>
      {children}
      <div ref={loadMoreRef} className="h-1 w-full">
        {isLoading && loader && <div className="py-4">{loader}</div>}
        {!hasMore && endMessage && <div className="py-4 text-center text-muted-foreground">{endMessage}</div>}
      </div>
    </div>
  )
}

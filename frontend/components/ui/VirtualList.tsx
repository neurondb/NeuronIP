'use client'

import React, { ReactNode, useMemo } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { cn } from '@/lib/utils/cn'

interface VirtualListProps<T> {
  items: T[]
  itemHeight: number | ((index: number) => number)
  containerHeight?: number
  renderItem: (item: T, index: number) => ReactNode
  className?: string
  overscan?: number
}

export default function VirtualList<T>({
  items,
  itemHeight,
  containerHeight = 400,
  renderItem,
  className,
  overscan = 5,
}: VirtualListProps<T>) {
  const parentRef = React.useRef<HTMLDivElement>(null)

  const rowVirtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: (index) => (typeof itemHeight === 'function' ? itemHeight(index) : itemHeight),
    overscan,
  })

  const items_ = rowVirtualizer.getVirtualItems()

  return (
    <div
      ref={parentRef}
      className={cn('overflow-auto', className)}
      style={{ height: `${containerHeight}px` }}
    >
      <div
        style={{
          height: `${rowVirtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {items_.map((virtualItem) => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualItem.size}px`,
              transform: `translateY(${virtualItem.start}px)`,
            }}
          >
            {renderItem(items[virtualItem.index], virtualItem.index)}
          </div>
        ))}
      </div>
    </div>
  )
}

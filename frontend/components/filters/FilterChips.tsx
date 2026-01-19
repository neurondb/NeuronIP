'use client'

import { motion } from 'framer-motion'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'

interface Filter {
  id: string
  label: string
  value: string
  type: string
}

interface FilterChipsProps {
  filters: Filter[]
  onRemove: (id: string) => void
  onClearAll?: () => void
  className?: string
}

export default function FilterChips({
  filters,
  onRemove,
  onClearAll,
  className,
}: FilterChipsProps) {
  if (filters.length === 0) return null

  return (
    <div className={cn('flex flex-wrap gap-2', className)}>
      {filters.map((filter, index) => (
        <motion.div
          key={filter.id}
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: index * 0.05 }}
          className="flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1 text-sm"
        >
          <span className="text-muted-foreground text-xs">{filter.type}:</span>
          <span>{filter.label}</span>
          <button
            onClick={() => onRemove(filter.id)}
            className="rounded-full p-0.5 hover:bg-destructive/10 transition-colors"
          >
            <XMarkIcon className="h-3 w-3" />
          </button>
        </motion.div>
      ))}
      {onClearAll && filters.length > 1 && (
        <button
          onClick={onClearAll}
          className="text-xs text-muted-foreground hover:text-foreground transition-colors px-2"
        >
          Clear all
        </button>
      )}
    </div>
  )
}

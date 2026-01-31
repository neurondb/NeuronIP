'use client'

import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Bars3Icon as GripVerticalIcon } from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'

import { cn } from '@/lib/utils/cn'

interface BlockDragHandleProps {
  id: string
  children: React.ReactNode
  className?: string
}

export default function BlockDragHandle({ id, children, className }: BlockDragHandleProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn('group relative flex items-start gap-2', className)}
    >
      <motion.button
        {...attributes}
        {...listeners}
        className={cn(
          'opacity-0 group-hover:opacity-100 transition-opacity cursor-grab active:cursor-grabbing',
          'p-1 rounded hover:bg-accent',
          isDragging && 'opacity-100'
        )}
        whileHover={{ scale: 1.1 }}
        whileTap={{ scale: 0.9 }}
      >
        <GripVerticalIcon className="h-4 w-4 text-muted-foreground" />
      </motion.button>
      <div className="flex-1">{children}</div>
    </div>
  )
}

'use client'

import { DndContext, DragEndEvent, DragOverlay, DragStartEvent, closestCorners, useDroppable } from '@dnd-kit/core'
import { SortableContext, useSortable, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { motion } from 'framer-motion'
import { useState, useMemo } from 'react'

import { cn } from '@/lib/utils/cn'

interface KanbanViewProps {
  data: any[]
  columns: Array<{
    id: string
    name: string
    type: 'text' | 'number' | 'date' | 'select' | 'multiSelect' | 'person' | 'file' | 'checkbox'
    options?: string[]
  }>
  onUpdate?: (rowId: string, columnId: string, value: any) => void
  onDelete?: (rowId: string) => void
  onCreate?: (data: Record<string, any>) => void
}

interface KanbanColumn {
  id: string
  title: string
  items: any[]
}

function KanbanCard({ item, columns }: { item: any; columns: KanbanViewProps['columns'] }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: item.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  const titleColumn = columns.find((col) => col.type === 'text') || columns[0]
  const title = item[titleColumn.id] || 'Untitled'

  return (
    <motion.div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={cn(
        'p-3 bg-card border border-border rounded-lg cursor-grab active:cursor-grabbing',
        'hover:shadow-md transition-shadow',
        isDragging && 'shadow-lg'
      )}
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
    >
      <div className="font-medium text-sm mb-1">{title}</div>
      <div className="text-xs text-muted-foreground">
        {columns.slice(1, 3).map((col) => {
          const value = item[col.id]
          if (!value) return null
          return (
            <div key={col.id} className="truncate">
              {col.name}: {String(value)}
            </div>
          )
        })}
      </div>
    </motion.div>
  )
}

function KanbanColumnComponent({
  column,
  items,
  columns,
}: {
  column: KanbanColumn
  items: any[]
  columns: KanbanViewProps['columns']
}) {
  const { setNodeRef } = useDroppable({
    id: column.id,
  })

  return (
    <div ref={setNodeRef} className="flex-1 min-w-[280px]">
      <div className="mb-3">
        <h3 className="font-semibold text-sm text-foreground">{column.title}</h3>
        <span className="text-xs text-muted-foreground">{items.length} items</span>
      </div>
      <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
        <div className="space-y-2">
          {items.map((item) => (
            <KanbanCard key={item.id} item={item} columns={columns} />
          ))}
        </div>
      </SortableContext>
    </div>
  )
}

export default function KanbanView({
  data,
  columns,
  onUpdate,
  onDelete,
  onCreate,
}: KanbanViewProps) {
  const [activeId, setActiveId] = useState<string | null>(null)

  // Find a select column to group by, or use a default status column
  const groupColumn = columns.find(
    (col) => col.type === 'select' || col.type === 'multiSelect'
  ) || {
    id: 'status',
    name: 'Status',
    options: ['Todo', 'In Progress', 'Done'],
  }

  const kanbanColumns: KanbanColumn[] = useMemo(() => {
    const defaultStatuses = groupColumn.options || ['Todo', 'In Progress', 'Done']
    return defaultStatuses.map((status) => ({
      id: status,
      title: status,
      items: data.filter((item) => {
        const value = item[groupColumn.id]
        if (Array.isArray(value)) {
          return value.includes(status)
        }
        return value === status
      }),
    }))
  }, [data, groupColumn])

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string)
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    setActiveId(null)

    if (!over) return

    const itemId = active.id as string
    const newStatus = over.id as string

    // Find the item
    const item = data.find((d) => d.id === itemId)
    if (!item) return

    // Update the item's status
    onUpdate?.(itemId, groupColumn.id, newStatus)
  }

  const activeItem = activeId ? data.find((item) => item.id === activeId) : null

  return (
    <div className="p-4 h-full overflow-x-auto">
      <DndContext
        collisionDetection={closestCorners}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        <div className="flex gap-4 h-full">
          {kanbanColumns.map((column) => (
            <KanbanColumnComponent
              key={column.id}
              column={column}
              items={column.items}
              columns={columns}
            />
          ))}
        </div>
        <DragOverlay>
          {activeItem && (
            <div className="p-3 bg-card border border-border rounded-lg shadow-lg w-64">
              <div className="font-medium text-sm">
                {activeItem[columns.find((col) => col.type === 'text')?.id || columns[0].id] ||
                  'Untitled'}
              </div>
            </div>
          )}
        </DragOverlay>
      </DndContext>
    </div>
  )
}

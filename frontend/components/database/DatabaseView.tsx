'use client'

import {
  TableCellsIcon,
  Squares2X2Icon,
  CalendarIcon,
  PhotoIcon,
  ListBulletIcon,
} from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import { useState } from 'react'

import { cn } from '@/lib/utils/cn'

import { Button } from '../ui/Button'

import CalendarView from './CalendarView'
import GalleryView from './GalleryView'
import KanbanView from './KanbanView'
import ListView from './ListView'
import TableView from './TableView'

export type DatabaseViewType = 'table' | 'kanban' | 'calendar' | 'gallery' | 'list'

interface DatabaseViewProps {
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
  onRowClick?: (rowId: string, row: any) => void
  defaultView?: DatabaseViewType
  className?: string
}

const viewIcons = {
  table: TableCellsIcon,
  kanban: Squares2X2Icon,
  calendar: CalendarIcon,
  gallery: PhotoIcon,
  list: ListBulletIcon,
}

const viewLabels = {
  table: 'Table',
  kanban: 'Board',
  calendar: 'Calendar',
  gallery: 'Gallery',
  list: 'List',
}

export default function DatabaseView({
  data,
  columns,
  onUpdate,
  onDelete,
  onCreate,
  onRowClick,
  defaultView = 'table',
  className,
}: DatabaseViewProps) {
  const [currentView, setCurrentView] = useState<DatabaseViewType>(defaultView)

  const renderView = () => {
    const commonProps = {
      data,
      columns,
      onUpdate,
      onDelete,
      onCreate,
      onRowClick,
    }

    switch (currentView) {
      case 'table':
        return <TableView {...commonProps} />
      case 'kanban':
        return <KanbanView {...commonProps} />
      case 'calendar':
        return <CalendarView {...commonProps} />
      case 'gallery':
        return <GalleryView {...commonProps} />
      case 'list':
        return <ListView {...commonProps} />
      default:
        return <TableView {...commonProps} />
    }
  }

  return (
    <div className={cn('flex flex-col h-full', className)}>
      {/* View Selector */}
      <div className="flex items-center justify-between border-b border-border p-2 bg-muted/30">
        <div className="flex items-center gap-1">
          {(['table', 'kanban', 'calendar', 'gallery', 'list'] as DatabaseViewType[]).map(
            (viewType) => {
              const Icon = viewIcons[viewType]
              const isActive = currentView === viewType
              return (
                <Button
                  key={viewType}
                  variant={isActive ? 'primary' : 'ghost'}
                  size="sm"
                  onClick={() => setCurrentView(viewType)}
                  className={cn('gap-2', isActive && 'bg-primary text-primary-foreground')}
                >
                  <Icon className="h-4 w-4" />
                  <span className="hidden sm:inline">{viewLabels[viewType]}</span>
                </Button>
              )
            }
          )}
        </div>
      </div>

      {/* View Content */}
      <div className="flex-1 overflow-auto">{renderView()}</div>
    </div>
  )
}

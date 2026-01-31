'use client'

import { PencilIcon, TrashIcon } from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'

import { cn } from '@/lib/utils/cn'

import { Button } from '../ui/Button'
import { Card, CardContent } from '../ui/Card'

interface ListViewProps {
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

export default function ListView({
  data,
  columns,
  onUpdate,
  onDelete,
  onCreate,
}: ListViewProps) {
  const titleColumn = columns.find((col) => col.type === 'text') || columns[0]
  const visibleColumns = columns.slice(0, 4) // Show first 4 columns

  return (
    <div className="p-4">
      <div className="space-y-2">
        {data.map((item, index) => {
          const title = item[titleColumn.id] || 'Untitled'

          return (
            <motion.div
              key={item.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
            >
              <Card className="hover:shadow-md transition-shadow">
                <CardContent className="p-4">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold text-base mb-2">{title}</h3>
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        {visibleColumns.map((col) => {
                          if (col.id === titleColumn.id) return null
                          const value = item[col.id]
                          if (!value) return null

                          return (
                            <div key={col.id}>
                              <div className="text-xs text-muted-foreground mb-1">{col.name}</div>
                              <div className="text-sm font-medium">
                                {col.type === 'date' && value
                                  ? new Date(value).toLocaleDateString()
                                  : String(value)}
                              </div>
                            </div>
                          )
                        })}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          console.log('Edit', item)
                        }}
                      >
                        <PencilIcon className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="sm" onClick={() => onDelete?.(item.id)}>
                        <TrashIcon className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}

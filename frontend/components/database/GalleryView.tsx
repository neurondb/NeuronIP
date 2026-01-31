'use client'

import { PhotoIcon } from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'

import { cn } from '@/lib/utils/cn'

import { Card, CardContent } from '../ui/Card'

interface GalleryViewProps {
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

export default function GalleryView({
  data,
  columns,
  onUpdate,
  onDelete,
  onCreate,
}: GalleryViewProps) {
  const titleColumn = columns.find((col) => col.type === 'text') || columns[0]
  const imageColumn = columns.find((col) => col.type === 'file' || col.id.includes('image'))
  const descriptionColumn = columns.find(
    (col) => col.type === 'text' && col.id !== titleColumn.id
  )

  return (
    <div className="p-4">
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {data.map((item) => {
          const title = item[titleColumn.id] || 'Untitled'
          const image = imageColumn ? item[imageColumn.id] : null
          const description = descriptionColumn ? item[descriptionColumn.id] : null

          return (
            <motion.div
              key={item.id}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              className="cursor-pointer"
            >
              <Card className="h-full overflow-hidden hover:shadow-lg transition-shadow">
                <div className="aspect-video bg-muted flex items-center justify-center relative">
                  {image ? (
                    <img
                      src={image}
                      alt={title}
                      className="w-full h-full object-cover"
                      onError={(e) => {
                        e.currentTarget.style.display = 'none'
                      }}
                    />
                  ) : (
                    <PhotoIcon className="h-12 w-12 text-muted-foreground" />
                  )}
                </div>
                <CardContent className="p-3">
                  <h3 className="font-semibold text-sm mb-1 truncate">{title}</h3>
                  {description && (
                    <p className="text-xs text-muted-foreground line-clamp-2">{description}</p>
                  )}
                  <div className="mt-2 flex flex-wrap gap-1">
                    {columns.slice(0, 3).map((col) => {
                      const value = item[col.id]
                      if (!value || col.id === titleColumn.id || col.id === imageColumn?.id) {
                        return null
                      }
                      return (
                        <span
                          key={col.id}
                          className="text-xs px-2 py-0.5 bg-muted rounded text-muted-foreground"
                        >
                          {String(value)}
                        </span>
                      )
                    })}
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

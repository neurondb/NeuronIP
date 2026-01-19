'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { CalendarIcon } from '@heroicons/react/24/outline'
import { Card } from '@/components/ui/Card'
import { cn } from '@/lib/utils/cn'
import { format } from 'date-fns'

interface DateRange {
  start: Date | null
  end: Date | null
}

interface DateRangePickerProps {
  value: DateRange
  onChange: (range: DateRange) => void
  presets?: { label: string; range: DateRange }[]
  className?: string
}

const defaultPresets = [
  { label: 'Today', range: { start: new Date(), end: new Date() } },
  {
    label: 'Last 7 days',
    range: { start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), end: new Date() },
  },
  {
    label: 'Last 30 days',
    range: { start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), end: new Date() },
  },
  {
    label: 'This month',
    range: {
      start: new Date(new Date().getFullYear(), new Date().getMonth(), 1),
      end: new Date(),
    },
  },
]

export default function DateRangePicker({
  value,
  onChange,
  presets = defaultPresets,
  className,
}: DateRangePickerProps) {
  const [isOpen, setIsOpen] = useState(false)

  const formatDate = (date: Date | null) => {
    if (!date) return ''
    return format(date, 'MMM d, yyyy')
  }

  const formatRange = () => {
    if (!value.start && !value.end) return 'Select date range'
    if (!value.start) return `Until ${formatDate(value.end)}`
    if (!value.end) return `From ${formatDate(value.start)}`
    return `${formatDate(value.start)} - ${formatDate(value.end)}`
  }

  return (
    <div className={cn('relative', className)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'flex items-center gap-2 rounded-lg border border-border bg-background px-4 py-2',
          'text-sm focus:outline-none focus:ring-2 focus:ring-ring'
        )}
      >
        <CalendarIcon className="h-4 w-4 text-muted-foreground" />
        <span>{formatRange()}</span>
      </button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="absolute top-full z-50 mt-2"
          >
            <Card className="w-80 p-4">
              <div className="space-y-4">
                <div>
                  <label className="text-xs font-medium mb-2 block">Quick Presets</label>
                  <div className="grid grid-cols-2 gap-2">
                    {presets.map((preset) => (
                      <button
                        key={preset.label}
                        onClick={() => {
                          onChange(preset.range)
                          setIsOpen(false)
                        }}
                        className="rounded border border-border px-3 py-2 text-sm hover:bg-accent transition-colors"
                      >
                        {preset.label}
                      </button>
                    ))}
                  </div>
                </div>
                <div className="space-y-2">
                  <div>
                    <label className="text-xs font-medium mb-1 block">Start Date</label>
                    <input
                      type="date"
                      value={value.start ? format(value.start, 'yyyy-MM-dd') : ''}
                      onChange={(e) =>
                        onChange({ ...value, start: e.target.value ? new Date(e.target.value) : null })
                      }
                      className="w-full rounded border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-medium mb-1 block">End Date</label>
                    <input
                      type="date"
                      value={value.end ? format(value.end, 'yyyy-MM-dd') : ''}
                      onChange={(e) =>
                        onChange({ ...value, end: e.target.value ? new Date(e.target.value) : null })
                      }
                      className="w-full rounded border border-border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    />
                  </div>
                </div>
              </div>
            </Card>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

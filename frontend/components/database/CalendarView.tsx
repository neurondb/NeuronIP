'use client'

import moment from 'moment'
import { Calendar, momentLocalizer } from 'react-big-calendar'
import 'react-big-calendar/lib/css/react-big-calendar.css'

const localizer = momentLocalizer(moment)

interface CalendarViewProps {
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

export default function CalendarView({
  data,
  columns,
  onUpdate,
  onDelete,
  onCreate,
}: CalendarViewProps) {
  const dateColumn = columns.find((col) => col.type === 'date') || columns[0]
  const titleColumn = columns.find((col) => col.type === 'text') || columns[0]

  const events = data
    .filter((item) => item[dateColumn.id])
    .map((item) => ({
      id: item.id,
      title: item[titleColumn.id] || 'Untitled',
      start: new Date(item[dateColumn.id]),
      end: new Date(item[dateColumn.id]),
      resource: item,
    }))

  return (
    <div className="p-4 h-full">
      <Calendar
        localizer={localizer}
        events={events}
        startAccessor="start"
        endAccessor="end"
        style={{ height: '100%' }}
        onSelectEvent={(event) => {
          console.log('Selected event:', event.resource)
        }}
      />
    </div>
  )
}

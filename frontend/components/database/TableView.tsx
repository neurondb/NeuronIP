'use client'

import { PencilIcon, TrashIcon } from '@heroicons/react/24/outline'
import { ColumnDef } from '@tanstack/react-table'
import { useState } from 'react'

import { Button } from '../ui/Button'
import { DataTable } from '../ui/DataTable'


interface TableViewProps {
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
}

export default function TableView({
  data,
  columns,
  onUpdate,
  onDelete,
  onCreate,
  onRowClick,
}: TableViewProps) {
  const tableColumns: ColumnDef<any>[] = columns.map((col) => ({
    id: col.id,
    accessorKey: col.id,
    header: col.name,
    cell: ({ row, getValue }) => {
      const value = getValue()
      if (col.type === 'checkbox') {
        return (
          <input
            type="checkbox"
            checked={!!value}
            onChange={(e) => onUpdate?.(row.original.id, col.id, e.target.checked)}
            className="rounded"
          />
        )
      }
      if (col.type === 'date' && value) {
        return new Date(value as string).toLocaleDateString()
      }
      return <span>{value as string}</span>
    },
  }))

  tableColumns.push({
    id: 'actions',
    header: 'Actions',
    cell: ({ row }) => (
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => {
            // Edit action
            console.log('Edit', row.original)
          }}
        >
          <PencilIcon className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onDelete?.(row.original.id)}
        >
          <TrashIcon className="h-4 w-4" />
        </Button>
      </div>
    ),
  })

  return (
    <div className="p-4">
      <DataTable 
        columns={tableColumns} 
        data={data}
        onRowClick={onRowClick ? (row) => onRowClick(row.id, row) : undefined}
      />
    </div>
  )
}

'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { ArrowDownTrayIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'

interface ResultsTableProps {
  data: unknown[]
  columns?: string[]
  isLoading?: boolean
}

export default function ResultsTable({ data, columns, isLoading }: ResultsTableProps) {
  const [sortColumn, setSortColumn] = useState<string | null>(null)
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc')

  const handleSort = (column: string) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortColumn(column)
      setSortDirection('asc')
    }
  }

  const exportToCSV = () => {
    if (!data || data.length === 0) return

    const headers = columns || Object.keys(data[0] as Record<string, unknown>)
    const rows = data.map((row) =>
      headers.map((header) => {
        const value = (row as Record<string, unknown>)[header]
        return `"${String(value || '').replace(/"/g, '""')}"`
      })
    )

    const csv = [
      headers.map((h) => `"${h}"`).join(','),
      ...rows.map((row) => row.join(',')),
    ].join('\n')

    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'warehouse-results.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6">
          <Loading size="md" />
        </CardContent>
      </Card>
    )
  }

  if (!data || data.length === 0) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="text-center text-muted-foreground">No results to display</div>
        </CardContent>
      </Card>
    )
  }

  const headers = columns || Object.keys(data[0] as Record<string, unknown>)
  const sortedData = [...data].sort((a, b) => {
    if (!sortColumn) return 0
    const aVal = (a as Record<string, unknown>)[sortColumn]
    const bVal = (b as Record<string, unknown>)[sortColumn]
    if (aVal === bVal) return 0
    // Convert to strings for comparison
    const aStr = String(aVal ?? '')
    const bStr = String(bVal ?? '')
    const comparison = aStr < bStr ? -1 : 1
    return sortDirection === 'asc' ? comparison : -comparison
  })

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Query Results</CardTitle>
          <Button variant="outline" size="sm" onClick={exportToCSV}>
            <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
            Export CSV
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                {headers.map((header) => (
                  <TableHead
                    key={header}
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => handleSort(header)}
                  >
                    <div className="flex items-center gap-2">
                      {header}
                      {sortColumn === header && (
                        <span>{sortDirection === 'asc' ? '↑' : '↓'}</span>
                      )}
                    </div>
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {sortedData.slice(0, 100).map((row, index) => (
                <TableRow key={index}>
                  {headers.map((header) => (
                    <TableCell key={header}>
                      {String((row as Record<string, unknown>)[header] || '')}
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        {data.length > 100 && (
          <p className="text-xs text-muted-foreground mt-4">
            Showing first 100 of {data.length} results
          </p>
        )}
      </CardContent>
    </Card>
  )
}

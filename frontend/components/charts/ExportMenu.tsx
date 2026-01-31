'use client'

import { ArrowDownTrayIcon } from '@heroicons/react/24/outline'
import { saveAs } from 'file-saver'
import * as React from 'react'

import { Button } from '@/components/ui/Button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/DropdownMenu'


export interface ChartDataSeries {
  name: string
  values: (string | number)[]
}

export interface ChartExportData {
  labels: string[]
  series?: ChartDataSeries[]
}

interface ExportMenuProps {
  chartId?: string
  chartTitle?: string
  data?: ChartExportData
  onExport?: (format: 'png' | 'svg' | 'csv') => void
}

export function ExportMenu({ chartId, chartTitle = 'chart', data, onExport }: ExportMenuProps) {
  const handleExport = (format: 'png' | 'svg' | 'csv') => {
    if (onExport) {
      onExport(format)
      return
    }

    if (format === 'png' || format === 'svg') {
      const element = chartId ? document.getElementById(chartId) : document.querySelector('svg')
      if (!element) return

      if (format === 'svg') {
        const svgData = new XMLSerializer().serializeToString(element as Node)
        const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' })
        saveAs(svgBlob, `${chartTitle}.svg`)
      } else {
        // PNG export would require canvas conversion
        // This is a simplified version
        const canvas = document.createElement('canvas')
        const ctx = canvas.getContext('2d')
        const img = new Image()
        const svgData = new XMLSerializer().serializeToString(element as Node)
        const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' })
        const url = URL.createObjectURL(svgBlob)

        img.onload = () => {
          canvas.width = img.width
          canvas.height = img.height
          ctx?.drawImage(img, 0, 0)
          canvas.toBlob((blob) => {
            if (blob) {
              saveAs(blob, `${chartTitle}.png`)
            }
            URL.revokeObjectURL(url)
          })
        }
        img.src = url
      }
    } else if (format === 'csv') {
      let csv: string
      if (data?.labels?.length) {
        const headers = ['label', ...(data.series?.map((s) => s.name) ?? ['value'])]
        const rows: string[] = [headers.map((h) => `"${String(h).replace(/"/g, '""')}"`).join(',')]
        const len = data.labels.length
        for (let i = 0; i < len; i++) {
          const label = String(data.labels[i] ?? '').replace(/"/g, '""')
          const cells = data.series?.length
            ? data.series.map((s) => `"${String(s.values[i] ?? '').replace(/"/g, '""')}"`)
            : ['']
          rows.push(`"${label}",${cells.join(',')}`)
        }
        csv = rows.join('\n')
      } else {
        csv = 'label,value\n'
      }
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
      saveAs(blob, `${chartTitle}.csv`)
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm">
          <ArrowDownTrayIcon className="h-4 w-4 mr-2" />
          Export
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => handleExport('png')}>Export as PNG</DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleExport('svg')}>Export as SVG</DropdownMenuItem>
        <DropdownMenuItem onClick={() => handleExport('csv')}>Export as CSV</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

'use client'

import * as React from 'react'
import { saveAs } from 'file-saver'
import { Button } from '@/components/ui/Button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/DropdownMenu'
import { ArrowDownTrayIcon } from '@heroicons/react/24/outline'

interface ExportMenuProps {
  chartId?: string
  chartTitle?: string
  onExport?: (format: 'png' | 'svg' | 'csv') => void
}

export function ExportMenu({ chartId, chartTitle = 'chart', onExport }: ExportMenuProps) {
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
      // CSV export would need chart data
      // This is a placeholder
      const csv = 'data,value\n1,10\n2,20\n3,30'
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

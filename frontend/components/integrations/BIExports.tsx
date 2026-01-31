'use client'

import { useState } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'

export default function BIExports() {
  const [biType, setBiType] = useState('tableau')
  const [queryId, setQueryId] = useState('')
  const [format, setFormat] = useState('csv')

  const exportQuery = async () => {
    try {
      const response = await fetch(`/api/v1/integrations/bi/export?query_id=${queryId}&bi_type=${biType}&format=${format}`)
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `export.${format}`
      a.click()
    } catch (error) {
      console.error('Failed to export:', error)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>BI Export</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-2">BI Tool</label>
          <select
            value={biType}
            onChange={(e) => setBiType(e.target.value)}
            className="w-full p-2 border rounded"
          >
            <option value="tableau">Tableau</option>
            <option value="powerbi">Power BI</option>
            <option value="looker">Looker</option>
            <option value="qlik">Qlik</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium mb-2">Query ID</label>
          <input
            type="text"
            value={queryId}
            onChange={(e) => setQueryId(e.target.value)}
            className="w-full p-2 border rounded"
          />
        </div>
        <div>
          <label className="block text-sm font-medium mb-2">Format</label>
          <select
            value={format}
            onChange={(e) => setFormat(e.target.value)}
            className="w-full p-2 border rounded"
          >
            <option value="csv">CSV</option>
            <option value="json">JSON</option>
            <option value="xlsx">Excel</option>
            <option value="parquet">Parquet</option>
          </select>
        </div>
        <Button onClick={exportQuery}>Export</Button>
      </CardContent>
    </Card>
  )
}

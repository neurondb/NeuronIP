export function exportToCSV(data: unknown[], filename = 'export.csv'): void {
  if (!data || data.length === 0) return

  const headers = Object.keys(data[0] as Record<string, unknown>)
  const rows = data.map((row) =>
    headers.map((header) => {
      const value = (row as Record<string, unknown>)[header]
      return `"${String(value || '').replace(/"/g, '""')}"`
    })
  )

  const csv = [headers.map((h) => `"${h}"`).join(','), ...rows.map((row) => row.join(','))].join('\n')

  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export function exportToJSON(data: unknown[], filename = 'export.json'): void {
  if (!data || data.length === 0) return

  const json = JSON.stringify(data, null, 2)
  const blob = new Blob([json], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export function exportToExcel(data: unknown[], filename = 'export.xlsx'): void {
  // For Excel export, we'll create a CSV with proper encoding
  // In production, you might want to use a library like xlsx
  exportToCSV(data, filename.replace('.xlsx', '.csv'))
}

export function exportChartAsPNG(canvas: HTMLCanvasElement, filename = 'chart.png'): void {
  canvas.toBlob((blob) => {
    if (!blob) return
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
  })
}

export function exportChartAsSVG(svgElement: SVGElement, filename = 'chart.svg'): void {
  const serializer = new XMLSerializer()
  const svgString = serializer.serializeToString(svgElement)
  const blob = new Blob([svgString], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

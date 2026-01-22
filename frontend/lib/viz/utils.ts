export interface ChartDataPoint {
  name: string
  value: number
  [key: string]: unknown
}

export interface ChartConfig {
  colors?: string[]
  animation?: boolean
  responsive?: boolean
}

export const defaultColors = [
  '#0ea5e9',
  '#8b5cf6',
  '#ec4899',
  '#f59e0b',
  '#10b981',
  '#ef4444',
  '#6366f1',
  '#14b8a6',
]

export function generateColors(count: number, customColors?: string[]): string[] {
  if (customColors && customColors.length >= count) {
    return customColors.slice(0, count)
  }

  const colors: string[] = []
  for (let i = 0; i < count; i++) {
    colors.push(defaultColors[i % defaultColors.length])
  }
  return colors
}

export function formatNumber(value: number, decimals = 0): string {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(decimals)}M`
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(decimals)}K`
  }
  return value.toFixed(decimals)
}

export function calculatePercentage(value: number, total: number): number {
  if (total === 0) return 0
  return (value / total) * 100
}

export function getMaxValue(data: ChartDataPoint[]): number {
  return Math.max(...data.map((d) => d.value), 0)
}

export function getMinValue(data: ChartDataPoint[]): number {
  return Math.min(...data.map((d) => d.value), 0)
}

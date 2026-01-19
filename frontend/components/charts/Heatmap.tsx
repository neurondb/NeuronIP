'use client'

import { motion } from 'framer-motion'
import { cn } from '@/lib/utils/cn'

interface HeatmapData {
  x: string
  y: string
  value: number
}

interface HeatmapProps {
  data: HeatmapData[]
  xLabels: string[]
  yLabels: string[]
  colors?: string[]
  cellSize?: number
  className?: string
}

export default function Heatmap({
  data,
  xLabels,
  yLabels,
  colors = ['#fef3c7', '#fde68a', '#fcd34d', '#fbbf24', '#f59e0b'],
  cellSize = 40,
  className,
}: HeatmapProps) {
  const maxValue = Math.max(...data.map((d) => d.value), 1)

  const getColor = (value: number) => {
    const ratio = value / maxValue
    const index = Math.floor(ratio * (colors.length - 1))
    return colors[index] || colors[colors.length - 1]
  }

  return (
    <div className={cn('overflow-auto', className)}>
      <svg
        width={xLabels.length * cellSize + 100}
        height={yLabels.length * cellSize + 50}
        className="w-full h-auto"
      >
        {/* X-axis labels */}
        {xLabels.map((label, i) => (
          <text
            key={`x-${i}`}
            x={i * cellSize + cellSize / 2}
            y={20}
            textAnchor="middle"
            className="text-xs fill-muted-foreground"
          >
            {label}
          </text>
        ))}

        {/* Y-axis labels */}
        {yLabels.map((label, i) => (
          <text
            key={`y-${i}`}
            x={10}
            y={i * cellSize + cellSize / 2 + 30}
            textAnchor="end"
            className="text-xs fill-muted-foreground"
          >
            {label}
          </text>
        ))}

        {/* Heatmap cells */}
        {data.map((d, i) => {
          const xIndex = xLabels.indexOf(d.x)
          const yIndex = yLabels.indexOf(d.y)
          if (xIndex === -1 || yIndex === -1) return null

          return (
            <motion.rect
              key={i}
              x={xIndex * cellSize + 30}
              y={yIndex * cellSize + 30}
              width={cellSize - 2}
              height={cellSize - 2}
              fill={getColor(d.value)}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: i * 0.01 }}
              className="cursor-pointer hover:opacity-80"
            >
              <title>{`${d.x} - ${d.y}: ${d.value}`}</title>
            </motion.rect>
          )
        })}
      </svg>
    </div>
  )
}

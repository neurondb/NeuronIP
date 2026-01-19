'use client'

import { motion } from 'framer-motion'
import ChartContainer from '@/components/charts/ChartContainer'
import LineChart from '@/components/charts/LineChart'
import BarChart from '@/components/charts/BarChart'

interface AnomalyChartProps {
  data: unknown[]
}

export default function AnomalyChart({ data }: AnomalyChartProps) {
  // Transform data for chart display
  const chartData = Array.isArray(data) ? data : []

  if (chartData.length === 0) {
    return (
      <ChartContainer title="Anomaly Detection" description="No anomaly data available">
        <div className="h-[300px] flex items-center justify-center text-muted-foreground">
          No data to display
        </div>
      </ChartContainer>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <ChartContainer
        title="Anomaly Detection Over Time"
        description="Track compliance anomalies and violations"
      >
        <div className="h-[400px]">
          <LineChart
            data={chartData as any[]}
            dataKeys={['anomalies', 'violations']}
            xAxisKey="date"
            colors={['#ef4444', '#f59e0b']}
          />
        </div>
      </ChartContainer>
    </motion.div>
  )
}

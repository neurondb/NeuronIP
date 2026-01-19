'use client'

import { BarChart as RechartsBarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

interface BarChartProps {
  data: unknown[]
  dataKeys: string[]
  xAxisKey: string
  colors?: string[]
  showGrid?: boolean
  showLegend?: boolean
}

export default function BarChart({
  data,
  dataKeys,
  xAxisKey,
  colors = ['#0ea5e9', '#10b981', '#f59e0b', '#ef4444'],
  showGrid = true,
  showLegend = true,
}: BarChartProps) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <RechartsBarChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
        {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />}
        <XAxis dataKey={xAxisKey} stroke="hsl(var(--muted-foreground))" />
        <YAxis stroke="hsl(var(--muted-foreground))" />
        <Tooltip
          contentStyle={{
            backgroundColor: 'hsl(var(--card))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '8px',
          }}
        />
        {showLegend && <Legend />}
        {dataKeys.map((key, index) => (
          <Bar
            key={key}
            dataKey={key}
            fill={colors[index % colors.length]}
            radius={[4, 4, 0, 0]}
          />
        ))}
      </RechartsBarChart>
    </ResponsiveContainer>
  )
}

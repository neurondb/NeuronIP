'use client'

import { PieChart as RechartsPieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts'

interface PieChartProps {
  data: { name: string; value: number }[]
  colors?: string[]
  showLegend?: boolean
}

export default function PieChart({
  data,
  colors = ['#0ea5e9', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'],
  showLegend = true,
}: PieChartProps) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <RechartsPieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
          outerRadius={80}
          fill="#8884d8"
          dataKey="value"
        >
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
          ))}
        </Pie>
        <Tooltip
          contentStyle={{
            backgroundColor: 'hsl(var(--card))',
            border: '1px solid hsl(var(--border))',
            borderRadius: '8px',
          }}
        />
        {showLegend && <Legend />}
      </RechartsPieChart>
    </ResponsiveContainer>
  )
}

'use client'

import { ResponsiveContainer, PolarGrid, PolarAngleAxis, PolarRadiusAxis, Radar, RadarChart as RechartsRadarChart, Legend } from 'recharts'

interface RadarChartProps {
  data: unknown[]
  dataKeys: string[]
  angleKey: string
  colors?: string[]
  showLegend?: boolean
}

export default function RadarChart({
  data,
  dataKeys,
  angleKey,
  colors = ['#0ea5e9', '#10b981', '#f59e0b'],
  showLegend = true,
}: RadarChartProps) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <RechartsRadarChart data={data}>
        <PolarGrid stroke="hsl(var(--border))" />
        <PolarAngleAxis dataKey={angleKey} stroke="hsl(var(--muted-foreground))" />
        <PolarRadiusAxis angle={90} domain={[0, 100]} stroke="hsl(var(--muted-foreground))" />
        {showLegend && <Legend />}
        {dataKeys.map((key, index) => (
          <Radar
            key={key}
            name={key}
            dataKey={key}
            stroke={colors[index % colors.length]}
            fill={colors[index % colors.length]}
            fillOpacity={0.6}
          />
        ))}
      </RechartsRadarChart>
    </ResponsiveContainer>
  )
}

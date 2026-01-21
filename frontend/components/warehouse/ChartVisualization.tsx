'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { LineChart, Line, BarChart, Bar, PieChart, Pie, Cell, ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

interface ChartVisualizationProps {
  chartType: string
  chartConfig: any
  data: any[]
}

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8', '#82ca9d']

export default function ChartVisualization({ chartType, chartConfig, data }: ChartVisualizationProps) {
  if (!chartType || !chartConfig || !data || data.length === 0) {
    return (
      <Card>
        <CardContent className="text-center py-8 text-muted-foreground">
          No chart data available
        </CardContent>
      </Card>
    )
  }

  const renderChart = () => {
    switch (chartType) {
      case 'line':
        return (
          <ResponsiveContainer width="100%" height={400}>
            <LineChart data={data}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey={chartConfig.x} />
              <YAxis />
              <Tooltip />
              <Legend />
              {(chartConfig.series || [chartConfig.y]).map((series: string, index: number) => (
                <Line
                  key={series}
                  type="monotone"
                  dataKey={series}
                  stroke={COLORS[index % COLORS.length]}
                  name={series}
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        )

      case 'bar':
        return (
          <ResponsiveContainer width="100%" height={400}>
            <BarChart data={data}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey={chartConfig.x} />
              <YAxis />
              <Tooltip />
              <Legend />
              {(chartConfig.series || [chartConfig.y]).map((series: string, index: number) => (
                <Bar key={series} dataKey={series} fill={COLORS[index % COLORS.length]} name={series} />
              ))}
            </BarChart>
          </ResponsiveContainer>
        )

      case 'pie':
        return (
          <ResponsiveContainer width="100%" height={400}>
            <PieChart>
              <Pie
                data={data}
                cx="50%"
                cy="50%"
                labelLine={false}
                label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                outerRadius={120}
                fill="#8884d8"
                dataKey={chartConfig.value}
                nameKey={chartConfig.category}
              >
                {data.map((entry: any, index: number) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        )

      case 'scatter':
        return (
          <ResponsiveContainer width="100%" height={400}>
            <ScatterChart data={data}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey={chartConfig.x} />
              <YAxis dataKey={chartConfig.y} />
              <Tooltip cursor={{ strokeDasharray: '3 3' }} />
              <Scatter dataKey={chartConfig.y} fill="#8884d8" />
            </ScatterChart>
          </ResponsiveContainer>
        )

      default:
        return (
          <div className="text-center py-8 text-muted-foreground">
            Chart type "{chartType}" not supported
          </div>
        )
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Chart Visualization</CardTitle>
        <CardDescription>Visual representation of query results</CardDescription>
      </CardHeader>
      <CardContent>{renderChart()}</CardContent>
    </Card>
  )
}

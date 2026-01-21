'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'

export default function QueryPerformance() {
  const { data: performance, isLoading } = useQuery({
    queryKey: ['observability-query-performance'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.observabilityQueryPerformance, {
        params: { limit: 100 },
      })
      return response.data
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Query Performance</CardTitle>
          <CardDescription>Real-time query execution metrics</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">Loading...</div>
        </CardContent>
      </Card>
    )
  }

  const chartData = (performance || []).slice(0, 20).map((p: any) => ({
    time: new Date(p.executed_at).toLocaleTimeString(),
    duration: p.duration,
    rowCount: p.row_count,
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle>Query Performance</CardTitle>
        <CardDescription>Real-time query execution metrics</CardDescription>
      </CardHeader>
      <CardContent>
        {chartData.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">No performance data available</div>
        ) : (
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="duration" stroke="#8884d8" name="Duration (s)" />
              <Line type="monotone" dataKey="rowCount" stroke="#82ca9d" name="Row Count" />
            </LineChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}

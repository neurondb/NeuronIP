'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'

export default function MetricsDashboard() {
  const [timeWindow, setTimeWindow] = useState('5m')

  const { data: realtimeMetrics, isLoading } = useQuery({
    queryKey: ['observability-realtime', timeWindow],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.observabilityRealtime, {
        params: { window: timeWindow },
      })
      return response.data
    },
    refetchInterval: 5000, // Refresh every 5 seconds for real-time feel
  })

  const { data: systemMetrics } = useQuery({
    queryKey: ['observability-system-metrics'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.observabilityMetrics)
      return response.data
    },
    refetchInterval: 10000,
  })

  const chartData = (realtimeMetrics?.data_points || []).map((dp: any) => ({
    time: new Date(dp.timestamp).toLocaleTimeString(),
    avgLatency: dp.avg_latency,
    maxLatency: dp.max_latency,
    minLatency: dp.min_latency,
    queryCount: dp.query_count,
    successCount: dp.success_count,
    errorCount: dp.error_count,
  }))

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>System Metrics</CardTitle>
              <CardDescription>Overall system performance metrics</CardDescription>
            </div>
            <select
              value={timeWindow}
              onChange={(e) => setTimeWindow(e.target.value)}
              className="rounded-lg border border-border bg-background px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="1m">1 minute</option>
              <option value="5m">5 minutes</option>
              <option value="15m">15 minutes</option>
              <option value="1h">1 hour</option>
            </select>
          </div>
        </CardHeader>
        <CardContent>
          {systemMetrics && (
            <div className="grid grid-cols-3 gap-4 mb-4">
              <div className="p-4 rounded-lg border border-border">
                <div className="text-sm text-muted-foreground">Average Latency</div>
                <div className="text-2xl font-bold">{systemMetrics.latency?.toFixed(3) || 0}s</div>
              </div>
              <div className="p-4 rounded-lg border border-border">
                <div className="text-sm text-muted-foreground">Throughput</div>
                <div className="text-2xl font-bold">{systemMetrics.throughput?.toFixed(2) || 0}/min</div>
              </div>
              <div className="p-4 rounded-lg border border-border">
                <div className="text-sm text-muted-foreground">Cost</div>
                <div className="text-2xl font-bold">${systemMetrics.cost?.toFixed(2) || 0}</div>
              </div>
            </div>
          )}

          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading metrics...</div>
          ) : chartData.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">No metrics data available</div>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="time" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="avgLatency" stroke="#8884d8" name="Avg Latency (s)" />
                <Line type="monotone" dataKey="maxLatency" stroke="#ff7300" name="Max Latency (s)" />
                <Line type="monotone" dataKey="queryCount" stroke="#82ca9d" name="Query Count" />
              </LineChart>
            </ResponsiveContainer>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

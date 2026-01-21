'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'

export default function ComplianceDashboard() {
  const [timeRange, setTimeRange] = useState('30d')

  const { data: report, isLoading } = useQuery({
    queryKey: ['compliance-report', timeRange],
    queryFn: async () => {
      const endTime = new Date()
      const startTime = new Date()
      startTime.setDate(startTime.getDate() - (timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90))

      const response = await apiClient.get(API_ENDPOINTS.complianceReport, {
        params: {
          start_time: startTime.toISOString(),
          end_time: endTime.toISOString(),
        },
      })
      return response.data
    },
  })

  const chartData = (report?.violations || []).map((v: any) => ({
    name: v.policy_name,
    violations: v.count,
    avgScore: v.avg_match_score,
  }))

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Compliance Dashboard</CardTitle>
            <CardDescription>Monitor compliance violations and trends</CardDescription>
          </div>
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="rounded-lg border border-border bg-background px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="7d">Last 7 days</option>
            <option value="30d">Last 30 days</option>
            <option value="90d">Last 90 days</option>
          </select>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading compliance data...</div>
        ) : !report ? (
          <div className="text-center py-8 text-muted-foreground">No compliance data available</div>
        ) : (
          <>
            <div className="grid grid-cols-3 gap-4">
              <div className="p-4 rounded-lg border border-border">
                <div className="text-sm text-muted-foreground">Total Violations</div>
                <div className="text-2xl font-bold">{report.total_violations || 0}</div>
              </div>
              <div className="p-4 rounded-lg border border-border">
                <div className="text-sm text-muted-foreground">Policies Monitored</div>
                <div className="text-2xl font-bold">{report.violations?.length || 0}</div>
              </div>
              <div className="p-4 rounded-lg border border-border">
                <div className="text-sm text-muted-foreground">Period</div>
                <div className="text-sm font-medium">
                  {new Date(report.period_start).toLocaleDateString()} -{' '}
                  {new Date(report.period_end).toLocaleDateString()}
                </div>
              </div>
            </div>

            {chartData.length > 0 && (
              <div>
                <h3 className="text-sm font-medium mb-2">Violations by Policy</h3>
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="violations" fill="#8884d8" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}

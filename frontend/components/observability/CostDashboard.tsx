'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8', '#82ca9d']

export default function CostDashboard() {
  const [timeRange, setTimeRange] = useState('30d')
  const [groupBy, setGroupBy] = useState('category')

  const { data: breakdown, isLoading } = useQuery({
    queryKey: ['observability-cost-breakdown', timeRange, groupBy],
    queryFn: async () => {
      const endTime = new Date()
      const startTime = new Date()
      startTime.setDate(startTime.getDate() - (timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90))

      const response = await apiClient.get(API_ENDPOINTS.observabilityCostBreakdown, {
        params: {
          start_time: startTime.toISOString(),
          end_time: endTime.toISOString(),
          group_by: groupBy,
        },
      })
      return response.data
    },
  })

  const pieData = (breakdown || []).map((item: any) => ({
    name: item.group_key,
    value: item.total_cost,
  }))

  const barData = (breakdown || []).map((item: any) => ({
    name: item.group_key,
    cost: item.total_cost,
    count: item.record_count,
  }))

  const totalCost = (breakdown || []).reduce((sum: number, item: any) => sum + item.total_cost, 0)

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Cost Tracking</CardTitle>
            <CardDescription>Monitor and analyze system costs</CardDescription>
          </div>
          <div className="flex gap-2">
            <select
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value)}
              className="rounded-lg border border-border bg-background px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="7d">Last 7 days</option>
              <option value="30d">Last 30 days</option>
              <option value="90d">Last 90 days</option>
            </select>
            <select
              value={groupBy}
              onChange={(e) => setGroupBy(e.target.value)}
              className="rounded-lg border border-border bg-background px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="category">By Category</option>
              <option value="resource_type">By Resource Type</option>
              <option value="user">By User</option>
            </select>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading cost data...</div>
        ) : !breakdown || (Array.isArray(breakdown) && breakdown.length === 0) ? (
          <div className="text-center py-8 text-muted-foreground">No cost data available</div>
        ) : (
          <>
            <div className="text-2xl font-bold">
              Total Cost: ${totalCost.toFixed(2)}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <h3 className="text-sm font-medium mb-2">Cost Distribution</h3>
                <ResponsiveContainer width="100%" height={250}>
                  <PieChart>
                    <Pie
                      data={pieData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="value"
                    >
                      {pieData.map((entry: any, index: number) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </div>

              <div>
                <h3 className="text-sm font-medium mb-2">Cost by {groupBy === 'category' ? 'Category' : groupBy === 'resource_type' ? 'Resource Type' : 'User'}</h3>
                <ResponsiveContainer width="100%" height={250}>
                  <BarChart data={barData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip />
                    <Bar dataKey="cost" fill="#8884d8" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}

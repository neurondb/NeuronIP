'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { useBillingDashboard, useBillingUsage, useBillingMetrics } from '@/lib/api/queries'
import { ChartBarIcon, CreditCardIcon, ArrowTrendingUpIcon } from '@heroicons/react/24/outline'
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'

export default function BillingDashboard() {
  const { data: dashboardData, isLoading: dashboardLoading } = useBillingDashboard()
  const { data: usageData, isLoading: usageLoading } = useBillingUsage()
  const { data: metricsData, isLoading: metricsLoading } = useBillingMetrics()

  if (dashboardLoading || usageLoading || metricsLoading) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Loading billing data...</p>
        </CardContent>
      </Card>
    )
  }

  const metrics = metricsData || {}
  const usage = usageData?.metrics || []
  const dashboard = dashboardData || {}

  // Prepare chart data
  const chartData = usage.slice(0, 30).map((item: any) => ({
    date: new Date(item.timestamp || item.created_at).toLocaleDateString(),
    value: item.value || 0,
    name: item.metric_name || 'Usage',
  }))

  return (
    <div className="space-y-4">
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Usage</CardTitle>
            <ChartBarIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {dashboard.total_usage?.toLocaleString() || metrics.total_usage?.toLocaleString() || '0'}
            </div>
            <p className="text-xs text-muted-foreground mt-1">All time</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">This Month</CardTitle>
            <ArrowTrendingUpIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {dashboard.monthly_usage?.toLocaleString() || metrics.monthly_usage?.toLocaleString() || '0'}
            </div>
            <p className="text-xs text-muted-foreground mt-1">Current period</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">API Calls</CardTitle>
            <CreditCardIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {dashboard.api_calls?.toLocaleString() || metrics.api_calls?.toLocaleString() || '0'}
            </div>
            <p className="text-xs text-muted-foreground mt-1">Total calls</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Queries</CardTitle>
            <ChartBarIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {dashboard.queries?.toLocaleString() || metrics.queries?.toLocaleString() || '0'}
            </div>
            <p className="text-xs text-muted-foreground mt-1">Total queries</p>
          </CardContent>
        </Card>
      </div>

      {/* Usage Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Usage Over Time</CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="value" stroke="#3b82f6" name="Usage" />
            </LineChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* Usage Metrics Table */}
      {usage.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Usage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 max-h-[400px] overflow-y-auto">
              {usage.slice(0, 20).map((item: any, index: number) => (
                <div
                  key={item.id || index}
                  className="p-3 border border-border rounded-lg flex items-center justify-between"
                >
                  <div>
                    <p className="font-medium text-sm">{item.metric_name || 'Usage'}</p>
                    <p className="text-xs text-muted-foreground">
                      {item.metric_type || 'Unknown'} • {new Date(item.timestamp || item.created_at).toLocaleString()}
                    </p>
                  </div>
                  <div className="text-lg font-semibold">{item.value || 0}</div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
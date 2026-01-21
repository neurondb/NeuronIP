'use client'

import { useSearchAnalytics, useWarehouseAnalytics, useWorkflowAnalytics, useComplianceAnalytics } from '@/lib/api/queries'
import MetricCard from '@/components/dashboard/MetricCard'
import ActivityFeed from '@/components/dashboard/ActivityFeed'
import QuickActions from '@/components/dashboard/QuickActions'
import ChartContainer from '@/components/charts/ChartContainer'
import LineChart from '@/components/charts/LineChart'
import { motion } from 'framer-motion'
import {
  MagnifyingGlassIcon,
  CubeIcon,
  CommandLineIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline'
import { staggerContainer, slideUp, transition } from '@/lib/animations/variants'

export default function DashboardPage() {
  // Fetch analytics data
  const { data: searchAnalytics, isLoading: searchLoading } = useSearchAnalytics()
  const { data: warehouseAnalytics, isLoading: warehouseLoading } = useWarehouseAnalytics()
  const { data: workflowAnalytics, isLoading: workflowLoading } = useWorkflowAnalytics()
  const { data: complianceAnalytics, isLoading: complianceLoading } = useComplianceAnalytics()

  // Mock data for metrics (replace with real analytics)
  const metrics = [
    {
      title: 'Total Searches',
      value: searchAnalytics?.total_searches || '1,234',
      description: 'Last 30 days',
      icon: <MagnifyingGlassIcon />,
      trend: { value: 12, isPositive: true },
    },
    {
      title: 'Warehouse Queries',
      value: warehouseAnalytics?.total_queries || '567',
      description: 'Last 30 days',
      icon: <CubeIcon />,
      trend: { value: 8, isPositive: true },
    },
    {
      title: 'Workflows Executed',
      value: workflowAnalytics?.total_executions || '89',
      description: 'Last 30 days',
      icon: <CommandLineIcon />,
      trend: { value: 5, isPositive: true },
    },
    {
      title: 'Compliance Checks',
      value: complianceAnalytics?.total_checks || '234',
      description: 'Last 30 days',
      icon: <ShieldCheckIcon />,
      trend: { value: 3, isPositive: true },
    },
  ]

  // Mock chart data
  const chartData = [
    { date: 'Mon', searches: 45, queries: 32 },
    { date: 'Tue', searches: 52, queries: 38 },
    { date: 'Wed', searches: 48, queries: 35 },
    { date: 'Thu', searches: 61, queries: 42 },
    { date: 'Fri', searches: 55, queries: 40 },
    { date: 'Sat', searches: 38, queries: 28 },
    { date: 'Sun', searches: 42, queries: 30 },
  ]

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      {/* Page Header */}
      <div className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Dashboard</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Overview of your NeuronIP platform activity
        </p>
      </div>

      {/* Metrics Grid */}
      <motion.div
        variants={staggerContainer}
        className="grid grid-cols-2 lg:grid-cols-4 gap-2 sm:gap-3 flex-shrink-0"
      >
        {metrics.map((metric, index) => (
          <motion.div key={metric.title} variants={slideUp} transition={{ ...transition, delay: index * 0.05 }}>
            <MetricCard {...metric} />
          </motion.div>
        ))}
      </motion.div>

      {/* Charts and Activity - Fill remaining space */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3 sm:gap-4 flex-1 min-h-0">
        {/* Activity Chart - Takes 2 columns on large screens */}
        <motion.div variants={slideUp} className="lg:col-span-2 flex flex-col min-h-0">
          <ChartContainer
            title="Activity Overview"
            description="Search and query activity over the last week"
            className="flex-1 flex flex-col min-h-0"
          >
            <div className="flex-1 min-h-0" style={{ minHeight: '350px' }}>
              <LineChart
                data={chartData}
                dataKeys={['searches', 'queries']}
                xAxisKey="date"
                colors={['#0ea5e9', '#10b981']}
              />
            </div>
          </ChartContainer>
        </motion.div>

        {/* Activity Feed - Takes 1 column */}
        <motion.div variants={slideUp} className="flex flex-col min-h-0">
          <ActivityFeed />
        </motion.div>
      </div>

      {/* Quick Actions - Compact at bottom */}
      <motion.div variants={slideUp} className="flex-shrink-0">
        <QuickActions />
      </motion.div>
    </motion.div>
  )
}

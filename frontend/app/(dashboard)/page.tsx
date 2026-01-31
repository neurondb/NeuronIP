'use client'

import {
  MagnifyingGlassIcon,
  CubeIcon,
  CommandLineIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'

import ChartContainer from '@/components/charts/ChartContainer'
import { LazyLineChart } from '@/components/charts/LazyChart'
import ActivityFeed from '@/components/dashboard/ActivityFeed'
import MetricCard from '@/components/dashboard/MetricCard'
import QuickActions from '@/components/dashboard/QuickActions'
import QuickStart from '@/components/dashboard/QuickStart'
import PageTemplate from '@/components/layout/PageTemplate'
import { staggerContainer, slideUp, transition } from '@/lib/animations/variants'
import { useSearchAnalytics, useWarehouseAnalytics, useWorkflowAnalytics, useComplianceAnalytics } from '@/lib/api/queries'
import { microcopy } from '@/lib/copy/microcopy'
import { useOnboarding } from '@/lib/hooks/useOnboarding'

export default function DashboardPage() {
  const { isCompleted, skipped } = useOnboarding()
  const showQuickStart = !isCompleted || skipped

  const { data: searchAnalytics } = useSearchAnalytics()
  const { data: warehouseAnalytics } = useWarehouseAnalytics()
  const { data: workflowAnalytics } = useWorkflowAnalytics()
  const { data: complianceAnalytics } = useComplianceAnalytics()

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
    <PageTemplate
      title={microcopy.dashboard.title}
      description={microcopy.dashboard.subtitle}
      archetype="dashboard"
    >
      {showQuickStart && (
        <motion.div variants={slideUp} className="flex-shrink-0">
          <QuickStart />
        </motion.div>
      )}

      <motion.div
        variants={staggerContainer}
        className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4 flex-shrink-0"
      >
        {metrics.map((metric, index) => (
          <motion.div key={metric.title} variants={slideUp} transition={{ ...transition, delay: index * 0.05 }}>
            <MetricCard {...metric} />
          </motion.div>
        ))}
      </motion.div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 sm:gap-6 flex-1 min-h-0">
        <motion.div variants={slideUp} className="lg:col-span-2 flex flex-col min-h-0">
          <ChartContainer
            title="Activity Overview"
            description="Search and query activity over the last week"
            className="flex-1 flex flex-col min-h-0"
          >
            <div className="flex-1 min-h-0" style={{ minHeight: '350px' }}>
              <LazyLineChart
                data={chartData}
                dataKeys={['searches', 'queries']}
                xAxisKey="date"
                colors={['#0ea5e9', '#10b981']}
              />
            </div>
          </ChartContainer>
        </motion.div>
        <motion.div variants={slideUp} className="flex flex-col min-h-0">
          <ActivityFeed />
        </motion.div>
      </div>

      <motion.div variants={slideUp} className="flex-shrink-0">
        <QuickActions />
      </motion.div>
    </PageTemplate>
  )
}

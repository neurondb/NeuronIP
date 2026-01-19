'use client'

import { motion } from 'framer-motion'
import ChartContainer from '@/components/charts/ChartContainer'
import LineChart from '@/components/charts/LineChart'
import BarChart from '@/components/charts/BarChart'
import { useSearchAnalytics, useWarehouseAnalytics, useWorkflowAnalytics, useComplianceAnalytics } from '@/lib/api/queries'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function AnalyticsPage() {
  const { data: searchAnalytics } = useSearchAnalytics()
  const { data: warehouseAnalytics } = useWarehouseAnalytics()
  const { data: workflowAnalytics } = useWorkflowAnalytics()
  const { data: complianceAnalytics } = useComplianceAnalytics()

  const chartData = [
    { date: 'Mon', searches: 45, queries: 32, workflows: 12, compliance: 8 },
    { date: 'Tue', searches: 52, queries: 38, workflows: 15, compliance: 10 },
    { date: 'Wed', searches: 48, queries: 35, workflows: 13, compliance: 9 },
    { date: 'Thu', searches: 61, queries: 42, workflows: 18, compliance: 12 },
    { date: 'Fri', searches: 55, queries: 40, workflows: 16, compliance: 11 },
    { date: 'Sat', searches: 38, queries: 28, workflows: 10, compliance: 7 },
    { date: 'Sun', searches: 42, queries: 30, workflows: 11, compliance: 8 },
  ]

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Analytics</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Comprehensive analytics and insights across all modules
        </p>
      </motion.div>

      {/* Analytics Grid */}
      <div className="grid gap-3 sm:gap-4 lg:grid-cols-2 flex-1 min-h-0 overflow-y-auto">
        <motion.div variants={slideUp}>
          <ChartContainer title="Search Analytics" description="Search activity over time">
            <div className="h-[350px]">
              <LineChart
                data={chartData}
                dataKeys={['searches']}
                xAxisKey="date"
                colors={['#0ea5e9']}
              />
            </div>
          </ChartContainer>
        </motion.div>

        <motion.div variants={slideUp}>
          <ChartContainer title="Warehouse Analytics" description="Query activity over time">
            <div className="h-[350px]">
              <BarChart
                data={chartData}
                dataKeys={['queries']}
                xAxisKey="date"
                colors={['#10b981']}
              />
            </div>
          </ChartContainer>
        </motion.div>

        <motion.div variants={slideUp}>
          <ChartContainer title="Workflow Analytics" description="Workflow execution over time">
            <div className="h-[350px]">
              <LineChart
                data={chartData}
                dataKeys={['workflows']}
                xAxisKey="date"
                colors={['#8b5cf6']}
              />
            </div>
          </ChartContainer>
        </motion.div>

        <motion.div variants={slideUp}>
          <ChartContainer title="Compliance Analytics" description="Compliance checks over time">
            <div className="h-[350px]">
              <BarChart
                data={chartData}
                dataKeys={['compliance']}
                xAxisKey="date"
                colors={['#f59e0b']}
              />
            </div>
          </ChartContainer>
        </motion.div>
      </div>
    </motion.div>
  )
}

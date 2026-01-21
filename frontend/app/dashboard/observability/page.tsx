'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import QueryPerformance from '@/components/observability/QueryPerformance'
import LogViewer from '@/components/observability/LogViewer'
import CostDashboard from '@/components/observability/CostDashboard'
import MetricsDashboard from '@/components/observability/MetricsDashboard'
import Button from '@/components/ui/Button'

export default function ObservabilityPage() {
  const [activeTab, setActiveTab] = useState<'metrics' | 'performance' | 'logs' | 'cost'>('metrics')

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Observability</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Query performance, latency, cost. Agent logs, model logs, workflow logs.
            </p>
          </div>
        </div>
      </motion.div>

      <motion.div variants={slideUp} className="flex-shrink-0">
        <div className="flex gap-2 border-b border-border">
          <Button
            variant={activeTab === 'metrics' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('metrics')}
            size="sm"
          >
            Metrics
          </Button>
          <Button
            variant={activeTab === 'performance' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('performance')}
            size="sm"
          >
            Performance
          </Button>
          <Button
            variant={activeTab === 'logs' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('logs')}
            size="sm"
          >
            Logs
          </Button>
          <Button
            variant={activeTab === 'cost' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('cost')}
            size="sm"
          >
            Cost
          </Button>
        </div>
      </motion.div>

      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        {activeTab === 'metrics' && <MetricsDashboard />}
        {activeTab === 'performance' && <QueryPerformance />}
        {activeTab === 'logs' && <LogViewer />}
        {activeTab === 'cost' && <CostDashboard />}
      </motion.div>
    </motion.div>
  )
}

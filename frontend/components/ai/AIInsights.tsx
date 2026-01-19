'use client'

import { motion } from 'framer-motion'
import { SparklesIcon, LightBulbIcon, ChartBarIcon } from '@heroicons/react/24/outline'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { slideUp, transition } from '@/lib/animations/variants'

interface Insight {
  id: string
  type: 'trend' | 'anomaly' | 'recommendation'
  title: string
  description: string
  severity?: 'low' | 'medium' | 'high'
}

interface AIInsightsProps {
  insights?: Insight[]
}

export default function AIInsights({ insights }: AIInsightsProps) {
  // Mock AI insights - in production, fetch from AI API
  const mockInsights: Insight[] = insights || [
    {
      id: '1',
      type: 'trend',
      title: 'Query Volume Increase',
      description: 'Search queries have increased by 23% compared to last week',
      severity: 'low',
    },
    {
      id: '2',
      type: 'anomaly',
      title: 'Unusual Pattern Detected',
      description: 'Workflow execution time increased significantly at 3 PM',
      severity: 'high',
    },
    {
      id: '3',
      type: 'recommendation',
      title: 'Optimization Opportunity',
      description: 'Consider caching frequently accessed warehouse schemas',
      severity: 'medium',
    },
  ]

  const getIcon = (type: Insight['type']) => {
    switch (type) {
      case 'trend':
        return <ChartBarIcon className="h-5 w-5" />
      case 'anomaly':
        return <LightBulbIcon className="h-5 w-5" />
      case 'recommendation':
        return <SparklesIcon className="h-5 w-5" />
    }
  }

  const getSeverityColor = (severity?: Insight['severity']) => {
    switch (severity) {
      case 'high':
        return 'text-red-600 dark:text-red-400'
      case 'medium':
        return 'text-yellow-600 dark:text-yellow-400'
      case 'low':
        return 'text-green-600 dark:text-green-400'
      default:
        return 'text-muted-foreground'
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <SparklesIcon className="h-5 w-5" />
          AI Insights
        </CardTitle>
        <CardDescription>Automated insights and recommendations</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {mockInsights.map((insight, index) => (
            <motion.div
              key={insight.id}
              variants={slideUp}
              initial="hidden"
              animate="visible"
              transition={{ ...transition, delay: index * 0.1 }}
              className="flex gap-3 p-3 rounded-lg border border-border hover:bg-accent transition-colors"
            >
              <div className={getSeverityColor(insight.severity)}>{getIcon(insight.type)}</div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium">{insight.title}</p>
                <p className="text-xs text-muted-foreground mt-1">{insight.description}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

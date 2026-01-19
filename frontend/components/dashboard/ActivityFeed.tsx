'use client'

import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { slideUp, transition } from '@/lib/animations/variants'

interface Activity {
  id: string
  type: 'search' | 'query' | 'workflow' | 'compliance'
  message: string
  timestamp: Date
}

export default function ActivityFeed() {
  const [activities, setActivities] = useState<Activity[]>([])

  useEffect(() => {
    // Simulated activities - replace with real data from API
    const mockActivities: Activity[] = [
      {
        id: '1',
        type: 'search',
        message: 'Semantic search executed for "customer data"',
        timestamp: new Date(Date.now() - 1000 * 60 * 5),
      },
      {
        id: '2',
        type: 'query',
        message: 'Warehouse query completed successfully',
        timestamp: new Date(Date.now() - 1000 * 60 * 15),
      },
      {
        id: '3',
        type: 'workflow',
        message: 'Workflow "Data Processing" executed',
        timestamp: new Date(Date.now() - 1000 * 60 * 30),
      },
    ]
    setActivities(mockActivities)

    // Poll for new activities
    const interval = setInterval(() => {
      // In production, fetch from API
    }, 30000)

    return () => clearInterval(interval)
  }, [])

  const getActivityIcon = (type: Activity['type']) => {
    switch (type) {
      case 'search':
        return '🔍'
      case 'query':
        return '📊'
      case 'workflow':
        return '⚙️'
      case 'compliance':
        return '✅'
      default:
        return '📝'
    }
  }

  const formatTime = (date: Date) => {
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)

    if (minutes < 1) return 'Just now'
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    if (hours < 24) return `${hours}h ago`
    const days = Math.floor(hours / 24)
    return `${days}d ago`
  }

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">Recent Activity</CardTitle>
        <CardDescription className="text-xs">Latest system events</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto min-h-0">
        <div className="space-y-2.5 sm:space-y-3">
          {activities.map((activity, index) => (
            <motion.div
              key={activity.id}
              variants={slideUp}
              initial="hidden"
              animate="visible"
              transition={{ ...transition, delay: index * 0.05 }}
              className="flex items-start gap-2.5 pb-2.5 border-b border-border last:border-0 last:pb-0"
            >
              <span className="text-base sm:text-lg flex-shrink-0">{getActivityIcon(activity.type)}</span>
              <div className="flex-1 min-w-0">
                <p className="text-xs sm:text-sm text-foreground leading-tight">{activity.message}</p>
                <p className="text-xs text-muted-foreground mt-0.5">{formatTime(activity.timestamp)}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import {
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ServerIcon,
  CpuChipIcon,
} from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import SimpleFooter from './SimpleFooter'

interface DashboardFooterProps {
  className?: string
  showStats?: boolean
  version?: string
}

interface SystemStats {
  status: 'healthy' | 'warning' | 'error'
  database: 'connected' | 'disconnected'
  api: 'online' | 'offline'
  uptime?: string
}

export default function DashboardFooter({
  className,
  showStats = true,
  version = '1.0.0',
}: DashboardFooterProps) {
  const [stats, setStats] = useState<SystemStats>({
    status: 'healthy',
    database: 'connected',
    api: 'online',
    uptime: '99.9%',
  })

  // Fetch system stats (mock for now - can be connected to health endpoint)
  useEffect(() => {
    // In a real implementation, fetch from /api/v1/health or similar endpoint
    const fetchStats = async () => {
      try {
        const response = await fetch('/api/v1/health')
        if (response.ok) {
          const data = await response.json()
          setStats({
            status: data.status === 'ok' ? 'healthy' : 'warning',
            database: data.checks?.database?.status === 'healthy' ? 'connected' : 'disconnected',
            api: 'online',
            uptime: '99.9%',
          })
        }
      } catch (error) {
        // Silently fail - use default stats
        console.debug('Failed to fetch health stats:', error)
      }
    }

    if (showStats) {
      fetchStats()
      // Refresh stats every 30 seconds
      const interval = setInterval(fetchStats, 30000)
      return () => clearInterval(interval)
    }
  }, [showStats])

  const statusColors = {
    healthy: 'text-green-600 dark:text-green-400',
    warning: 'text-yellow-600 dark:text-yellow-400',
    error: 'text-red-600 dark:text-red-400',
  }

  const StatusIcon = stats.status === 'healthy' ? CheckCircleIcon : XCircleIcon

  return (
    <footer
      className={cn(
        'border-t border-border bg-card/50 backdrop-blur supports-[backdrop-filter]:bg-card/30',
        className
      )}
    >
      <div className="py-6 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          {showStats && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3 }}
              className="flex flex-wrap items-center justify-between gap-4 mb-4 pb-4 border-b border-border"
            >
              {/* System Status */}
              <div className="flex flex-wrap items-center gap-4 sm:gap-6">
                <div className="flex items-center gap-2">
                  <StatusIcon className={cn('h-4 w-4', statusColors[stats.status])} />
                  <span className="text-sm text-muted-foreground">
                    System Status: <span className={cn('font-medium', statusColors[stats.status])}>
                      {stats.status.charAt(0).toUpperCase() + stats.status.slice(1)}
                    </span>
                  </span>
                </div>

                {/* Database Status */}
                <div className="flex items-center gap-2">
                  <ServerIcon className={cn(
                    'h-4 w-4',
                    stats.database === 'connected' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
                  )} />
                  <span className="text-sm text-muted-foreground">
                    DB: <span className={cn(
                      'font-medium',
                      stats.database === 'connected' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
                    )}>
                      {stats.database === 'connected' ? 'Connected' : 'Disconnected'}
                    </span>
                  </span>
                </div>

                {/* API Status */}
                <div className="flex items-center gap-2">
                  <CpuChipIcon className={cn(
                    'h-4 w-4',
                    stats.api === 'online' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
                  )} />
                  <span className="text-sm text-muted-foreground">
                    API: <span className={cn(
                      'font-medium',
                      stats.api === 'online' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
                    )}>
                      {stats.api === 'online' ? 'Online' : 'Offline'}
                    </span>
                  </span>
                </div>

                {/* Uptime */}
                {stats.uptime && (
                  <div className="flex items-center gap-2">
                    <ClockIcon className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">
                      Uptime: <span className="font-medium text-foreground">{stats.uptime}</span>
                    </span>
                  </div>
                )}
              </div>

              {/* Version */}
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground">
                  v{version}
                </span>
              </div>
            </motion.div>
          )}

          {/* Simple Footer */}
          <SimpleFooter showLinks={true} />
        </div>
      </div>
    </footer>
  )
}

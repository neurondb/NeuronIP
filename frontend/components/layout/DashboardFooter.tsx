'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import {
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ServerIcon,
  CpuChipIcon,
  PuzzlePieceIcon,
} from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import SimpleFooter from './SimpleFooter'
import StatusBadge from '@/components/ui/StatusBadge'
import { HealthResponse, SystemStats } from '@/lib/api/types'
import Tooltip from '@/components/ui/Tooltip'

interface DashboardFooterProps {
  className?: string
  showStats?: boolean
  version?: string
}

export default function DashboardFooter({
  className,
  showStats = true,
  version = '1.0.0',
}: DashboardFooterProps) {
  const [stats, setStats] = useState<SystemStats>({
    status: 'healthy',
    database: 'connected',
    mcp: 'unavailable',
    api: 'online',
    uptime: undefined,
  })
  const [healthData, setHealthData] = useState<HealthResponse | null>(null)

  // Fetch system stats from health endpoint
  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch('/api/v1/health')
        if (response.ok) {
          const data: HealthResponse = await response.json()
          setHealthData(data)

          // Parse health response to system stats
          const dbCheck = data.checks?.database
          const mcpCheck = data.checks?.mcp
          const uptimeCheck = data.checks?.uptime

          const newStats: SystemStats = {
            status:
              data.status === 'ok'
                ? 'healthy'
                : data.status === 'warning'
                ? 'warning'
                : 'error',
            database:
              dbCheck?.status === 'healthy' ? 'connected' : 'disconnected',
            mcp:
              mcpCheck?.status === 'healthy'
                ? 'connected'
                : mcpCheck?.status === 'error'
                ? 'disconnected'
                : 'unavailable',
            api: 'online',
            uptime: uptimeCheck?.message?.replace('Server uptime: ', '') || undefined,
          }

          setStats(newStats)
        } else {
          // API is offline
          setStats((prev) => ({
            ...prev,
            api: 'offline',
            status: 'error',
          }))
        }
      } catch (error) {
        // Silently fail - use default stats or mark as offline
        console.debug('Failed to fetch health stats:', error)
        setStats((prev) => ({
          ...prev,
          api: 'offline',
          status: 'error',
        }))
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
                {/* System Status Badge */}
                <Tooltip
                  content={
                    <div>
                      <p className="font-medium mb-1">System Status</p>
                      <p className="text-xs">
                        {stats.status === 'healthy'
                          ? 'All systems operational'
                          : stats.status === 'warning'
                          ? 'Some services may be degraded'
                          : 'System errors detected'}
                      </p>
                      {healthData && (
                        <div className="mt-2 space-y-1 text-xs">
                          {Object.entries(healthData.checks || {}).map(([key, check]) => (
                            check && (
                              <div key={key}>
                                <span className="capitalize">{key}:</span>{' '}
                                <span
                                  className={
                                    check.status === 'healthy'
                                      ? 'text-green-300'
                                      : check.status === 'error'
                                      ? 'text-red-300'
                                      : 'text-yellow-300'
                                  }
                                >
                                  {check.status}
                                </span>
                              </div>
                            )
                          ))}
                        </div>
                      )}
                    </div>
                  }
                  variant="info"
                >
                  <div className="flex items-center gap-2 cursor-pointer">
                    <StatusIcon className={cn('h-4 w-4', statusColors[stats.status])} />
                    <span className="text-sm text-muted-foreground">
                      System: <span className={cn('font-medium', statusColors[stats.status])}>
                        {stats.status.charAt(0).toUpperCase() + stats.status.slice(1)}
                      </span>
                    </span>
                  </div>
                </Tooltip>

                {/* PostgreSQL Status */}
                <StatusBadge
                  status={
                    stats.database === 'connected'
                      ? 'connected'
                      : stats.database === 'disconnected'
                      ? 'disconnected'
                      : 'error'
                  }
                  label="PostgreSQL"
                  details={
                    <div>
                      <p className="font-medium mb-1">PostgreSQL Connection</p>
                      <p className="text-xs">
                        {healthData?.checks?.database?.message || 'Connection status unknown'}
                      </p>
                    </div>
                  }
                  size="sm"
                  showIcon={true}
                />

                {/* MCP Status */}
                <StatusBadge
                  status={
                    stats.mcp === 'connected'
                      ? 'connected'
                      : stats.mcp === 'disconnected'
                      ? 'disconnected'
                      : stats.mcp === 'unavailable'
                      ? 'error'
                      : 'error'
                  }
                  label="MCP"
                  details={
                    <div>
                      <p className="font-medium mb-1">MCP (Model Context Protocol)</p>
                      <p className="text-xs">
                        {healthData?.checks?.mcp?.message ||
                          'MCP client not configured or unavailable'}
                      </p>
                    </div>
                  }
                  size="sm"
                  showIcon={true}
                />

                {/* API Status */}
                <StatusBadge
                  status={stats.api === 'online' ? 'connected' : 'disconnected'}
                  label="API"
                  details={
                    <div>
                      <p className="font-medium mb-1">API Service</p>
                      <p className="text-xs">
                        {stats.api === 'online'
                          ? 'API service is responding'
                          : 'API service is not responding'}
                      </p>
                    </div>
                  }
                  size="sm"
                  showIcon={true}
                />

                {/* Uptime */}
                {stats.uptime && (
                  <Tooltip
                    content={
                      <div>
                        <p className="font-medium mb-1">Server Uptime</p>
                        <p className="text-xs">Time since last server restart</p>
                      </div>
                    }
                    variant="info"
                  >
                    <div className="flex items-center gap-2 cursor-pointer">
                      <ClockIcon className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">
                        Uptime: <span className="font-medium text-foreground">{stats.uptime}</span>
                      </span>
                    </div>
                  </Tooltip>
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

'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import {
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ServerIcon,
  CpuChipIcon,
  CubeIcon,
  MagnifyingGlassIcon,
  CommandLineIcon,
  SparklesIcon,
  SignalIcon,
  UserCircleIcon,
  BellIcon,
  CircleStackIcon,
  PuzzlePieceIcon,
} from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import StatusBadge from '@/components/ui/StatusBadge'
import { HealthResponse, SystemStats } from '@/lib/api/types'
import Tooltip from '@/components/ui/Tooltip'
import { useWebSocket } from '@/lib/hooks/useWebSocket'
import { useAppStore } from '@/lib/store/useAppStore'

interface StatusBarProps {
  version?: string
  className?: string
}

interface ServiceStatus {
  name: string
  status: 'connected' | 'disconnected' | 'error' | 'warning'
  icon: React.ComponentType<{ className?: string }>
  checkKey?: string
}

export default function StatusBar({ version = '1.0.0', className }: StatusBarProps) {
  const { sidebarCollapsed } = useAppStore()
  const [stats, setStats] = useState<SystemStats>({
    status: 'healthy',
    database: 'connected',
    mcp: 'unavailable',
    api: 'online',
    uptime: undefined,
  })
  const [healthData, setHealthData] = useState<HealthResponse | null>(null)
  const [notificationCount, setNotificationCount] = useState(0)
  const [userName, setUserName] = useState<string>('User')

  // WebSocket connection status
  const { isConnected: wsConnected } = useWebSocket({
    url: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws',
    enabled: true,
  })

  // Fetch system stats from health endpoint
  useEffect(() => {
    const fetchStats = async () => {
      try {
        const controller = new AbortController()
        const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout

        const response = await fetch('/api/v1/health', {
          signal: controller.signal,
          headers: {
            'Accept': 'application/json',
          },
        })

        clearTimeout(timeoutId)

        if (response.ok) {
          const data: HealthResponse = await response.json()
          setHealthData(data)

          const dbCheck = data.checks?.database
          const mcpCheck = data.checks?.mcp
          const uptimeCheck = data.checks?.uptime

          // Determine overall status based on checks
          let overallStatus: 'healthy' | 'warning' | 'error' = 'healthy'
          
          if (data.status === 'unhealthy') {
            overallStatus = 'error'
          } else if (data.status === 'warning') {
            overallStatus = 'warning'
          } else if (dbCheck?.status === 'error') {
            overallStatus = 'error'
          } else if (dbCheck?.status === 'warning' || mcpCheck?.status === 'warning') {
            overallStatus = 'warning'
          }

          const newStats: SystemStats = {
            status: overallStatus,
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
          // API responded but with error status
          setStats((prev) => ({
            ...prev,
            api: 'offline',
            status: 'error',
          }))
        }
      } catch (error) {
        // Network error or timeout
        if (error instanceof Error && error.name === 'AbortError') {
          // Timeout - consider API as slow/offline
          setStats((prev) => ({
            ...prev,
            api: 'offline',
            status: 'warning',
          }))
        } else {
          // Other errors - API is offline
          setStats((prev) => ({
            ...prev,
            api: 'offline',
            status: 'error',
          }))
        }
        console.debug('Failed to fetch health stats:', error)
      }
    }

    // Fetch immediately
    fetchStats()
    // Refresh stats every 15 seconds
    const interval = setInterval(fetchStats, 15000)
    return () => clearInterval(interval)
  }, [])

  // Get user info
  useEffect(() => {
    const user = localStorage.getItem('user')
    if (user) {
      try {
        const userData = JSON.parse(user)
        setUserName(userData.name || userData.email || 'User')
      } catch {
        // Ignore parse errors
      }
    }
  }, [])

  // Mock notification count (replace with real data)
  useEffect(() => {
    // In production, fetch from API
    setNotificationCount(0)
  }, [])

  const getServiceStatus = (checkKey?: string): 'connected' | 'disconnected' | 'error' | 'warning' => {
    if (!checkKey || !healthData?.checks?.[checkKey]) {
      return 'disconnected'
    }
    const check = healthData.checks[checkKey]
    if (check?.status === 'healthy') return 'connected'
    if (check?.status === 'warning') return 'warning'
    return 'error'
  }

  const services: ServiceStatus[] = [
    {
      name: 'Database',
      status: stats.database === 'connected' ? 'connected' : 'disconnected',
      icon: CircleStackIcon,
      checkKey: 'database',
    },
    {
      name: 'Warehouse',
      status: getServiceStatus('warehouse'),
      icon: CubeIcon,
      checkKey: 'warehouse',
    },
    {
      name: 'Semantic',
      status: getServiceStatus('semantic'),
      icon: MagnifyingGlassIcon,
      checkKey: 'semantic',
    },
    {
      name: 'Workflows',
      status: getServiceStatus('workflows'),
      icon: CommandLineIcon,
      checkKey: 'workflows',
    },
    {
      name: 'Agents',
      status: getServiceStatus('agents'),
      icon: SparklesIcon,
      checkKey: 'agents',
    },
    {
      name: 'MCP',
      status:
        stats.mcp === 'connected'
          ? 'connected'
          : stats.mcp === 'disconnected'
          ? 'disconnected'
          : 'error',
      icon: PuzzlePieceIcon,
      checkKey: 'mcp',
    },
    {
      name: 'WebSocket',
      status: wsConnected ? 'connected' : 'disconnected',
      icon: SignalIcon,
    },
  ]

  const statusColors = {
    healthy: 'text-green-600 dark:text-green-400',
    warning: 'text-yellow-600 dark:text-yellow-400',
    error: 'text-red-600 dark:text-red-400',
  }

  const StatusIcon = stats.status === 'healthy' ? CheckCircleIcon : XCircleIcon

  // Sidebar width: 256px (w-64) when expanded, 64px (w-16) when collapsed on lg screens
  // On mobile (< lg), sidebar is hidden (translate-x-full)
  const sidebarLeft = sidebarCollapsed ? 'lg:left-16' : 'lg:left-64'

  return (
    <footer
      className={cn(
        'h-8 border-t border-border bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/80',
        'flex items-center justify-between px-3 sm:px-4 text-xs',
        'fixed bottom-0 left-0 right-0 z-50 transition-all duration-300',
        sidebarLeft,
        className
      )}
    >
      {/* Left side - System Status & Services */}
      <div className="flex items-center gap-2 sm:gap-3 overflow-x-auto scrollbar-hide">
        {/* Overall System Status */}
        <Tooltip
          content={
            <div>
              <p className="font-medium mb-1">System Status</p>
              <p className="text-xs mb-2">
                {stats.status === 'healthy'
                  ? 'All systems operational'
                  : stats.status === 'warning'
                  ? 'Some services may be degraded'
                  : 'System errors detected'}
              </p>
              {healthData && (
                <div className="mt-2 space-y-1 text-xs border-t border-border/50 pt-2">
                  {Object.entries(healthData.checks || {}).map(([key, check]) => (
                    check && (
                      <div key={key} className="flex items-center justify-between gap-2">
                        <span className="capitalize">{key}:</span>
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
              {stats.api === 'offline' && (
                <p className="text-xs mt-2 text-red-300">
                  API service is not responding. Check your connection.
                </p>
              )}
            </div>
          }
          variant="info"
        >
          <div className="flex items-center gap-1.5 cursor-pointer flex-shrink-0">
            <StatusIcon className={cn('h-3.5 w-3.5', statusColors[stats.status])} />
            <span className={cn('font-medium hidden sm:inline', statusColors[stats.status])}>
              {stats.status.charAt(0).toUpperCase() + stats.status.slice(1)}
            </span>
          </div>
        </Tooltip>

        <div className="h-4 w-px bg-border flex-shrink-0" />

        {/* Service Statuses */}
        {services.map((service) => {
          const Icon = service.icon
          const serviceStatus = service.status
          const check = service.checkKey ? healthData?.checks?.[service.checkKey] : null

          return (
            <Tooltip
              key={service.name}
              content={
                <div>
                  <p className="font-medium mb-1">{service.name} Service</p>
                  <p className="text-xs">
                    {serviceStatus === 'connected'
                      ? 'Service is operational'
                      : serviceStatus === 'warning'
                      ? 'Service may be degraded'
                      : 'Service is unavailable'}
                  </p>
                  {check?.message && (
                    <p className="text-xs mt-1 text-muted-foreground">{check.message}</p>
                  )}
                </div>
              }
              variant="info"
            >
              <div className="flex items-center gap-1.5 cursor-pointer flex-shrink-0">
                <Icon
                  className={cn(
                    'h-3.5 w-3.5',
                    serviceStatus === 'connected'
                      ? 'text-green-500 dark:text-green-400'
                      : serviceStatus === 'warning'
                      ? 'text-yellow-500 dark:text-yellow-400'
                      : 'text-red-500 dark:text-red-400'
                  )}
                />
                <span className="text-muted-foreground hidden lg:inline text-xs">
                  {service.name}
                </span>
              </div>
            </Tooltip>
          )
        })}
      </div>

      {/* Right side - User Info, Notifications, Uptime, Version */}
      <div className="flex items-center gap-2 sm:gap-3 flex-shrink-0">
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
            <div className="flex items-center gap-1.5 cursor-pointer hidden md:flex">
              <ClockIcon className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="text-muted-foreground text-xs">{stats.uptime}</span>
            </div>
          </Tooltip>
        )}

        {/* Notifications */}
        {notificationCount > 0 && (
          <Tooltip content={`${notificationCount} unread notifications`} variant="info">
            <div className="flex items-center gap-1.5 cursor-pointer">
              <div className="relative">
                <BellIcon className="h-3.5 w-3.5 text-muted-foreground" />
                {notificationCount > 0 && (
                  <span className="absolute -top-1 -right-1 h-2 w-2 bg-red-500 rounded-full" />
                )}
              </div>
              <span className="text-muted-foreground text-xs hidden lg:inline">
                {notificationCount}
              </span>
            </div>
          </Tooltip>
        )}

        {/* User */}
        <Tooltip content={`Logged in as ${userName}`} variant="info">
          <div className="flex items-center gap-1.5 cursor-pointer">
            <UserCircleIcon className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-muted-foreground text-xs hidden lg:inline truncate max-w-[100px]">
              {userName}
            </span>
          </div>
        </Tooltip>

        <div className="h-4 w-px bg-border" />

        {/* Version */}
        <span className="text-muted-foreground text-xs font-mono">v{version}</span>
      </div>
    </footer>
  )
}

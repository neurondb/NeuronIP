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
  ArrowPathIcon,
  StopIcon,
  InformationCircleIcon,
  TagIcon,
} from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import StatusBadge from '@/components/ui/StatusBadge'
import { HealthResponse, SystemStats } from '@/lib/api/types'
import Tooltip from '@/components/ui/Tooltip'
import { useWebSocket } from '@/lib/hooks/useWebSocket'
import { useAppStore } from '@/lib/store/useAppStore'
import { showToast } from '@/components/ui/Toast'

interface StatusBarProps {
  version?: string
  className?: string
}

interface ServiceStatus {
  name: string
  label: string
  status: 'connected' | 'disconnected' | 'error' | 'warning' | 'connecting'
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
  const [connecting, setConnecting] = useState<Set<string>>(new Set())

  // WebSocket connection status
  const { status } = useWebSocket()
  const wsConnected = status === 'connected'

  // Fetch system stats from health endpoint
  const fetchStats = async (showConnecting?: boolean): Promise<SystemStats | null> => {
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

      // If we get a response (even non-200), API is online
      // The API itself is healthy if we can reach it
      let data: HealthResponse
      try {
        data = await response.json()
      } catch (parseError) {
        // If response is not JSON, still consider API online
        const newStats: SystemStats = {
          ...stats,
          api: 'online',
          status: 'warning',
        }
        setStats(newStats)
        if (showConnecting) {
          setConnecting((prev) => {
            const next = new Set(prev)
            next.delete('postgresql')
            next.delete('mcp')
            next.delete('api')
            return next
          })
        }
        return newStats
      }

      setHealthData(data)

      const dbCheck = data.checks?.database
      const mcpCheck = data.checks?.mcp
      const uptimeCheck = data.checks?.uptime

      // Determine overall status based on checks
      // API is always "online" if we got a response
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
        api: 'online', // API is always online if we got a response
        uptime: uptimeCheck?.message?.replace('Server uptime: ', '') || undefined,
      }

      setStats(newStats)
      
      if (showConnecting) {
        setConnecting((prev) => {
          const next = new Set(prev)
          next.delete('postgresql')
          next.delete('mcp')
          next.delete('api')
          return next
        })
      }
      
      return newStats
    } catch (error) {
      // Network error or timeout - API is truly offline
      const newStats: SystemStats = {
        ...stats,
        api: 'offline', // Only mark offline if we can't reach the endpoint
        status: error instanceof Error && error.name === 'AbortError' ? 'warning' : 'error',
      }
      setStats(newStats)
      console.debug('Failed to fetch health stats:', error)
      if (showConnecting) {
        setConnecting((prev) => {
          const next = new Set(prev)
          next.delete('postgresql')
          next.delete('mcp')
          next.delete('api')
          return next
        })
      }
      return newStats
    }
  }

  useEffect(() => {
    // Fetch immediately
    fetchStats()
    // Refresh stats every 15 seconds
    const interval = setInterval(() => fetchStats(), 15000)
    return () => clearInterval(interval)
  }, [])

  // Handle service reconnect
  const handleReconnect = async (service: 'postgresql' | 'mcp' | 'api') => {
    setConnecting((prev) => new Set(prev).add(service))
    
    try {
      // Force refresh health check
      const updatedStats = await fetchStats(true)
      
      if (updatedStats) {
        const isConnected = 
          (service === 'postgresql' && updatedStats.database === 'connected') ||
          (service === 'mcp' && updatedStats.mcp === 'connected') ||
          (service === 'api' && updatedStats.api === 'online')
        
        if (isConnected || service === 'api') {
          showToast(`${service === 'postgresql' ? 'PostgreSQL' : service.toUpperCase()} reconnected successfully`, 'success')
        } else {
          showToast(`Failed to reconnect to ${service === 'postgresql' ? 'PostgreSQL' : service.toUpperCase()}`, 'error')
        }
      } else {
        showToast(`Error checking ${service === 'postgresql' ? 'PostgreSQL' : service.toUpperCase()} status`, 'error')
      }
    } catch (error) {
      showToast(`Error reconnecting to ${service === 'postgresql' ? 'PostgreSQL' : service.toUpperCase()}`, 'error')
      setConnecting((prev) => {
        const next = new Set(prev)
        next.delete(service)
        return next
      })
    }
  }

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

  // Main services to display prominently
  const mainServices: ServiceStatus[] = [
    {
      name: 'postgresql',
      label: 'PostgreSQL',
      status: connecting.has('postgresql')
        ? 'connecting'
        : stats.database === 'connected'
        ? 'connected'
        : 'disconnected',
      icon: CircleStackIcon,
      checkKey: 'database',
    },
    {
      name: 'mcp',
      label: 'MCP',
      status: connecting.has('mcp')
        ? 'connecting'
        : stats.mcp === 'connected'
        ? 'connected'
        : stats.mcp === 'disconnected'
        ? 'disconnected'
        : 'error',
      icon: PuzzlePieceIcon,
      checkKey: 'mcp',
    },
    {
      name: 'api',
      label: 'API',
      status: connecting.has('api')
        ? 'connecting'
        : stats.api === 'online'
        ? 'connected'
        : 'disconnected',
      icon: ServerIcon,
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
        'h-8 border-t border-border/50',
        'bg-muted/50 dark:bg-muted/30 backdrop-blur-sm',
        'flex items-center justify-between px-3 sm:px-4 text-xs',
        'fixed bottom-0 left-0 right-0 z-50 transition-all duration-300',
        'shadow-[0_-1px_3px_rgba(0,0,0,0.1)] dark:shadow-[0_-1px_3px_rgba(0,0,0,0.3)]',
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
          <div className="flex items-center gap-1.5 cursor-pointer flex-shrink-0 px-1.5 py-0.5 rounded hover:bg-background/30 dark:hover:bg-background/20">
            <StatusIcon className={cn('h-3.5 w-3.5 flex-shrink-0', statusColors[stats.status])} />
            <span className={cn('font-medium hidden sm:inline text-xs', statusColors[stats.status])}>
              {stats.status.charAt(0).toUpperCase() + stats.status.slice(1)}
            </span>
          </div>
        </Tooltip>

        <div className="h-4 w-px bg-border/50 flex-shrink-0" />

        {/* Main Service Statuses - PostgreSQL, MCP, API */}
        {mainServices.map((service) => {
          const Icon = service.icon
          const serviceStatus = service.status
          const check = service.checkKey ? healthData?.checks?.[service.checkKey] : null
          const isConnecting = connecting.has(service.name)
          const isConnected = serviceStatus === 'connected'

          return (
            <Tooltip
              key={service.name}
              content={
                <div>
                  <p className="font-medium mb-1">{service.label} Service</p>
                  <p className="text-xs mb-2">
                    {serviceStatus === 'connected' || isConnecting
                      ? isConnecting
                        ? 'Reconnecting...'
                        : 'Service is operational'
                      : serviceStatus === 'warning'
                      ? 'Service may be degraded'
                      : 'Service is unavailable'}
                  </p>
                  {check?.message && (
                    <p className="text-xs mb-2 text-muted-foreground">{check.message}</p>
                  )}
                  <div className="mt-2 pt-2 border-t border-border/50">
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleReconnect(service.name as 'postgresql' | 'mcp' | 'api')
                      }}
                      disabled={isConnecting}
                      className="text-xs text-blue-400 hover:text-blue-300 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                    >
                      {isConnecting ? (
                        <>
                          <ArrowPathIcon className="h-3 w-3 animate-spin" />
                          Reconnecting...
                        </>
                      ) : isConnected ? (
                        <>
                          <ArrowPathIcon className="h-3 w-3" />
                          Reconnect
                        </>
                      ) : (
                        <>
                          <ArrowPathIcon className="h-3 w-3" />
                          Connect
                        </>
                      )}
                    </button>
                  </div>
                </div>
              }
              variant="info"
            >
              <button
                onClick={() => handleReconnect(service.name as 'postgresql' | 'mcp' | 'api')}
                disabled={isConnecting}
                className={cn(
                  'flex items-center gap-1.5 flex-shrink-0 rounded px-2 py-0.5 transition-all',
                  'hover:bg-background/30 dark:hover:bg-background/20',
                  'disabled:opacity-50 disabled:cursor-not-allowed',
                  isConnecting && 'cursor-wait'
                )}
              >
                {isConnecting ? (
                  <ArrowPathIcon className="h-3.5 w-3.5 text-yellow-500 dark:text-yellow-400 animate-spin flex-shrink-0" />
                ) : (
                  <Icon
                    className={cn(
                      'h-3.5 w-3.5 flex-shrink-0',
                      serviceStatus === 'connected'
                        ? 'text-green-600 dark:text-green-400'
                        : serviceStatus === 'warning'
                        ? 'text-yellow-600 dark:text-yellow-400'
                        : 'text-red-600 dark:text-red-400'
                    )}
                  />
                )}
                <span className="text-foreground/90 dark:text-foreground/80 hidden lg:inline text-xs font-medium">
                  {service.label}
                </span>
                {isConnected && !isConnecting && (
                  <span className="h-1.5 w-1.5 rounded-full bg-green-500 dark:bg-green-400 flex-shrink-0" />
                )}
              </button>
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
            <div className="flex items-center gap-1.5 cursor-pointer hidden md:flex px-1.5 py-0.5 rounded hover:bg-background/30 dark:hover:bg-background/20">
              <ClockIcon className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
              <span className="text-muted-foreground text-xs">{stats.uptime}</span>
            </div>
          </Tooltip>
        )}

        {/* Notifications */}
        {notificationCount > 0 && (
          <Tooltip content={`${notificationCount} unread notifications`} variant="info">
            <div className="flex items-center gap-1.5 cursor-pointer px-1.5 py-0.5 rounded hover:bg-background/30 dark:hover:bg-background/20">
              <div className="relative flex-shrink-0">
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
          <div className="flex items-center gap-1.5 cursor-pointer px-1.5 py-0.5 rounded hover:bg-background/30 dark:hover:bg-background/20">
            <UserCircleIcon className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
            <span className="text-muted-foreground text-xs hidden lg:inline truncate max-w-[100px]">
              {userName}
            </span>
          </div>
        </Tooltip>

        <div className="h-4 w-px bg-border/50" />

        {/* Version */}
        <Tooltip content={`Application version ${version}`} variant="info">
          <div className="flex items-center gap-1.5 cursor-pointer px-1.5 py-0.5 rounded hover:bg-background/30 dark:hover:bg-background/20">
            <TagIcon className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
            <span className="text-muted-foreground text-xs font-mono">v{version}</span>
          </div>
        </Tooltip>
      </div>
    </footer>
  )
}

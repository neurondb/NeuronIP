'use client'

import {
  ServerIcon,
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  Cog6ToothIcon,
  PlayIcon,
  PauseIcon,
  CalendarDaysIcon,
  KeyIcon,
} from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { slideUp } from '@/lib/animations/variants'
import apiClient from '@/lib/api/client'

interface Connector {
  id: string
  name: string
  connector_type?: string
  type?: string
  sync_status?: string
  status?: 'idle' | 'syncing' | 'error' | 'paused'
  lastSyncAt?: string
  last_sync_at?: string
  nextSyncAt?: string
  enabled: boolean
  schedule?: string
  metadata?: Record<string, unknown>
  [key: string]: unknown
}

interface ConnectorListProps {
  onEdit?: (connector: Connector) => void
  onSchedule?: (connector: Connector) => void
  onCredentials?: (connector: Connector) => void
}

export default function ConnectorList({ onEdit, onSchedule, onCredentials }: ConnectorListProps) {
  const [connectors, setConnectors] = useState<Connector[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    apiClient
      .get<Connector[]>('/connectors')
      .then(({ data }) => {
        setConnectors(Array.isArray(data) ? data : [])
        setLoading(false)
      })
      .catch((err) => {
        console.error('Failed to fetch connectors:', err)
        setConnectors([])
        setLoading(false)
      })
  }, [])

  const normStatus = (c: Connector): string =>
    (c.sync_status || c.status || 'idle') as string
  const normType = (c: Connector): string =>
    (c.connector_type || c.type || '') as string
  const normLastSync = (c: Connector): string | undefined =>
    (c.last_sync_at as string) || c.lastSyncAt

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'syncing': return 'text-blue-500'
      case 'error': return 'text-red-500'
      case 'paused': return 'text-yellow-500'
      default: return 'text-green-500'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'syncing': return <ClockIcon className="h-5 w-5 animate-spin" />
      case 'error': return <XCircleIcon className="h-5 w-5" />
      case 'paused': return <PauseIcon className="h-5 w-5" />
      default: return <CheckCircleIcon className="h-5 w-5" />
    }
  }

  if (loading) {
    return <div className="text-muted-foreground">Loading connectors...</div>
  }

  if (connectors.length === 0) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <ServerIcon className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground mb-4">No connectors configured</p>
          <Button>Add Connector</Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {connectors.map((connector, index) => {
        const status = normStatus(connector)
        const lastSync = normLastSync(connector)
        const schedule = connector.schedule ?? (connector.metadata as Record<string, unknown> | undefined)?.schedule as string | undefined
        return (
          <motion.div key={connector.id} variants={slideUp} transition={{ delay: index * 0.05 }}>
            <Card className="hover:shadow-lg transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle className="text-lg">{connector.name}</CardTitle>
                    <CardDescription className="mt-1">{normType(connector) || 'Connector'}</CardDescription>
                  </div>
                  <div className={`flex items-center gap-2 ${getStatusColor(status)}`}>
                    {getStatusIcon(status)}
                    <span className="text-sm capitalize">{status}</span>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-2 text-sm">
                  {lastSync && (
                    <div className="flex items-center gap-2 text-muted-foreground">
                      <ClockIcon className="h-4 w-4" />
                      <span>Last sync: {new Date(lastSync).toLocaleString()}</span>
                    </div>
                  )}
                  {schedule && (
                    <div className="text-muted-foreground">
                      Schedule: {String(schedule)}
                    </div>
                  )}
                  <div className="flex flex-wrap gap-2 pt-2">
                    {onEdit && (
                      <Button variant="outline" size="sm" onClick={() => onEdit(connector)}>
                        <Cog6ToothIcon className="h-4 w-4 mr-1" />
                        Configure
                      </Button>
                    )}
                    {onSchedule && (
                      <Button variant="outline" size="sm" onClick={() => onSchedule(connector)}>
                        <CalendarDaysIcon className="h-4 w-4 mr-1" />
                        Schedule
                      </Button>
                    )}
                    {onCredentials && (
                      <Button variant="outline" size="sm" onClick={() => onCredentials(connector)}>
                        <KeyIcon className="h-4 w-4 mr-1" />
                        Credentials
                      </Button>
                    )}
                    <Button variant="outline" size="sm">
                      <PlayIcon className="h-4 w-4 mr-1" />
                      Sync Now
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )
      })}
    </div>
  )
}

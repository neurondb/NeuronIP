'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { 
  ServerIcon, 
  CheckCircleIcon, 
  XCircleIcon, 
  ClockIcon,
  Cog6ToothIcon,
  PlayIcon,
  PauseIcon
} from '@heroicons/react/24/outline'
import { slideUp } from '@/lib/animations/variants'

interface Connector {
  id: string
  name: string
  type: string
  status: 'idle' | 'syncing' | 'error' | 'paused'
  lastSyncAt?: string
  nextSyncAt?: string
  enabled: boolean
  schedule?: string
}

export default function ConnectorList() {
  const [connectors, setConnectors] = useState<Connector[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Fetch connectors from API
    fetch('/api/v1/data-sources')
      .then(res => res.json())
      .then(data => {
        setConnectors(data)
        setLoading(false)
      })
      .catch(err => {
        console.error('Failed to fetch connectors:', err)
        setLoading(false)
      })
  }, [])

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
      {connectors.map((connector, index) => (
        <motion.div key={connector.id} variants={slideUp} transition={{ delay: index * 0.05 }}>
          <Card className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
                  <CardTitle className="text-lg">{connector.name}</CardTitle>
                  <CardDescription className="mt-1">{connector.type}</CardDescription>
                </div>
                <div className={`flex items-center gap-2 ${getStatusColor(connector.status)}`}>
                  {getStatusIcon(connector.status)}
                  <span className="text-sm capitalize">{connector.status}</span>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm">
                {connector.lastSyncAt && (
                  <div className="flex items-center gap-2 text-muted-foreground">
                    <ClockIcon className="h-4 w-4" />
                    <span>Last sync: {new Date(connector.lastSyncAt).toLocaleString()}</span>
                  </div>
                )}
                {connector.schedule && (
                  <div className="text-muted-foreground">
                    Schedule: {connector.schedule}
                  </div>
                )}
                <div className="flex gap-2 pt-2">
                  <Button variant="outline" size="sm">
                    <Cog6ToothIcon className="h-4 w-4 mr-1" />
                    Configure
                  </Button>
                  <Button variant="outline" size="sm">
                    <PlayIcon className="h-4 w-4 mr-1" />
                    Sync Now
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      ))}
    </div>
  )
}

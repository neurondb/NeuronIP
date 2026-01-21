'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'

export default function LogViewer() {
  const [logType, setLogType] = useState('')
  const [level, setLevel] = useState('')
  const [limit, setLimit] = useState(100)

  const { data: logs, isLoading, refetch } = useQuery({
    queryKey: ['observability-logs', logType, level, limit],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.observabilityLogs, {
        params: {
          log_type: logType || undefined,
          level: level || undefined,
          limit,
        },
      })
      return response.data
    },
    refetchInterval: 10000, // Refresh every 10 seconds
  })

  const getLevelColor = (level: string) => {
    switch (level.toLowerCase()) {
      case 'error':
      case 'critical':
        return 'text-red-600'
      case 'warning':
        return 'text-yellow-600'
      case 'info':
        return 'text-blue-600'
      default:
        return 'text-muted-foreground'
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>System Logs</CardTitle>
        <CardDescription>View and filter system logs</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex gap-2 flex-wrap">
          <select
            value={logType}
            onChange={(e) => setLogType(e.target.value)}
            className="rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="">All Types</option>
            <option value="query">Query</option>
            <option value="agent">Agent</option>
            <option value="workflow">Workflow</option>
            <option value="system">System</option>
            <option value="error">Error</option>
          </select>

          <select
            value={level}
            onChange={(e) => setLevel(e.target.value)}
            className="rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
          >
            <option value="">All Levels</option>
            <option value="debug">Debug</option>
            <option value="info">Info</option>
            <option value="warning">Warning</option>
            <option value="error">Error</option>
            <option value="critical">Critical</option>
          </select>

          <input
            type="number"
            value={limit}
            onChange={(e) => setLimit(parseInt(e.target.value) || 100)}
            placeholder="Limit"
            className="w-24 rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
          />

          <Button onClick={() => refetch()} size="sm">
            Refresh
          </Button>
        </div>

        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading logs...</div>
        ) : !logs || (Array.isArray(logs) && logs.length === 0) ? (
          <div className="text-center py-8 text-muted-foreground">No logs found</div>
        ) : (
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {(Array.isArray(logs) ? logs : []).map((log: any) => (
              <div
                key={log.id}
                className="p-3 rounded-lg border border-border bg-background text-sm font-mono"
              >
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs text-muted-foreground">
                    {new Date(log.timestamp).toLocaleString()}
                  </span>
                  <span className={`text-xs font-semibold ${getLevelColor(log.level)}`}>
                    {log.level.toUpperCase()}
                  </span>
                  <span className="text-xs text-muted-foreground">{log.log_type}</span>
                </div>
                <div className="text-foreground">{log.message}</div>
                {log.context && Object.keys(log.context).length > 0 && (
                  <details className="mt-2">
                    <summary className="text-xs text-muted-foreground cursor-pointer">Context</summary>
                    <pre className="mt-1 text-xs bg-muted p-2 rounded overflow-x-auto">
                      {JSON.stringify(log.context, null, 2)}
                    </pre>
                  </details>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

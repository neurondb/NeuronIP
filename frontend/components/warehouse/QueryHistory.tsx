'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'
import Button from '@/components/ui/Button'
import { ClockIcon } from '@heroicons/react/24/outline'

interface QueryHistoryProps {
  onSelectQuery?: (query: string) => void
}

export default function QueryHistory({ onSelectQuery }: QueryHistoryProps) {
  const { data: history, isLoading } = useQuery({
    queryKey: ['warehouse-query-history'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.warehouseQueryHistory)
      return response.data
    },
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>Query History</CardTitle>
        <CardDescription>Recently executed queries</CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading history...</div>
        ) : !history || (Array.isArray(history) && history.length === 0) ? (
          <div className="text-center py-8 text-muted-foreground">No query history</div>
        ) : (
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {(Array.isArray(history) ? history : []).map((item: any) => (
              <div
                key={item.query_id}
                className="p-3 rounded-lg border border-border hover:bg-accent transition-colors cursor-pointer"
                onClick={() => onSelectQuery?.(item.natural_language_query || item.generated_sql)}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="font-medium text-sm">{item.natural_language_query || 'SQL Query'}</div>
                    {item.generated_sql && (
                      <pre className="text-xs text-muted-foreground mt-1 overflow-x-auto">
                        {item.generated_sql.substring(0, 100)}
                        {item.generated_sql.length > 100 ? '...' : ''}
                      </pre>
                    )}
                    <div className="flex items-center gap-2 mt-2 text-xs text-muted-foreground">
                      <ClockIcon className="h-3 w-3" />
                      {new Date(item.created_at).toLocaleString()}
                      {item.execution_time_ms && (
                        <span className="ml-2">({item.execution_time_ms}ms)</span>
                      )}
                    </div>
                  </div>
                  <span
                    className={`text-xs px-2 py-1 rounded ${
                      item.status === 'completed'
                        ? 'bg-green-500/10 text-green-600'
                        : item.status === 'failed'
                        ? 'bg-red-500/10 text-red-600'
                        : 'bg-yellow-500/10 text-yellow-600'
                    }`}
                  >
                    {item.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

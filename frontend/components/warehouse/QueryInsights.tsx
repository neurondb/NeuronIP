'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { useMutation } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'
import Button from '@/components/ui/Button'
import { LightBulbIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline'

interface QueryInsightsProps {
  sql: string
}

export default function QueryInsights({ sql }: QueryInsightsProps) {
  const { data: suggestions, mutate: getSuggestions, isPending } = useMutation({
    mutationFn: async (sqlQuery: string) => {
      const response = await apiClient.post(API_ENDPOINTS.warehouseOptimize, { sql: sqlQuery })
      return response.data
    },
  })

  const handleAnalyze = () => {
    if (sql.trim()) {
      getSuggestions(sql)
    }
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'high':
        return 'text-red-600 bg-red-500/10'
      case 'medium':
        return 'text-yellow-600 bg-yellow-500/10'
      case 'low':
        return 'text-blue-600 bg-blue-500/10'
      default:
        return 'text-gray-600 bg-gray-500/10'
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Query Insights</CardTitle>
            <CardDescription>Optimization suggestions and performance insights</CardDescription>
          </div>
          <Button onClick={handleAnalyze} disabled={isPending || !sql.trim()} size="sm">
            Analyze
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        {isPending ? (
          <div className="text-center py-8 text-muted-foreground">Analyzing query...</div>
        ) : !suggestions || (suggestions.suggestions && suggestions.suggestions.length === 0) ? (
          <div className="text-center py-8 text-muted-foreground">
            Click "Analyze" to get optimization suggestions
          </div>
        ) : (
          <div className="space-y-3">
            {(suggestions.suggestions || []).map((suggestion: any, index: number) => (
              <div
                key={index}
                className={`p-3 rounded-lg border border-border ${getSeverityColor(suggestion.severity)}`}
              >
                <div className="flex items-start gap-2">
                  {suggestion.severity === 'high' ? (
                    <ExclamationTriangleIcon className="h-5 w-5 flex-shrink-0 mt-0.5" />
                  ) : (
                    <LightBulbIcon className="h-5 w-5 flex-shrink-0 mt-0.5" />
                  )}
                  <div className="flex-1">
                    <div className="font-medium text-sm mb-1">{suggestion.type}</div>
                    <div className="text-sm">{suggestion.description}</div>
                    {suggestion.sql && (
                      <pre className="text-xs mt-2 bg-background/50 p-2 rounded overflow-x-auto">
                        {suggestion.sql}
                      </pre>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

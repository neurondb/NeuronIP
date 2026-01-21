'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useMetrics, useDeleteMetric, useSearchMetrics } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { MagnifyingGlassIcon, PlusIcon, TrashIcon } from '@heroicons/react/24/outline'

interface MetricCatalogProps {
  onSelectMetric?: (metricId: string) => void
  onCreateNew?: () => void
}

export default function MetricCatalog({ onSelectMetric, onCreateNew }: MetricCatalogProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const { data: metrics, isLoading } = useMetrics()
  const deleteMetric = useDeleteMetric()
  const searchMetrics = useSearchMetrics()

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      return
    }
    try {
      const results = await searchMetrics.mutateAsync({ query: searchQuery })
      // Results would be handled by the query cache
    } catch (error: any) {
      showToast('Search failed', 'error')
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this metric?')) {
      return
    }
    try {
      await deleteMetric.mutateAsync(id)
      showToast('Metric deleted successfully', 'success')
    } catch (error: any) {
      showToast('Failed to delete metric', 'error')
    }
  }

  const displayMetrics = metrics || []

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Metric Catalog</CardTitle>
          {onCreateNew && (
            <Button onClick={onCreateNew} size="sm">
              <PlusIcon className="h-4 w-4 mr-1" />
              New Metric
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Search */}
        <div className="flex gap-2">
          <div className="flex-1 relative">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-muted-foreground" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
              placeholder="Search metrics..."
              className="w-full pl-10 rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <Button onClick={handleSearch} disabled={searchMetrics.isPending}>
            Search
          </Button>
        </div>

        {/* Metrics List */}
        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading metrics...</div>
        ) : displayMetrics.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground mb-4">No metrics found. Create one to get started.</p>
            {onCreateNew && (
              <Button onClick={onCreateNew} size="md">
                Create Your First Metric
              </Button>
            )}
          </div>
        ) : (
          <div className="space-y-2">
            {displayMetrics.map((metric: any) => (
              <div
                key={metric.id}
                className="p-4 rounded-lg border border-border hover:bg-accent transition-colors cursor-pointer"
                onClick={() => onSelectMetric?.(metric.id)}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="font-medium text-foreground">{metric.name || metric.display_name}</h3>
                    {metric.business_term && (
                      <p className="text-sm text-muted-foreground mt-1">{metric.business_term}</p>
                    )}
                    {metric.kpi_type && (
                      <span className="inline-block mt-2 px-2 py-1 text-xs rounded bg-primary/10 text-primary">
                        {metric.kpi_type}
                      </span>
                    )}
                    {metric.status && (
                      <span
                        className={`inline-block mt-2 ml-2 px-2 py-1 text-xs rounded ${
                          metric.status === 'approved'
                            ? 'bg-green-500/10 text-green-600'
                            : metric.status === 'draft'
                            ? 'bg-yellow-500/10 text-yellow-600'
                            : 'bg-gray-500/10 text-gray-600'
                        }`}
                      >
                        {metric.status}
                      </span>
                    )}
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDelete(metric.id)
                    }}
                    disabled={deleteMetric.isPending}
                  >
                    <TrashIcon className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

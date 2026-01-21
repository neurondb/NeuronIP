'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useCatalogDatasets, useCatalogSearch } from '@/lib/api/queries'
import { MagnifyingGlassIcon, CubeIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'

interface DatasetListProps {
  onSelectDataset?: (datasetId: string) => void
  onCreateNew?: () => void
}

export default function DatasetList({ onSelectDataset, onCreateNew }: DatasetListProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const { data: datasetsData, isLoading } = useCatalogDatasets()
  const searchMutation = useCatalogSearch()

  const datasets = datasetsData?.datasets || []
  const displayDatasets = searchQuery ? [] : datasets

  const handleSearch = async () => {
    if (!searchQuery.trim()) return
    await searchMutation.mutateAsync(searchQuery)
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Loading datasets...</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Search Catalog</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <div className="flex-1 relative">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search datasets, fields, owners..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                className="w-full pl-10 pr-4 py-2 rounded-lg border border-border bg-background focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <Button onClick={handleSearch} disabled={!searchQuery.trim() || searchMutation.isPending}>
              Search
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Datasets ({datasets.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {displayDatasets.length === 0 ? (
            <div className="text-center py-12">
              <CubeIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4 opacity-50" />
              <p className="text-muted-foreground mb-4">No datasets found. Create one to get started.</p>
              {onCreateNew && (
                <Button onClick={onCreateNew} size="md">
                  Add Dataset
                </Button>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {displayDatasets.map((dataset: any) => (
                <motion.div
                  key={dataset.id}
                  whileHover={{ y: -2 }}
                  className="border border-border rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer"
                  onClick={() => onSelectDataset?.(dataset.id)}
                >
                  <div className="flex items-start gap-3">
                    <CubeIcon className="h-6 w-6 text-blue-600 flex-shrink-0 mt-1" />
                    <div className="flex-1 min-w-0">
                      <h3 className="font-semibold text-sm truncate">{dataset.name || dataset.id}</h3>
                      {dataset.description && (
                        <p className="text-xs text-muted-foreground mt-1 line-clamp-2">{dataset.description}</p>
                      )}
                      <div className="flex flex-wrap gap-2 mt-2">
                        {dataset.owner && (
                          <span className="text-xs px-2 py-0.5 bg-muted rounded">Owner: {dataset.owner}</span>
                        )}
                        {dataset.tags && dataset.tags.length > 0 && (
                          <span className="text-xs px-2 py-0.5 bg-muted rounded">
                            {dataset.tags.length} tag{dataset.tags.length !== 1 ? 's' : ''}
                          </span>
                        )}
                      </div>
                      {dataset.fields && (
                        <p className="text-xs text-muted-foreground mt-2">
                          {dataset.fields.length} field{dataset.fields.length !== 1 ? 's' : ''}
                        </p>
                      )}
                    </div>
                  </div>
                </motion.div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { useCatalogDataset } from '@/lib/api/queries'

interface DatasetDetailProps {
  datasetId: string
}

export default function DatasetDetail({ datasetId }: DatasetDetailProps) {
  const { data: dataset, isLoading } = useCatalogDataset(datasetId)

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Loading dataset details...</p>
        </CardContent>
      </Card>
    )
  }

  if (!dataset) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Dataset not found</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{dataset.name || dataset.id}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {dataset.description && (
            <div>
              <h4 className="text-sm font-semibold mb-2">Description</h4>
              <p className="text-sm text-muted-foreground">{dataset.description}</p>
            </div>
          )}

          {dataset.owner && (
            <div>
              <h4 className="text-sm font-semibold mb-2">Owner</h4>
              <p className="text-sm text-muted-foreground">{dataset.owner}</p>
            </div>
          )}

          {dataset.tags && dataset.tags.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold mb-2">Tags</h4>
              <div className="flex flex-wrap gap-2">
                {dataset.tags.map((tag: string, index: number) => (
                  <span key={index} className="text-xs px-2 py-1 bg-muted rounded">
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}

          {dataset.fields && dataset.fields.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold mb-2">Fields ({dataset.fields.length})</h4>
              <div className="space-y-2">
                {dataset.fields.map((field: any, index: number) => (
                  <div key={field.id || index} className="p-2 border border-border rounded-lg text-sm">
                    <div className="flex items-center justify-between">
                      <span className="font-medium">{field.field_name || field.name}</span>
                      <span className="text-xs text-muted-foreground">{field.field_type || field.type}</span>
                    </div>
                    {field.description && (
                      <p className="text-xs text-muted-foreground mt-1">{field.description}</p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {dataset.schema_info && (
            <div>
              <h4 className="text-sm font-semibold mb-2">Schema</h4>
              <pre className="text-xs bg-muted p-3 rounded-lg overflow-x-auto">
                {JSON.stringify(dataset.schema_info, null, 2)}
              </pre>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
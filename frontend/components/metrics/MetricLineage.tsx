'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { useMetricLineage, useAddMetricLineage } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { useState } from 'react'
import Button from '@/components/ui/Button'

interface MetricLineageProps {
  metricId: string
}

export default function MetricLineage({ metricId }: MetricLineageProps) {
  const { data: lineage, isLoading } = useMetricLineage(metricId)
  const addLineage = useAddMetricLineage()
  const [showAddForm, setShowAddForm] = useState(false)
  const [relationshipType, setRelationshipType] = useState('uses')
  const [dependsOnMetricId, setDependsOnMetricId] = useState('')
  const [dependsOnTable, setDependsOnTable] = useState('')
  const [dependsOnColumn, setDependsOnColumn] = useState('')

  const handleAddLineage = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!dependsOnMetricId && !dependsOnTable) {
      showToast('Either metric ID or table must be provided', 'warning')
      return
    }

    try {
      await addLineage.mutateAsync({
        id: metricId,
        data: {
          depends_on_metric_id: dependsOnMetricId || null,
          depends_on_table: dependsOnTable || null,
          depends_on_column: dependsOnColumn || null,
          relationship_type: relationshipType,
        },
      })
      showToast('Lineage relationship added', 'success')
      setShowAddForm(false)
      setDependsOnMetricId('')
      setDependsOnTable('')
      setDependsOnColumn('')
    } catch (error: any) {
      showToast('Failed to add lineage', 'error')
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Metric Lineage</CardTitle>
            <CardDescription>Dependencies and relationships for this metric</CardDescription>
          </div>
          <Button onClick={() => setShowAddForm(!showAddForm)} size="sm">
            Add Relationship
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {showAddForm && (
          <form onSubmit={handleAddLineage} className="p-4 rounded-lg border border-border space-y-3">
            <div>
              <label className="text-sm font-medium mb-2 block">Relationship Type</label>
              <select
                value={relationshipType}
                onChange={(e) => setRelationshipType(e.target.value)}
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="uses">Uses</option>
                <option value="derived_from">Derived From</option>
                <option value="aggregates">Aggregates</option>
                <option value="filters">Filters</option>
              </select>
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block">Depends On Metric ID (optional)</label>
              <input
                type="text"
                value={dependsOnMetricId}
                onChange={(e) => setDependsOnMetricId(e.target.value)}
                placeholder="UUID of dependent metric"
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium mb-2 block">Table (optional)</label>
                <input
                  type="text"
                  value={dependsOnTable}
                  onChange={(e) => setDependsOnTable(e.target.value)}
                  placeholder="table_name"
                  className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </div>

              <div>
                <label className="text-sm font-medium mb-2 block">Column (optional)</label>
                <input
                  type="text"
                  value={dependsOnColumn}
                  onChange={(e) => setDependsOnColumn(e.target.value)}
                  placeholder="column_name"
                  className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </div>
            </div>

            <div className="flex gap-2 justify-end">
              <Button type="button" variant="outline" onClick={() => setShowAddForm(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={addLineage.isPending}>
                Add Relationship
              </Button>
            </div>
          </form>
        )}

        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading lineage...</div>
        ) : !lineage || (Array.isArray(lineage) && lineage.length === 0) ? (
          <div className="text-center py-8 text-muted-foreground">
            No lineage relationships found. Add one to get started.
          </div>
        ) : (
          <div className="space-y-2">
            {(Array.isArray(lineage) ? lineage : []).map((rel: any) => (
              <div key={rel.id} className="p-3 rounded-lg border border-border">
                <div className="flex items-center justify-between">
                  <div>
                    <span className="font-medium text-foreground">{rel.relationship_type}</span>
                    {rel.depends_on_metric_id && (
                      <span className="ml-2 text-sm text-muted-foreground">
                        Metric: {rel.depends_on_metric_id}
                      </span>
                    )}
                    {rel.depends_on_table && (
                      <span className="ml-2 text-sm text-muted-foreground">
                        Table: {rel.depends_on_table}
                        {rel.depends_on_column && `.${rel.depends_on_column}`}
                      </span>
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

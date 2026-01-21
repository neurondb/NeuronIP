'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import MetricCatalog from '@/components/metrics/MetricCatalog'
import MetricEditor from '@/components/metrics/MetricEditor'
import MetricLineage from '@/components/metrics/MetricLineage'
import { useMetric, useCalculateMetric } from '@/lib/api/queries'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { showToast } from '@/components/ui/Toast'

export default function MetricsPage() {
  const [view, setView] = useState<'catalog' | 'editor' | 'lineage'>('catalog')
  const [selectedMetricId, setSelectedMetricId] = useState<string | null>(null)
  const [showEditor, setShowEditor] = useState(false)

  const { data: selectedMetric } = useMetric(selectedMetricId || '', !!selectedMetricId)
  const calculateMetric = useCalculateMetric()

  const handleSelectMetric = (metricId: string) => {
    setSelectedMetricId(metricId)
    setView('lineage')
  }

  const handleCreateNew = () => {
    setSelectedMetricId(null)
    setShowEditor(true)
    setView('editor')
  }

  const handleEditorSuccess = () => {
    setShowEditor(false)
    setView('catalog')
  }

  const handleCalculate = async () => {
    if (!selectedMetricId) {
      showToast('Please select a metric first', 'warning')
      return
    }

    try {
      const result = await calculateMetric.mutateAsync({ id: selectedMetricId })
      showToast(`Metric value: ${result.value}`, 'success')
    } catch (error: any) {
      showToast('Failed to calculate metric', 'error')
    }
  }

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Metrics / Semantic Layer</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Define KPIs, dimensions, business terms. Reusable metrics for analytics and agents.
            </p>
          </div>
          <div className="flex gap-2">
            <Button onClick={handleCreateNew} variant="outline">
              New Metric
            </Button>
            {selectedMetricId && (
              <Button onClick={handleCalculate} disabled={calculateMetric.isPending}>
                Calculate
              </Button>
            )}
          </div>
        </div>
      </motion.div>

      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        <div className="flex gap-2 mb-4">
          <Button
            variant={view === 'catalog' ? 'primary' : 'outline'}
            onClick={() => setView('catalog')}
            size="sm"
          >
            Catalog
          </Button>
          {selectedMetricId && (
            <>
              <Button
                variant={view === 'lineage' ? 'primary' : 'outline'}
                onClick={() => setView('lineage')}
                size="sm"
              >
                Lineage
              </Button>
              <Button
                variant={view === 'editor' ? 'primary' : 'outline'}
                onClick={() => {
                  setView('editor')
                  setShowEditor(true)
                }}
                size="sm"
              >
                Edit
              </Button>
            </>
          )}
        </div>

        {view === 'catalog' && (
          <MetricCatalog onSelectMetric={handleSelectMetric} onCreateNew={handleCreateNew} />
        )}

        {view === 'editor' && showEditor && (
          <MetricEditor
            metricId={selectedMetricId || undefined}
            onSuccess={handleEditorSuccess}
            onCancel={() => {
              setShowEditor(false)
              setView('catalog')
            }}
          />
        )}

        {view === 'lineage' && selectedMetricId && (
          <div className="space-y-4">
            {selectedMetric && (
              <Card>
                <CardHeader>
                  <CardTitle>{selectedMetric.name || selectedMetric.display_name}</CardTitle>
                </CardHeader>
                <CardContent>
                  {selectedMetric.definition && (
                    <div>
                      <p className="text-sm font-medium mb-1">Definition:</p>
                      <pre className="text-xs bg-muted p-3 rounded-lg overflow-x-auto">
                        {selectedMetric.definition || selectedMetric.sql_expression}
                      </pre>
                    </div>
                  )}
                </CardContent>
              </Card>
            )}
            <MetricLineage metricId={selectedMetricId} />
          </div>
        )}
      </motion.div>
    </motion.div>
  )
}

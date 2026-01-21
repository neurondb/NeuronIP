'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useCreateMetric, useUpdateMetric, useMetric } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface MetricEditorProps {
  metricId?: string
  onSuccess?: () => void
  onCancel?: () => void
}

export default function MetricEditor({ metricId, onSuccess, onCancel }: MetricEditorProps) {
  const [name, setName] = useState('')
  const [definition, setDefinition] = useState('')
  const [kpiType, setKpiType] = useState('')
  const [businessTerm, setBusinessTerm] = useState('')
  const [reusable, setReusable] = useState(true)

  const { data: existingMetric } = useMetric(metricId || '', !!metricId)
  const createMetric = useCreateMetric()
  const updateMetric = useUpdateMetric()

  // Populate form if editing
  if (metricId && existingMetric && !name) {
    setName(existingMetric.name || '')
    setDefinition(existingMetric.definition || existingMetric.sql_expression || '')
    setKpiType(existingMetric.kpi_type || '')
    setBusinessTerm(existingMetric.business_term || '')
    setReusable(existingMetric.reusable !== false)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name || !definition) {
      showToast('Name and definition are required', 'warning')
      return
    }

    const metricData = {
      name,
      definition,
      kpi_type: kpiType || null,
      business_term: businessTerm || null,
      reusable,
    }

    try {
      if (metricId) {
        await updateMetric.mutateAsync({ id: metricId, data: metricData })
        showToast('Metric updated successfully', 'success')
      } else {
        await createMetric.mutateAsync(metricData)
        showToast('Metric created successfully', 'success')
      }
      onSuccess?.()
    } catch (error: any) {
      showToast(error?.response?.data?.message || 'Failed to save metric', 'error')
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{metricId ? 'Edit Metric' : 'Create Metric'}</CardTitle>
        <CardDescription>
          Define a metric with SQL expression, KPI type, and business terms
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-2 block">Name *</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Total Revenue"
              className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-2 block">SQL Definition *</label>
            <textarea
              value={definition}
              onChange={(e) => setDefinition(e.target.value)}
              placeholder="SELECT SUM(amount) FROM transactions WHERE status = 'completed'"
              rows={6}
              className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring resize-none font-mono text-sm"
              required
            />
            <p className="text-xs text-muted-foreground mt-1">
              Enter the SQL expression that calculates this metric
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium mb-2 block">KPI Type</label>
              <select
                value={kpiType}
                onChange={(e) => setKpiType(e.target.value)}
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="">Select type...</option>
                <option value="revenue">Revenue</option>
                <option value="growth">Growth</option>
                <option value="efficiency">Efficiency</option>
                <option value="quality">Quality</option>
                <option value="other">Other</option>
              </select>
            </div>

            <div>
              <label className="text-sm font-medium mb-2 block">Business Term</label>
              <input
                type="text"
                value={businessTerm}
                onChange={(e) => setBusinessTerm(e.target.value)}
                placeholder="e.g., Monthly Recurring Revenue"
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="reusable"
              checked={reusable}
              onChange={(e) => setReusable(e.target.checked)}
              className="rounded border-border"
            />
            <label htmlFor="reusable" className="text-sm font-medium">
              Reusable metric (can be used by other metrics)
            </label>
          </div>

          <div className="flex gap-2 justify-end">
            {onCancel && (
              <Button type="button" variant="outline" onClick={onCancel}>
                Cancel
              </Button>
            )}
            <Button type="submit" disabled={createMetric.isPending || updateMetric.isPending}>
              {metricId ? 'Update Metric' : 'Create Metric'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

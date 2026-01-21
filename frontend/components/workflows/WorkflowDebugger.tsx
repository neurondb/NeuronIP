'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useWorkflowExecutionLogs, useWorkflowExecutionMetrics } from '@/lib/api/queries'

interface WorkflowDebuggerProps {
  executionId: string
}

export default function WorkflowDebugger({ executionId }: WorkflowDebuggerProps) {
  const [selectedStep, setSelectedStep] = useState<string | null>(null)
  const { data: logs } = useWorkflowExecutionLogs(executionId, true)
  const { data: metrics } = useWorkflowExecutionMetrics(executionId, true)

  const stepLogs = selectedStep
    ? logs?.logs?.filter((log: any) => log.step === selectedStep) || []
    : logs?.logs || []

  const stepMetrics = selectedStep
    ? metrics?.metrics?.filter((metric: any) => metric.step === selectedStep) || []
    : metrics?.metrics || []

  const steps: string[] = Array.from(new Set(logs?.logs?.map((log: any) => log.step) || [])) as string[]

  const handleExportLogs = () => {
    if (!logs?.logs || logs.logs.length === 0) {
      return
    }

    const logData = JSON.stringify(logs.logs, null, 2)
    const blob = new Blob([logData], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `workflow-execution-${executionId}-logs-${new Date().toISOString().split('T')[0]}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const handleExportMetrics = () => {
    if (!metrics?.metrics || metrics.metrics.length === 0) {
      return
    }

    // Export as CSV
    const headers = ['Name', 'Value', 'Unit', 'Description', 'Timestamp']
    const csvRows = [
      headers.join(','),
      ...metrics.metrics.map((metric: any) =>
        [
          `"${metric.name || ''}"`,
          metric.value || '',
          `"${metric.unit || ''}"`,
          `"${(metric.description || '').replace(/"/g, '""')}"`,
          metric.timestamp || new Date().toISOString(),
        ].join(',')
      ),
    ]
    const csvData = csvRows.join('\n')
    const blob = new Blob([csvData], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `workflow-execution-${executionId}-metrics-${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const handleGenerateReport = () => {
    const reportData = {
      executionId,
      generatedAt: new Date().toISOString(),
      summary: {
        totalLogs: logs?.logs?.length || 0,
        totalMetrics: metrics?.metrics?.length || 0,
        steps: steps.length,
      },
      logs: logs?.logs || [],
      metrics: metrics?.metrics || [],
    }

    const reportContent = `# Workflow Execution Report
Execution ID: ${executionId}
Generated: ${new Date().toLocaleString()}

## Summary
- Total Logs: ${reportData.summary.totalLogs}
- Total Metrics: ${reportData.summary.totalMetrics}
- Steps: ${reportData.summary.steps}

## Logs
${(logs?.logs || [])
  .map(
    (log: any) =>
      `[${new Date(log.timestamp).toLocaleString()}] [${log.level?.toUpperCase() || 'INFO'}] ${log.message}`
  )
  .join('\n')}

## Metrics
${(metrics?.metrics || [])
  .map((metric: any) => `- ${metric.name}: ${metric.value}${metric.unit ? ' ' + metric.unit : ''}`)
  .join('\n')}

## Raw Data
\`\`\`json
${JSON.stringify(reportData, null, 2)}
\`\`\`
`

    const blob = new Blob([reportContent], { type: 'text/markdown' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `workflow-execution-${executionId}-report-${new Date().toISOString().split('T')[0]}.md`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Workflow Debugger</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Step Filter */}
        <div className="mb-4">
          <label className="text-sm font-medium mb-2 block">Filter by Step</label>
          <div className="flex gap-2 flex-wrap">
            <button
              onClick={() => setSelectedStep(null)}
              className={`px-3 py-1 rounded text-sm ${
                selectedStep === null
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              All Steps
            </button>
            {steps.map((step: string) => (
              <button
                key={step}
                onClick={() => setSelectedStep(step)}
                className={`px-3 py-1 rounded text-sm ${
                  selectedStep === step
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {step}
              </button>
            ))}
          </div>
        </div>

        {/* Debug Info */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Logs */}
          <div>
            <h3 className="font-semibold mb-2">Execution Logs</h3>
            <div className="bg-gray-900 text-green-400 p-4 rounded font-mono text-xs max-h-[400px] overflow-y-auto">
              {stepLogs.length > 0 ? (
                stepLogs.map((log: any, idx: number) => (
                  <div key={idx} className="mb-1">
                    <span className="text-gray-500">
                      [{new Date(log.timestamp).toLocaleTimeString()}]
                    </span>{' '}
                    <span className={log.level === 'error' ? 'text-red-400' : ''}>{log.message}</span>
                  </div>
                ))
              ) : (
                <div className="text-gray-500">No logs available</div>
              )}
            </div>
          </div>

          {/* Metrics */}
          <div>
            <h3 className="font-semibold mb-2">Performance Metrics</h3>
            <div className="space-y-2">
              {stepMetrics.length > 0 ? (
                stepMetrics.map((metric: any, idx: number) => (
                  <div key={idx} className="p-3 bg-gray-50 rounded">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">{metric.name}</span>
                      <span className="text-lg font-bold">{metric.value}</span>
                    </div>
                    {metric.unit && (
                      <div className="text-xs text-muted-foreground mt-1">{metric.unit}</div>
                    )}
                    {metric.description && (
                      <div className="text-xs text-muted-foreground mt-1">{metric.description}</div>
                    )}
                  </div>
                ))
              ) : (
                <div className="text-sm text-muted-foreground">No metrics available</div>
              )}
            </div>
          </div>
        </div>

        {/* Debug Actions */}
        <div className="mt-4 flex gap-2">
          <Button size="sm" variant="outline" onClick={handleExportLogs}>
            Export Logs
          </Button>
          <Button size="sm" variant="outline" onClick={handleExportMetrics}>
            Export Metrics
          </Button>
          <Button size="sm" variant="outline" onClick={handleGenerateReport}>
            Generate Report
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

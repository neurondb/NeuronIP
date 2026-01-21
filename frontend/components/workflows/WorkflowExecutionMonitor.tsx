'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import {
  useWorkflowExecutionStatus,
  useWorkflowExecutionLogs,
  useWorkflowExecutionMetrics,
  useWorkflowExecutionDecisions,
  useRecoverWorkflowExecution,
} from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface WorkflowExecutionMonitorProps {
  executionId: string
  workflowId: string
}

export default function WorkflowExecutionMonitor({
  executionId,
  workflowId,
}: WorkflowExecutionMonitorProps) {
  const [activeTab, setActiveTab] = useState<'status' | 'logs' | 'metrics' | 'decisions'>('status')
  const { data: status, isLoading: statusLoading } = useWorkflowExecutionStatus(executionId)
  const { data: logs, isLoading: logsLoading } = useWorkflowExecutionLogs(executionId, activeTab === 'logs')
  const { data: metrics, isLoading: metricsLoading } = useWorkflowExecutionMetrics(
    executionId,
    activeTab === 'metrics'
  )
  const { data: decisions, isLoading: decisionsLoading } = useWorkflowExecutionDecisions(
    executionId,
    activeTab === 'decisions'
  )
  const { mutate: recoverExecution, isPending: isRecovering } = useRecoverWorkflowExecution(executionId)

  const handleRecover = () => {
    recoverExecution(undefined, {
      onSuccess: () => {
        showToast('Workflow recovery initiated', 'success')
      },
      onError: () => {
        showToast('Failed to recover workflow', 'error')
      },
    })
  }

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'completed':
        return 'text-green-600 bg-green-50'
      case 'failed':
        return 'text-red-600 bg-red-50'
      case 'running':
        return 'text-blue-600 bg-blue-50'
      default:
        return 'text-gray-600 bg-gray-50'
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Execution Monitor</CardTitle>
          {status?.status === 'failed' && (
            <Button onClick={handleRecover} disabled={isRecovering} size="sm">
              {isRecovering ? 'Recovering...' : 'Recover'}
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {/* Status Overview */}
        <div className="mb-4">
          <div className="flex items-center gap-4 mb-2">
            <span className="text-sm font-medium">Status:</span>
            <span
              className={`px-3 py-1 rounded-full text-sm font-semibold ${getStatusColor(status?.status)}`}
            >
              {status?.status || 'Unknown'}
            </span>
          </div>
          {status?.started_at && (
            <div className="text-sm text-muted-foreground">
              Started: {new Date(status.started_at).toLocaleString()}
            </div>
          )}
          {status?.completed_at && (
            <div className="text-sm text-muted-foreground">
              Completed: {new Date(status.completed_at).toLocaleString()}
            </div>
          )}
          {status?.error_message && (
            <div className="mt-2 p-3 bg-red-50 border border-red-200 rounded text-sm text-red-800">
              {status.error_message}
            </div>
          )}
        </div>

        {/* Tabs */}
        <div className="flex gap-2 border-b border-border mb-4">
          {(['status', 'logs', 'metrics', 'decisions'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              {tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        <div className="min-h-[300px]">
          {activeTab === 'status' && (
            <div>
              {statusLoading ? (
                <div className="text-center py-8 text-muted-foreground">Loading status...</div>
              ) : (
                <div className="space-y-4">
                  <div>
                    <h4 className="font-semibold mb-2">Current Step</h4>
                    <p className="text-sm text-muted-foreground">{status?.current_step || 'N/A'}</p>
                  </div>
                  <div>
                    <h4 className="font-semibold mb-2">Completed Steps</h4>
                    <div className="flex flex-wrap gap-2">
                      {status?.completed_steps && status.completed_steps.length > 0 ? (
                        status.completed_steps.map((step: string) => (
                          <span
                            key={step}
                            className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs"
                          >
                            {step}
                          </span>
                        ))
                      ) : (
                        <span className="text-sm text-muted-foreground">None</span>
                      )}
                    </div>
                  </div>
                  {status?.steps && status.steps.length > 0 && (
                    <div>
                      <h4 className="font-semibold mb-2">Step Details</h4>
                      <div className="space-y-2">
                        {status.steps.map((step: any) => (
                          <div key={step.step_id} className="p-2 bg-gray-50 rounded text-sm">
                            <div className="flex items-center justify-between">
                              <span className="font-medium">{step.step_id}</span>
                              <span
                                className={`px-2 py-1 rounded text-xs ${
                                  step.status === 'completed'
                                    ? 'bg-green-100 text-green-800'
                                    : step.status === 'failed'
                                    ? 'bg-red-100 text-red-800'
                                    : step.status === 'running'
                                    ? 'bg-blue-100 text-blue-800'
                                    : 'bg-gray-100 text-gray-800'
                                }`}
                              >
                                {step.status}
                              </span>
                            </div>
                            {step.error_message && (
                              <div className="mt-1 text-xs text-red-600">{step.error_message}</div>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}

          {activeTab === 'logs' && (
            <div>
              {logsLoading ? (
                <div className="text-center py-8 text-muted-foreground">Loading logs...</div>
              ) : (
                <div className="space-y-2 max-h-[400px] overflow-y-auto">
                  {logs?.logs?.map((log: any, idx: number) => (
                    <div key={idx} className="p-2 bg-gray-50 rounded text-sm font-mono">
                      <div className="text-xs text-muted-foreground mb-1">
                        {new Date(log.timestamp).toLocaleString()}
                      </div>
                      <div>{log.message}</div>
                      {log.level && (
                        <span className={`text-xs px-2 py-0.5 rounded ${
                          log.level === 'error' ? 'bg-red-100 text-red-800' :
                          log.level === 'warn' ? 'bg-yellow-100 text-yellow-800' :
                          'bg-blue-100 text-blue-800'
                        }`}>
                          {log.level}
                        </span>
                      )}
                    </div>
                  )) || <div className="text-sm text-muted-foreground">No logs available</div>}
                </div>
              )}
            </div>
          )}

          {activeTab === 'metrics' && (
            <div>
              {metricsLoading ? (
                <div className="text-center py-8 text-muted-foreground">Loading metrics...</div>
              ) : (
                <div className="grid grid-cols-2 gap-4">
                  {metrics?.metrics?.map((metric: any, idx: number) => (
                    <div key={idx} className="p-3 bg-gray-50 rounded">
                      <div className="text-sm font-semibold">{metric.name}</div>
                      <div className="text-lg font-bold">{metric.value}</div>
                      {metric.unit && <div className="text-xs text-muted-foreground">{metric.unit}</div>}
                    </div>
                  )) || <div className="text-sm text-muted-foreground">No metrics available</div>}
                </div>
              )}
            </div>
          )}

          {activeTab === 'decisions' && (
            <div>
              {decisionsLoading ? (
                <div className="text-center py-8 text-muted-foreground">Loading decisions...</div>
              ) : (
                <div className="space-y-3">
                  {decisions?.decisions?.map((decision: any, idx: number) => (
                    <div key={idx} className="p-3 bg-gray-50 rounded">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-semibold">{decision.step}</span>
                        <span className={`text-xs px-2 py-1 rounded ${
                          decision.outcome === 'success' ? 'bg-green-100 text-green-800' :
                          decision.outcome === 'failure' ? 'bg-red-100 text-red-800' :
                          'bg-yellow-100 text-yellow-800'
                        }`}>
                          {decision.outcome}
                        </span>
                      </div>
                      {decision.reason && (
                        <div className="text-sm text-muted-foreground">{decision.reason}</div>
                      )}
                    </div>
                  )) || <div className="text-sm text-muted-foreground">No decisions available</div>}
                </div>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

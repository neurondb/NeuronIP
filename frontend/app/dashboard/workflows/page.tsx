'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import WorkflowBuilder from '@/components/workflows/WorkflowBuilder'
import WorkflowExecutionMonitor from '@/components/workflows/WorkflowExecutionMonitor'
import WorkflowTemplates from '@/components/workflows/WorkflowTemplates'
import WorkflowDebugger from '@/components/workflows/WorkflowDebugger'
import WorkflowVersionManager from '@/components/workflows/WorkflowVersionManager'
import WorkflowScheduler from '@/components/workflows/WorkflowScheduler'
import ExecutionLog from '@/components/workflows/ExecutionLog'
import {
  useWorkflows,
  useExecuteWorkflow,
  useWorkflow,
  useWorkflowMonitoring,
} from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

interface Execution {
  id: string
  workflowId: string
  status: 'running' | 'completed' | 'failed'
  startedAt: Date
  completedAt?: Date
  duration?: number
}

export default function WorkflowsPage() {
  const [activeTab, setActiveTab] = useState<'builder' | 'execute' | 'monitor' | 'templates' | 'versions' | 'scheduler'>('builder')
  const [selectedWorkflowId, setSelectedWorkflowId] = useState<string>('')
  const [selectedExecutionId, setSelectedExecutionId] = useState<string>('')
  const [workflowId, setWorkflowId] = useState('')
  const [executions, setExecutions] = useState<Execution[]>([])
  const { data: workflows } = useWorkflows()
  const { mutate: executeWorkflow, isPending } = useExecuteWorkflow(workflowId || '')
  const { data: workflow } = useWorkflow(selectedWorkflowId, !!selectedWorkflowId)
  const { data: monitoring } = useWorkflowMonitoring(selectedWorkflowId, '24h', !!selectedWorkflowId)

  const handleExecute = () => {
    if (!workflowId.trim()) {
      showToast('Please enter a workflow ID', 'warning')
      return
    }

    const execution: Execution = {
      id: Date.now().toString(),
      workflowId,
      status: 'running',
      startedAt: new Date(),
    }

    setExecutions((prev) => [execution, ...prev])
    setSelectedExecutionId(execution.id)

    executeWorkflow(
      {},
      {
        onSuccess: (data) => {
          setExecutions((prev) =>
            prev.map((e) =>
              e.id === execution.id
                ? {
                    ...e,
                    status: 'completed',
                    completedAt: new Date(),
                    duration: Date.now() - execution.startedAt.getTime(),
                  }
                : e
            )
          )
          showToast('Workflow executed successfully', 'success')
        },
        onError: () => {
          setExecutions((prev) =>
            prev.map((e) =>
              e.id === execution.id
                ? {
                    ...e,
                    status: 'failed',
                    completedAt: new Date(),
                  }
                : e
            )
          )
          showToast('Workflow execution failed', 'error')
        },
      }
    )
  }

  const tabs = [
    { id: 'builder', label: 'Builder' },
    { id: 'execute', label: 'Execute' },
    { id: 'monitor', label: 'Monitor' },
    { id: 'templates', label: 'Templates' },
    { id: 'versions', label: 'Versions' },
    { id: 'scheduler', label: 'Scheduler' },
  ] as const

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Workflows</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Build, execute, and monitor your automated workflows
        </p>
      </motion.div>

      {/* Workflow Selector */}
      {workflows && (
        <motion.div variants={slideUp} className="flex-shrink-0">
          <Card>
            <CardContent className="pt-4">
              <div className="flex gap-2 items-center">
                <label className="text-sm font-medium">Select Workflow:</label>
                <select
                  value={selectedWorkflowId}
                  onChange={(e) => {
                    setSelectedWorkflowId(e.target.value)
                    if (e.target.value) {
                      setWorkflowId(e.target.value)
                    }
                  }}
                  className="flex-1 rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="">-- Select a workflow --</option>
                  {workflows.workflows?.map((wf: any) => (
                    <option key={wf.id} value={wf.id}>
                      {wf.name || wf.id}
                    </option>
                  ))}
                </select>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      )}

      {/* Tabs */}
      <motion.div variants={slideUp} className="flex-shrink-0">
        <div className="flex gap-2 border-b border-border overflow-x-auto">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors whitespace-nowrap ${
                activeTab === tab.id
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </motion.div>

      {/* Tab Content */}
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-auto">
        {activeTab === 'builder' && (
          <div className="h-full">
            <WorkflowBuilder />
          </div>
        )}

        {activeTab === 'execute' && (
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Execute Workflow</CardTitle>
                <CardDescription>Run a workflow by ID</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2 sm:gap-3">
                  <input
                    type="text"
                    value={workflowId}
                    onChange={(e) => setWorkflowId(e.target.value)}
                    placeholder="Enter workflow ID"
                    className="flex-1 rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                  <Button onClick={handleExecute} disabled={!workflowId.trim() || isPending}>
                    Execute
                  </Button>
                </div>
              </CardContent>
            </Card>

            {executions.length > 0 && (
              <div>
                <ExecutionLog executions={executions} />
              </div>
            )}
          </div>
        )}

        {activeTab === 'monitor' && (
          <div className="space-y-4">
            {selectedExecutionId ? (
              <>
                <WorkflowExecutionMonitor
                  executionId={selectedExecutionId}
                  workflowId={workflowId}
                />
                <WorkflowDebugger executionId={selectedExecutionId} />
              </>
            ) : executions.length > 0 ? (
              <>
                <Card>
                  <CardHeader>
                    <CardTitle>Select Execution</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2">
                      {executions.map((exec) => (
                        <button
                          key={exec.id}
                          onClick={() => setSelectedExecutionId(exec.id)}
                          className="w-full p-3 border border-border rounded-lg hover:bg-gray-50 text-left"
                        >
                          <div className="flex items-center justify-between">
                            <span className="font-medium">Execution {exec.id}</span>
                            <span
                              className={`px-2 py-1 rounded text-xs ${
                                exec.status === 'completed'
                                  ? 'bg-green-100 text-green-800'
                                  : exec.status === 'failed'
                                  ? 'bg-red-100 text-red-800'
                                  : 'bg-blue-100 text-blue-800'
                              }`}
                            >
                              {exec.status}
                            </span>
                          </div>
                          <div className="text-xs text-muted-foreground mt-1">
                            {exec.startedAt.toLocaleString()}
                          </div>
                        </button>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </>
            ) : (
              <Card>
                <CardContent className="py-8 text-center text-muted-foreground">
                  No executions available. Execute a workflow first.
                </CardContent>
              </Card>
            )}

            {selectedWorkflowId && monitoring && (
              <Card>
                <CardHeader>
                  <CardTitle>Workflow Monitoring</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div>
                      <div className="text-sm text-muted-foreground">Total Executions</div>
                      <div className="text-2xl font-bold">{monitoring.total_executions || 0}</div>
                    </div>
                    <div>
                      <div className="text-sm text-muted-foreground">Success Rate</div>
                      <div className="text-2xl font-bold">
                        {monitoring.success_rate ? `${(monitoring.success_rate * 100).toFixed(1)}%` : '0%'}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm text-muted-foreground">Avg Duration</div>
                      <div className="text-2xl font-bold">
                        {monitoring.avg_duration ? `${monitoring.avg_duration.toFixed(2)}s` : '0s'}
                      </div>
                    </div>
                    <div>
                      <div className="text-sm text-muted-foreground">Failed</div>
                      <div className="text-2xl font-bold text-red-600">
                        {monitoring.failed_executions || 0}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        )}

        {activeTab === 'templates' && (
          <div>
            <WorkflowTemplates />
          </div>
        )}

        {activeTab === 'versions' && selectedWorkflowId ? (
          <div>
            <WorkflowVersionManager workflowId={selectedWorkflowId} />
          </div>
        ) : activeTab === 'versions' ? (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              Please select a workflow to view versions
            </CardContent>
          </Card>
        ) : null}

        {activeTab === 'scheduler' && selectedWorkflowId ? (
          <div>
            <WorkflowScheduler workflowId={selectedWorkflowId} />
          </div>
        ) : activeTab === 'scheduler' ? (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              Please select a workflow to schedule
            </CardContent>
          </Card>
        ) : null}
      </motion.div>
    </motion.div>
  )
}

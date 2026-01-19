'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import ExecutionLog from '@/components/workflows/ExecutionLog'
import { useExecuteWorkflow, useWorkflow } from '@/lib/api/queries'
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
  const [workflowId, setWorkflowId] = useState('')
  const [executions, setExecutions] = useState<Execution[]>([])
  const { mutate: executeWorkflow, isPending } = useExecuteWorkflow(workflowId || '')

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
          Execute and monitor your automated workflows
        </p>
      </motion.div>

      {/* Workflow Execution */}
      <motion.div variants={slideUp}>
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
      </motion.div>

      {/* Execution Log */}
      {executions.length > 0 && (
        <motion.div variants={slideUp} className="flex-1 min-h-0">
          <ExecutionLog executions={executions} />
        </motion.div>
      )}
    </motion.div>
  )
}

'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { useCreateWorkflow } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import {
  SparklesIcon,
  DocumentTextIcon,
  LinkIcon,
  ClockIcon,
  CheckCircleIcon,
} from '@heroicons/react/24/outline'

interface WorkflowWizardData {
  name: string
  description: string
  template: string | null
  steps: Array<{
    id: string
    name: string
    type: string
    config: Record<string, any>
  }>
  connections: Array<{
    from: string
    to: string
  }>
  triggers: {
    type: 'manual' | 'schedule' | 'webhook' | 'event'
    schedule?: string
    webhookUrl?: string
    eventType?: string
  }
}

// Step 1: Name and Description
function NameStep({ data, updateData }: WizardStepProps) {
  const workflowData = (data as WorkflowWizardData) || { name: '', description: '' }

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-2">Workflow Name *</label>
        <Input
          value={workflowData.name || ''}
          onChange={(e) => updateData({ name: e.target.value })}
          placeholder="e.g., Customer Support Automation"
          required
        />
      </div>
      <div>
        <label className="block text-sm font-medium mb-2">Description</label>
        <Textarea
          value={workflowData.description || ''}
          onChange={(e) => updateData({ description: e.target.value })}
          placeholder="Describe what this workflow does..."
          rows={4}
        />
      </div>
    </div>
  )
}

// Step 2: Template Selection
function TemplateStep({ data, updateData }: WizardStepProps) {
  const workflowData = (data as WorkflowWizardData) || { template: null }
  const templates = [
    { id: 'customer-support', name: 'Customer Support', description: 'Automated ticket handling' },
    { id: 'data-processing', name: 'Data Processing', description: 'ETL and data transformation' },
    { id: 'notification', name: 'Notification', description: 'Send alerts and notifications' },
    { id: 'custom', name: 'Start from Scratch', description: 'Build your own workflow' },
  ]

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Choose a template to get started quickly, or start from scratch.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {templates.map((template) => (
          <Card
            key={template.id}
            hover
            className={workflowData.template === template.id ? 'ring-2 ring-primary' : ''}
            onClick={() => updateData({ template: template.id })}
          >
            <CardContent className="p-4">
              <h3 className="font-semibold mb-1">{template.name}</h3>
              <p className="text-sm text-muted-foreground">{template.description}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

// Step 3: Add Steps
function StepsStep({ data, updateData }: WizardStepProps) {
  const workflowData = (data as WorkflowWizardData) || { steps: [] }
  const [newStepName, setNewStepName] = useState('')
  const [newStepType, setNewStepType] = useState('agent')

  const stepTypes = [
    { id: 'agent', name: 'AI Agent', description: 'Use an AI agent to process data' },
    { id: 'script', name: 'Script', description: 'Run a custom script' },
    { id: 'condition', name: 'Condition', description: 'Branch based on conditions' },
    { id: 'parallel', name: 'Parallel', description: 'Run multiple steps in parallel' },
  ]

  const addStep = () => {
    if (!newStepName.trim()) return
    const newStep = {
      id: `step-${Date.now()}`,
      name: newStepName,
      type: newStepType,
      config: {},
    }
    updateData({ steps: [...(workflowData.steps || []), newStep] })
    setNewStepName('')
  }

  const removeStep = (stepId: string) => {
    updateData({
      steps: workflowData.steps?.filter((s) => s.id !== stepId) || [],
    })
  }

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        <Input
          value={newStepName}
          onChange={(e) => setNewStepName(e.target.value)}
          placeholder="Step name"
          className="flex-1"
        />
        <select
          value={newStepType}
          onChange={(e) => setNewStepType(e.target.value)}
          className="px-3 py-2 border rounded-lg"
        >
          {stepTypes.map((type) => (
            <option key={type.id} value={type.id}>
              {type.name}
            </option>
          ))}
        </select>
        <Button onClick={addStep}>Add Step</Button>
      </div>
      <div className="space-y-2">
        {workflowData.steps?.map((step) => (
          <Card key={step.id}>
            <CardContent className="p-4 flex items-center justify-between">
              <div>
                <h4 className="font-medium">{step.name}</h4>
                <p className="text-sm text-muted-foreground">{step.type}</p>
              </div>
              <Button variant="ghost" size="sm" onClick={() => removeStep(step.id)}>
                Remove
              </Button>
            </CardContent>
          </Card>
        ))}
        {(!workflowData.steps || workflowData.steps.length === 0) && (
          <p className="text-sm text-muted-foreground text-center py-8">
            No steps added yet. Add your first step above.
          </p>
        )}
      </div>
    </div>
  )
}

// Step 4: Configure Connections
function ConnectionsStep({ data, updateData }: WizardStepProps) {
  const workflowData = (data as WorkflowWizardData) || { steps: [], connections: [] }
  const [fromStep, setFromStep] = useState('')
  const [toStep, setToStep] = useState('')

  const addConnection = () => {
    if (!fromStep || !toStep || fromStep === toStep) return
    const connection = { from: fromStep, to: toStep }
    const existing = workflowData.connections || []
    if (existing.some((c) => c.from === fromStep && c.to === toStep)) return
    updateData({ connections: [...existing, connection] })
    setFromStep('')
    setToStep('')
  }

  const removeConnection = (index: number) => {
    const connections = workflowData.connections || []
    updateData({ connections: connections.filter((_, i) => i !== index) })
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Define the flow between steps. Each connection represents a path from one step to another.
      </p>
      <div className="flex gap-2">
        <select
          value={fromStep}
          onChange={(e) => setFromStep(e.target.value)}
          className="flex-1 px-3 py-2 border rounded-lg"
        >
          <option value="">From Step</option>
          {workflowData.steps?.map((step) => (
            <option key={step.id} value={step.id}>
              {step.name}
            </option>
          ))}
        </select>
        <select
          value={toStep}
          onChange={(e) => setToStep(e.target.value)}
          className="flex-1 px-3 py-2 border rounded-lg"
        >
          <option value="">To Step</option>
          {workflowData.steps?.map((step) => (
            <option key={step.id} value={step.id}>
              {step.name}
            </option>
          ))}
        </select>
        <Button onClick={addConnection}>Add Connection</Button>
      </div>
      <div className="space-y-2">
        {workflowData.connections?.map((conn, index) => {
          const from = workflowData.steps?.find((s) => s.id === conn.from)
          const to = workflowData.steps?.find((s) => s.id === conn.to)
          return (
            <Card key={index}>
              <CardContent className="p-4 flex items-center justify-between">
                <div>
                  <span className="font-medium">{from?.name || conn.from}</span>
                  <span className="mx-2 text-muted-foreground">→</span>
                  <span className="font-medium">{to?.name || conn.to}</span>
                </div>
                <Button variant="ghost" size="sm" onClick={() => removeConnection(index)}>
                  Remove
                </Button>
              </CardContent>
            </Card>
          )
        })}
        {(!workflowData.connections || workflowData.connections.length === 0) && (
          <p className="text-sm text-muted-foreground text-center py-8">
            No connections defined yet. Add connections above to define the workflow flow.
          </p>
        )}
      </div>
    </div>
  )
}

// Step 5: Triggers and Schedule
function TriggersStep({ data, updateData }: WizardStepProps) {
  const workflowData = (data as WorkflowWizardData) || {
    triggers: { type: 'manual' },
  }

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-2">Trigger Type *</label>
        <select
          value={workflowData.triggers?.type || 'manual'}
          onChange={(e) =>
            updateData({
              triggers: { ...workflowData.triggers, type: e.target.value as any },
            })
          }
          className="w-full px-3 py-2 border rounded-lg"
        >
          <option value="manual">Manual (Run on demand)</option>
          <option value="schedule">Scheduled (Run on a schedule)</option>
          <option value="webhook">Webhook (Triggered by HTTP request)</option>
          <option value="event">Event (Triggered by system events)</option>
        </select>
      </div>
      {workflowData.triggers?.type === 'schedule' && (
        <div>
          <label className="block text-sm font-medium mb-2">Schedule (Cron Expression)</label>
          <Input
            value={workflowData.triggers?.schedule || ''}
            onChange={(e) =>
              updateData({
                triggers: { ...workflowData.triggers, schedule: e.target.value },
              })
            }
            placeholder="0 0 * * * (daily at midnight)"
          />
        </div>
      )}
      {workflowData.triggers?.type === 'webhook' && (
        <div>
          <label className="block text-sm font-medium mb-2">Webhook URL</label>
          <Input
            value={workflowData.triggers?.webhookUrl || ''}
            onChange={(e) =>
              updateData({
                triggers: { ...workflowData.triggers, webhookUrl: e.target.value },
              })
            }
            placeholder="https://api.example.com/webhook"
          />
        </div>
      )}
      {workflowData.triggers?.type === 'event' && (
        <div>
          <label className="block text-sm font-medium mb-2">Event Type</label>
          <Input
            value={workflowData.triggers?.eventType || ''}
            onChange={(e) =>
              updateData({
                triggers: { ...workflowData.triggers, eventType: e.target.value },
              })
            }
            placeholder="e.g., data.ingested, ticket.created"
          />
        </div>
      )}
    </div>
  )
}

// Step 6: Review
function ReviewStep({ data }: WizardStepProps) {
  const workflowData = data as WorkflowWizardData

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Workflow Summary</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <h4 className="font-semibold mb-2">Name</h4>
            <p>{workflowData.name || 'Not set'}</p>
          </div>
          {workflowData.description && (
            <div>
              <h4 className="font-semibold mb-2">Description</h4>
              <p className="text-sm text-muted-foreground">{workflowData.description}</p>
            </div>
          )}
          <div>
            <h4 className="font-semibold mb-2">Steps ({workflowData.steps?.length || 0})</h4>
            <ul className="list-disc list-inside space-y-1">
              {workflowData.steps?.map((step) => (
                <li key={step.id} className="text-sm">
                  {step.name} ({step.type})
                </li>
              ))}
            </ul>
          </div>
          <div>
            <h4 className="font-semibold mb-2">
              Connections ({workflowData.connections?.length || 0})
            </h4>
            <ul className="list-disc list-inside space-y-1">
              {workflowData.connections?.map((conn, i) => {
                const from = workflowData.steps?.find((s) => s.id === conn.from)
                const to = workflowData.steps?.find((s) => s.id === conn.to)
                return (
                  <li key={i} className="text-sm">
                    {from?.name || conn.from} → {to?.name || conn.to}
                  </li>
                )
              })}
            </ul>
          </div>
          <div>
            <h4 className="font-semibold mb-2">Trigger</h4>
            <p className="text-sm capitalize">{workflowData.triggers?.type || 'manual'}</p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

interface WorkflowCreationWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function WorkflowCreationWizard({
  onComplete,
  onCancel,
}: WorkflowCreationWizardProps) {
  const { mutate: createWorkflow, isPending } = useCreateWorkflow()

  const handleComplete = async (data: WorkflowWizardData) => {
    const workflowDefinition = {
      steps: data.steps.map((step) => ({
        id: step.id,
        name: step.name,
        type: step.type,
        config: step.config,
        next_steps: data.connections
          ?.filter((c) => c.from === step.id)
          .map((c) => c.to) || [],
      })),
      start_step: data.steps?.[0]?.id || '',
    }

    createWorkflow(
      {
        name: data.name,
        description: data.description,
        workflow_definition: workflowDefinition,
        enabled: true,
      },
      {
        onSuccess: () => {
          showToast('Workflow created successfully', 'success')
          if (onComplete) onComplete()
        },
        onError: (error: any) => {
          showToast(error?.message || 'Failed to create workflow', 'error')
        },
      }
    )
  }

  const validateName = (data: any): boolean => {
    return !!(data as WorkflowWizardData).name?.trim()
  }

  const validateSteps = (data: any): boolean => {
    const workflowData = data as WorkflowWizardData
    return !!(workflowData.steps && workflowData.steps.length > 0)
  }

  const steps: WizardStep[] = [
    {
      id: 'name',
      title: 'Name & Description',
      description: 'Give your workflow a name and description',
      component: NameStep,
      validate: validateName,
    },
    {
      id: 'template',
      title: 'Choose Template',
      description: 'Select a template or start from scratch',
      component: TemplateStep,
      canSkip: true,
    },
    {
      id: 'steps',
      title: 'Add Steps',
      description: 'Define the steps in your workflow',
      component: StepsStep,
      validate: validateSteps,
    },
    {
      id: 'connections',
      title: 'Configure Connections',
      description: 'Define how steps connect to each other',
      component: ConnectionsStep,
      canSkip: true,
    },
    {
      id: 'triggers',
      title: 'Set Triggers',
      description: 'Configure when and how the workflow runs',
      component: TriggersStep,
    },
    {
      id: 'review',
      title: 'Review & Create',
      description: 'Review your workflow configuration',
      component: ReviewStep,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Create Workflow"
      description="Build a new workflow step by step"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

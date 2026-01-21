'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { useCreateAgent } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface AgentWizardData {
  name: string
  purpose: string
  type: string
  capabilities: string[]
  memory: {
    enabled: boolean
    contextWindow: number
  }
  behavior: {
    instructions: string
    temperature: number
  }
}

const agentTypes = [
  { id: 'customer-support', name: 'Customer Support', description: 'Handle customer inquiries' },
  { id: 'data-analysis', name: 'Data Analysis', description: 'Analyze and interpret data' },
  { id: 'content-generation', name: 'Content Generation', description: 'Generate content' },
  { id: 'custom', name: 'Custom', description: 'Build a custom agent' },
]

const availableCapabilities = [
  'semantic-search',
  'warehouse-query',
  'document-generation',
  'data-analysis',
  'code-execution',
]

function NameStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as AgentWizardData) || { name: '', purpose: '' }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Create an AI agent with specific capabilities and behavior. Agents can execute workflows,
        search data, and interact with your systems.
      </p>
      <div>
        <Input
          label="Agent Name *"
          value={wizardData.name || ''}
          onChange={(e) => updateData({ name: e.target.value })}
          placeholder="e.g., Support Assistant"
          required
        />
      </div>
      <div>
        <Textarea
          label="Purpose"
          value={wizardData.purpose || ''}
          onChange={(e) => updateData({ purpose: e.target.value })}
          placeholder="Describe what this agent does..."
          rows={4}
        />
      </div>
    </div>
  )
}

function TypeStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as AgentWizardData) || { type: '' }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Choose the type of agent you want to create.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {agentTypes.map((type) => (
          <Card
            key={type.id}
            hover
            className={wizardData.type === type.id ? 'ring-2 ring-primary' : ''}
            onClick={() => updateData({ type: type.id })}
          >
            <CardContent className="p-4">
              <h3 className="font-semibold mb-1">{type.name}</h3>
              <p className="text-sm text-muted-foreground">{type.description}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

function CapabilitiesStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as AgentWizardData) || { capabilities: [] }
  const [selected, setSelected] = useState<string[]>(wizardData.capabilities || [])

  const toggleCapability = (capability: string) => {
    const newSelection = selected.includes(capability)
      ? selected.filter((c) => c !== capability)
      : [...selected, capability]
    setSelected(newSelection)
    updateData({ capabilities: newSelection })
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Select the capabilities this agent should have.
      </p>
      <div className="space-y-2">
        {availableCapabilities.map((cap) => (
          <div key={cap} className="flex items-center gap-2">
            <input
              type="checkbox"
              id={`cap-${cap}`}
              checked={selected.includes(cap)}
              onChange={() => toggleCapability(cap)}
              className="rounded"
            />
            <label htmlFor={`cap-${cap}`} className="text-sm capitalize">
              {cap.replace('-', ' ')}
            </label>
          </div>
        ))}
      </div>
    </div>
  )
}

function MemoryStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as AgentWizardData) || {
    memory: { enabled: false, contextWindow: 4000 },
  }
  const [enabled, setEnabled] = useState(wizardData.memory?.enabled || false)
  const [contextWindow, setContextWindow] = useState(
    wizardData.memory?.contextWindow || 4000
  )

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="memory-enabled"
          checked={enabled}
          onChange={(e) => {
            setEnabled(e.target.checked)
            updateData({
              memory: { ...wizardData.memory, enabled: e.target.checked },
            })
          }}
          className="rounded"
        />
        <label htmlFor="memory-enabled" className="text-sm font-medium">
          Enable Long-term Memory
        </label>
      </div>
      {enabled && (
        <div>
          <Input
            label="Context Window (tokens)"
            type="number"
            value={contextWindow}
            onChange={(e) => {
              const value = parseInt(e.target.value) || 4000
              setContextWindow(value)
              updateData({
                memory: { ...wizardData.memory, contextWindow: value },
              })
            }}
            min={1000}
            max={32000}
          />
        </div>
      )}
    </div>
  )
}

function BehaviorStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as AgentWizardData) || {
    behavior: { instructions: '', temperature: 0.7 },
  }

  return (
    <div className="space-y-4">
      <div>
        <Textarea
          label="Agent Instructions"
          value={wizardData.behavior?.instructions || ''}
          onChange={(e) =>
            updateData({
              behavior: { ...wizardData.behavior, instructions: e.target.value },
            })
          }
          placeholder="Describe how the agent should behave, what tone to use, etc."
          rows={6}
        />
      </div>
      <div>
        <Input
          label="Temperature (0-1)"
          type="number"
          step="0.1"
          min="0"
          max="1"
          value={wizardData.behavior?.temperature || 0.7}
          onChange={(e) =>
            updateData({
              behavior: {
                ...wizardData.behavior,
                temperature: parseFloat(e.target.value) || 0.7,
              },
            })
          }
        />
      </div>
    </div>
  )
}

function TestStep({ data }: WizardStepProps) {
  const [testQuery, setTestQuery] = useState('')
  const [testResult, setTestResult] = useState<string>('')

  const handleTest = () => {
    setTestResult('Testing agent... (This is a mock test)')
    setTimeout(() => {
      setTestResult('Agent responded successfully!')
    }, 2000)
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Test your agent with a sample query to ensure it's configured correctly.
      </p>
      <div>
        <Input
          label="Test Query"
          value={testQuery}
          onChange={(e) => setTestQuery(e.target.value)}
          placeholder="Enter a test question..."
        />
      </div>
      <Button onClick={handleTest} disabled={!testQuery.trim()}>
        Test Agent
      </Button>
      {testResult && (
        <div className="p-3 rounded-lg bg-muted text-sm">{testResult}</div>
      )}
    </div>
  )
}

function ReviewStep({ data }: WizardStepProps) {
  const wizardData = data as AgentWizardData

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="p-4 space-y-3">
          <div>
            <h4 className="font-semibold mb-1">Name</h4>
            <p className="text-sm">{wizardData.name || 'Not set'}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Type</h4>
            <p className="text-sm capitalize">{wizardData.type || 'Not set'}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Capabilities ({wizardData.capabilities?.length || 0})</h4>
            <p className="text-sm">{wizardData.capabilities?.join(', ') || 'None'}</p>
          </div>
          {wizardData.memory?.enabled && (
            <div>
              <h4 className="font-semibold mb-1">Memory</h4>
              <p className="text-sm">Enabled ({wizardData.memory?.contextWindow} tokens)</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

interface AgentCreationWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function AgentCreationWizard({
  onComplete,
  onCancel,
}: AgentCreationWizardProps) {
  const { mutate: createAgent, isPending } = useCreateAgent()

  const handleComplete = async (data: AgentWizardData) => {
    createAgent(
      {
        name: data.name,
        agent_type: data.type,
        config: {
          purpose: data.purpose,
          capabilities: data.capabilities,
          memory: data.memory,
          behavior: data.behavior,
        },
        status: 'draft',
      },
      {
        onSuccess: () => {
          showToast('Agent created successfully', 'success')
          if (onComplete) onComplete()
        },
        onError: (error: any) => {
          showToast(error?.message || 'Failed to create agent', 'error')
        },
      }
    )
  }

  const validateName = (data: any): boolean => {
    return !!(data as AgentWizardData).name?.trim()
  }

  const validateType = (data: any): boolean => {
    return !!(data as AgentWizardData).type
  }

  const steps: WizardStep[] = [
    {
      id: 'name',
      title: 'Name & Purpose',
      description: 'Give your agent a name and describe its purpose',
      component: NameStep,
      validate: validateName,
    },
    {
      id: 'type',
      title: 'Choose Agent Type',
      description: 'Select the type of agent',
      component: TypeStep,
      validate: validateType,
    },
    {
      id: 'capabilities',
      title: 'Configure Capabilities',
      description: 'Select what the agent can do',
      component: CapabilitiesStep,
    },
    {
      id: 'memory',
      title: 'Set Up Memory',
      description: 'Configure long-term memory settings',
      component: MemoryStep,
      canSkip: true,
    },
    {
      id: 'behavior',
      title: 'Define Behavior',
      description: 'Set agent instructions and parameters',
      component: BehaviorStep,
    },
    {
      id: 'test',
      title: 'Test Agent',
      description: 'Test your agent with a sample query',
      component: TestStep,
      canSkip: true,
    },
    {
      id: 'review',
      title: 'Review & Deploy',
      description: 'Review your agent configuration',
      component: ReviewStep,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Create Agent"
      description="Build a new AI agent step by step"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

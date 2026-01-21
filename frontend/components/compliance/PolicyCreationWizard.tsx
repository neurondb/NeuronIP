'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { showToast } from '@/components/ui/Toast'
import Tooltip from '@/components/ui/Tooltip'
import Warning from '@/components/ui/Warning'
import HelpText from '@/components/ui/HelpText'
import { InformationCircleIcon } from '@heroicons/react/24/outline'

interface PolicyWizardData {
  name: string
  description: string
  type: string
  rules: Array<{
    condition: string
    action: string
  }>
  actions: {
    alert: boolean
    block: boolean
    log: boolean
  }
  scope: {
    resources: string[]
    users: string[]
  }
}

const policyTypes = [
  { id: 'data-protection', name: 'Data Protection', description: 'Protect sensitive data' },
  { id: 'access-control', name: 'Access Control', description: 'Control who can access what' },
  { id: 'compliance', name: 'Compliance', description: 'Ensure regulatory compliance' },
  { id: 'custom', name: 'Custom', description: 'Create a custom policy' },
]

function NameStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as PolicyWizardData) || { name: '', description: '' }

  return (
    <div className="space-y-4">
      <Warning
        message="Policies use semantic matching and anomaly detection. Ensure rules are clearly defined to avoid false positives."
        severity="medium"
        title="Policy Design"
      />
      <HelpText
        variant="inline"
        content={
          <p>
            Compliance policies help ensure regulatory compliance and data protection. Policies use
            semantic matching to identify violations.
          </p>
        }
        link="/docs/features/compliance"
        linkText="Learn more about compliance"
      />
      <div>
        <div className="flex items-center gap-2 mb-1">
          <label className="block text-sm font-medium">Policy Name *</label>
          <Tooltip
            content="Use a clear, descriptive name that identifies the policy's purpose, e.g., 'GDPR Data Protection' or 'HIPAA Compliance'"
            variant="info"
          >
            <InformationCircleIcon className="h-4 w-4 text-muted-foreground cursor-help" />
          </Tooltip>
        </div>
        <Input
          value={wizardData.name || ''}
          onChange={(e) => updateData({ name: e.target.value })}
          placeholder="e.g., GDPR Data Protection Policy"
          required
        />
      </div>
      <div>
        <div className="flex items-center gap-2 mb-1">
          <label className="block text-sm font-medium">Description</label>
          <Tooltip
            content="Describe what this policy enforces and which regulations or standards it helps comply with"
            variant="info"
          >
            <InformationCircleIcon className="h-4 w-4 text-muted-foreground cursor-help" />
          </Tooltip>
        </div>
        <Textarea
          value={wizardData.description || ''}
          onChange={(e) => updateData({ description: e.target.value })}
          placeholder="Describe what this policy does..."
          rows={4}
        />
      </div>
    </div>
  )
}

function TypeStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as PolicyWizardData) || { type: '' }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Choose the type of policy you want to create.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {policyTypes.map((type) => (
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

function RulesStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as PolicyWizardData) || { rules: [] }
  const [newCondition, setNewCondition] = useState('')
  const [newAction, setNewAction] = useState('block')

  const addRule = () => {
    if (!newCondition.trim()) return
    const newRule = {
      condition: newCondition,
      action: newAction,
    }
    updateData({ rules: [...(wizardData.rules || []), newRule] })
    setNewCondition('')
  }

  const removeRule = (index: number) => {
    updateData({
      rules: wizardData.rules?.filter((_, i) => i !== index) || [],
    })
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Define the rules and conditions for this policy.
      </p>
      <div className="space-y-2">
        <Input
          label="Condition"
          value={newCondition}
          onChange={(e) => setNewCondition(e.target.value)}
          placeholder="e.g., data contains PII"
        />
        <select
          value={newAction}
          onChange={(e) => setNewAction(e.target.value)}
          className="w-full px-3 py-2 border rounded-lg"
        >
          <option value="block">Block</option>
          <option value="alert">Alert</option>
          <option value="log">Log Only</option>
        </select>
        <Button onClick={addRule} disabled={!newCondition.trim()}>
          Add Rule
        </Button>
      </div>
      <div className="space-y-2 mt-4">
        {wizardData.rules?.map((rule, index) => (
          <Card key={index}>
            <CardContent className="p-4 flex items-center justify-between">
              <div>
                <p className="font-medium">{rule.condition}</p>
                <p className="text-sm text-muted-foreground">Action: {rule.action}</p>
              </div>
              <Button variant="ghost" size="sm" onClick={() => removeRule(index)}>
                Remove
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

function ActionsStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as PolicyWizardData) || {
    actions: { alert: true, block: false, log: true },
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Configure what happens when a policy violation is detected.
      </p>
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="action-alert"
            checked={wizardData.actions?.alert || false}
            onChange={(e) =>
              updateData({
                actions: { ...wizardData.actions, alert: e.target.checked },
              })
            }
            className="rounded"
          />
          <label htmlFor="action-alert" className="text-sm font-medium">
            Send Alert
          </label>
        </div>
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="action-block"
            checked={wizardData.actions?.block || false}
            onChange={(e) =>
              updateData({
                actions: { ...wizardData.actions, block: e.target.checked },
              })
            }
            className="rounded"
          />
          <label htmlFor="action-block" className="text-sm font-medium">
            Block Action
          </label>
        </div>
        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="action-log"
            checked={wizardData.actions?.log || false}
            onChange={(e) =>
              updateData({
                actions: { ...wizardData.actions, log: e.target.checked },
              })
            }
            className="rounded"
          />
          <label htmlFor="action-log" className="text-sm font-medium">
            Log Violation
          </label>
        </div>
      </div>
    </div>
  )
}

function ScopeStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as PolicyWizardData) || {
    scope: { resources: [], users: [] },
  }
  const [resourceInput, setResourceInput] = useState('')
  const [userInput, setUserInput] = useState('')

  const addResource = () => {
    if (!resourceInput.trim()) return
    updateData({
      scope: {
        ...wizardData.scope,
        resources: [...(wizardData.scope?.resources || []), resourceInput.trim()],
      },
    })
    setResourceInput('')
  }

  const addUser = () => {
    if (!userInput.trim()) return
    updateData({
      scope: {
        ...wizardData.scope,
        users: [...(wizardData.scope?.users || []), userInput.trim()],
      },
    })
    setUserInput('')
  }

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-2">Resources</label>
        <div className="flex gap-2">
          <Input
            value={resourceInput}
            onChange={(e) => setResourceInput(e.target.value)}
            placeholder="Resource ID or pattern"
            onKeyPress={(e) => e.key === 'Enter' && addResource()}
          />
          <Button onClick={addResource}>Add</Button>
        </div>
        <div className="mt-2 flex flex-wrap gap-2">
          {wizardData.scope?.resources?.map((resource, i) => (
            <span
              key={i}
              className="px-2 py-1 bg-muted rounded text-sm flex items-center gap-1"
            >
              {resource}
              <button
                onClick={() => {
                  updateData({
                    scope: {
                      ...wizardData.scope,
                      resources: wizardData.scope?.resources?.filter((_, idx) => idx !== i) || [],
                    },
                  })
                }}
                className="text-muted-foreground hover:text-foreground"
              >
                ×
              </button>
            </span>
          ))}
        </div>
      </div>
      <div>
        <label className="block text-sm font-medium mb-2">Users (Optional)</label>
        <div className="flex gap-2">
          <Input
            value={userInput}
            onChange={(e) => setUserInput(e.target.value)}
            placeholder="User ID or email"
            onKeyPress={(e) => e.key === 'Enter' && addUser()}
          />
          <Button onClick={addUser}>Add</Button>
        </div>
        <div className="mt-2 flex flex-wrap gap-2">
          {wizardData.scope?.users?.map((user, i) => (
            <span
              key={i}
              className="px-2 py-1 bg-muted rounded text-sm flex items-center gap-1"
            >
              {user}
              <button
                onClick={() => {
                  updateData({
                    scope: {
                      ...wizardData.scope,
                      users: wizardData.scope?.users?.filter((_, idx) => idx !== i) || [],
                    },
                  })
                }}
                className="text-muted-foreground hover:text-foreground"
              >
                ×
              </button>
            </span>
          ))}
        </div>
      </div>
    </div>
  )
}

function ReviewStep({ data }: WizardStepProps) {
  const wizardData = data as PolicyWizardData

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
            <h4 className="font-semibold mb-1">Rules ({wizardData.rules?.length || 0})</h4>
            <ul className="list-disc list-inside space-y-1 text-sm">
              {wizardData.rules?.map((rule, i) => (
                <li key={i}>
                  {rule.condition} → {rule.action}
                </li>
              ))}
            </ul>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Actions</h4>
            <p className="text-sm">
              {[
                wizardData.actions?.alert && 'Alert',
                wizardData.actions?.block && 'Block',
                wizardData.actions?.log && 'Log',
              ]
                .filter(Boolean)
                .join(', ') || 'None'}
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

interface PolicyCreationWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function PolicyCreationWizard({
  onComplete,
  onCancel,
}: PolicyCreationWizardProps) {
  const handleComplete = async (data: PolicyWizardData) => {
    // In real implementation, call API to create policy
    showToast('Policy created successfully', 'success')
    if (onComplete) onComplete()
  }

  const validateName = (data: any): boolean => {
    return !!(data as PolicyWizardData).name?.trim()
  }

  const validateType = (data: any): boolean => {
    return !!(data as PolicyWizardData).type
  }

  const steps: WizardStep[] = [
    {
      id: 'name',
      title: 'Name & Description',
      description: 'Give your policy a name and description',
      component: NameStep,
      validate: validateName,
    },
    {
      id: 'type',
      title: 'Choose Policy Type',
      description: 'Select the type of policy',
      component: TypeStep,
      validate: validateType,
    },
    {
      id: 'rules',
      title: 'Define Rules',
      description: 'Set up policy rules and conditions',
      component: RulesStep,
    },
    {
      id: 'actions',
      title: 'Set Violation Actions',
      description: 'Configure what happens on violation',
      component: ActionsStep,
    },
    {
      id: 'scope',
      title: 'Assign Scope',
      description: 'Define which resources and users this applies to',
      component: ScopeStep,
      canSkip: true,
    },
    {
      id: 'review',
      title: 'Review & Activate',
      description: 'Review your policy configuration',
      component: ReviewStep,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Create Compliance Policy"
      description="Set up a new compliance policy"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

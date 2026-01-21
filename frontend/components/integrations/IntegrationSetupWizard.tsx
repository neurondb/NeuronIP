'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { useCreateIntegration, useTestIntegration } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline'

interface IntegrationWizardData {
  type: string
  name: string
  credentials: Record<string, string>
  settings: Record<string, any>
  webhooks: {
    enabled: boolean
    url?: string
    events?: string[]
  }
}

const integrationTypes = [
  { id: 'slack', name: 'Slack', description: 'Connect to Slack workspace' },
  { id: 'teams', name: 'Microsoft Teams', description: 'Connect to Teams' },
  { id: 'zendesk', name: 'Zendesk', description: 'Connect to Zendesk' },
  { id: 'salesforce', name: 'Salesforce', description: 'Connect to Salesforce' },
]

function TypeStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as IntegrationWizardData) || { type: '' }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Select the type of integration you want to set up.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {integrationTypes.map((type) => (
          <Card
            key={type.id}
            hover
            className={wizardData.type === type.id ? 'ring-2 ring-primary' : ''}
            onClick={() => updateData({ type: type.id, name: type.name })}
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

function CredentialsStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as IntegrationWizardData) || { credentials: {} }
  const [apiKey, setApiKey] = useState(wizardData.credentials?.apiKey || '')
  const [apiSecret, setApiSecret] = useState(wizardData.credentials?.apiSecret || '')

  const handleUpdate = () => {
    updateData({
      credentials: {
        apiKey,
        apiSecret,
      },
    })
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Enter your integration credentials. These will be securely stored.
      </p>
      <div>
        <Input
          label="API Key"
          type="password"
          value={apiKey}
          onChange={(e) => {
            setApiKey(e.target.value)
            handleUpdate()
          }}
          placeholder="Enter API key"
          required
        />
      </div>
      <div>
        <Input
          label="API Secret"
          type="password"
          value={apiSecret}
          onChange={(e) => {
            setApiSecret(e.target.value)
            handleUpdate()
          }}
          placeholder="Enter API secret"
          required
        />
      </div>
    </div>
  )
}

function SettingsStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as IntegrationWizardData) || { settings: {} }
  const [webhookUrl, setWebhookUrl] = useState(wizardData.settings?.webhookUrl || '')

  return (
    <div className="space-y-4">
      <div>
        <Input
          label="Webhook URL (Optional)"
          value={webhookUrl}
          onChange={(e) => {
            setWebhookUrl(e.target.value)
            updateData({ settings: { ...wizardData.settings, webhookUrl: e.target.value } })
          }}
          placeholder="https://api.example.com/webhook"
        />
      </div>
    </div>
  )
}

function TestStep({ data }: WizardStepProps) {
  const wizardData = data as IntegrationWizardData
  const [testResult, setTestResult] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')
  const [testError, setTestError] = useState<string>('')

  const handleTest = () => {
    setTestResult('testing')
    setTestError('')
    // Simulate test - in real implementation, use testIntegration mutation
    setTimeout(() => {
      // Mock test result
      const success = Math.random() > 0.3
      if (success) {
        setTestResult('success')
      } else {
        setTestResult('error')
        setTestError('Connection failed. Please check your credentials.')
      }
    }, 2000)
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Test the connection to ensure everything is configured correctly.
      </p>
      <Button onClick={handleTest} disabled={testResult === 'testing'} isLoading={testResult === 'testing'}>
        Test Connection
      </Button>
      {testResult === 'success' && (
        <div className="flex items-center gap-2 text-green-600">
          <CheckCircleIcon className="h-5 w-5" />
          <span>Connection successful!</span>
        </div>
      )}
      {testResult === 'error' && (
        <div className="flex items-center gap-2 text-destructive">
          <XCircleIcon className="h-5 w-5" />
          <span>{testError || 'Connection failed'}</span>
        </div>
      )}
    </div>
  )
}

function WebhooksStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as IntegrationWizardData) || { webhooks: { enabled: false } }
  const [enabled, setEnabled] = useState(wizardData.webhooks?.enabled || false)
  const [url, setUrl] = useState(wizardData.webhooks?.url || '')

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="webhooks-enabled"
          checked={enabled}
          onChange={(e) => {
            setEnabled(e.target.checked)
            updateData({
              webhooks: { ...wizardData.webhooks, enabled: e.target.checked },
            })
          }}
          className="rounded"
        />
        <label htmlFor="webhooks-enabled" className="text-sm font-medium">
          Enable Webhooks
        </label>
      </div>
      {enabled && (
        <div>
          <Input
            label="Webhook URL"
            value={url}
            onChange={(e) => {
              setUrl(e.target.value)
              updateData({
                webhooks: { ...wizardData.webhooks, url: e.target.value },
              })
            }}
            placeholder="https://api.example.com/webhook"
          />
        </div>
      )}
    </div>
  )
}

function ReviewStep({ data }: WizardStepProps) {
  const wizardData = data as IntegrationWizardData

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="p-4 space-y-3">
          <div>
            <h4 className="font-semibold mb-1">Integration Type</h4>
            <p className="text-sm">{wizardData.type || 'Not set'}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Name</h4>
            <p className="text-sm">{wizardData.name || 'Not set'}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Credentials</h4>
            <p className="text-sm">
              {wizardData.credentials?.apiKey ? 'Configured' : 'Not configured'}
            </p>
          </div>
          {wizardData.webhooks?.enabled && (
            <div>
              <h4 className="font-semibold mb-1">Webhooks</h4>
              <p className="text-sm">Enabled</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

interface IntegrationSetupWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function IntegrationSetupWizard({
  onComplete,
  onCancel,
}: IntegrationSetupWizardProps) {
  const { mutate: createIntegration, isPending } = useCreateIntegration()

  const handleComplete = async (data: IntegrationWizardData) => {
    createIntegration(
      {
        type: data.type,
        name: data.name,
        config: {
          credentials: data.credentials,
          settings: data.settings,
          webhooks: data.webhooks,
        },
        enabled: true,
      },
      {
        onSuccess: () => {
          showToast('Integration created successfully', 'success')
          if (onComplete) onComplete()
        },
        onError: (error: any) => {
          showToast(error?.message || 'Failed to create integration', 'error')
        },
      }
    )
  }

  const validateType = (data: any): boolean => {
    return !!(data as IntegrationWizardData).type
  }

  const validateCredentials = (data: any): boolean => {
    const wizardData = data as IntegrationWizardData
    return !!(wizardData.credentials?.apiKey && wizardData.credentials?.apiSecret)
  }

  const steps: WizardStep[] = [
    {
      id: 'type',
      title: 'Choose Integration Type',
      description: 'Select the service you want to integrate',
      component: TypeStep,
      validate: validateType,
    },
    {
      id: 'credentials',
      title: 'Provide Credentials',
      description: 'Enter your API credentials',
      component: CredentialsStep,
      validate: validateCredentials,
    },
    {
      id: 'settings',
      title: 'Configure Settings',
      description: 'Set up integration-specific settings',
      component: SettingsStep,
      canSkip: true,
    },
    {
      id: 'test',
      title: 'Test Connection',
      description: 'Verify the connection works',
      component: TestStep,
      canSkip: true,
    },
    {
      id: 'webhooks',
      title: 'Configure Webhooks',
      description: 'Set up webhook notifications (optional)',
      component: WebhooksStep,
      canSkip: true,
      isOptional: true,
    },
    {
      id: 'review',
      title: 'Review & Activate',
      description: 'Review your configuration',
      component: ReviewStep,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Set Up Integration"
      description="Connect an external service to NeuronIP"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

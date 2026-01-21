'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { useCreateAPIKey } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import {
  SparklesIcon,
  KeyIcon,
  ServerIcon,
  MagnifyingGlassIcon,
  RocketLaunchIcon,
  CheckCircleIcon,
} from '@heroicons/react/24/outline'

interface OnboardingWizardData {
  apiKey: {
    name: string
    created: boolean
    key?: string
  }
  dataSource: {
    connected: boolean
    type?: string
  }
  collection: {
    created: boolean
    name?: string
  }
}

function WelcomeStep() {
  return (
    <div className="space-y-4 text-center">
      <div className="flex justify-center mb-4">
        <SparklesIcon className="h-16 w-16 text-primary" />
      </div>
      <h3 className="text-2xl font-bold">Welcome to NeuronIP!</h3>
      <p className="text-muted-foreground">
        Let's get you started with a quick setup. This will only take a few minutes.
      </p>
      <Card className="mt-6">
        <CardContent className="p-4">
          <h4 className="font-semibold mb-2">What you'll set up:</h4>
          <ul className="text-sm text-left space-y-1 text-muted-foreground">
            <li>• Create your first API key</li>
            <li>• Connect a data source (optional)</li>
            <li>• Set up semantic search (optional)</li>
            <li>• Explore key features</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  )
}

function APIKeyStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as OnboardingWizardData) || { apiKey: { name: '', created: false } }
  const [keyName, setKeyName] = useState(wizardData.apiKey?.name || '')
  const { mutate: createAPIKey, isPending } = useCreateAPIKey()
  const [createdKey, setCreatedKey] = useState<string>('')

  const handleCreate = () => {
    if (!keyName.trim()) {
      showToast('Please enter a name for your API key', 'warning')
      return
    }

    createAPIKey(
      { name: keyName },
      {
        onSuccess: (response: any) => {
          const key = response.key || response.api_key
          setCreatedKey(key)
          updateData({
            apiKey: { name: keyName, created: true, key },
          })
          showToast('API key created successfully', 'success')
        },
        onError: (error: any) => {
          showToast(error?.message || 'Failed to create API key', 'error')
        },
      }
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3 mb-4">
        <KeyIcon className="h-8 w-8 text-primary" />
        <div>
          <h3 className="font-semibold">Create Your First API Key</h3>
          <p className="text-sm text-muted-foreground">
            API keys allow you to authenticate requests to the NeuronIP API
          </p>
        </div>
      </div>
      {!wizardData.apiKey?.created ? (
        <div className="space-y-4">
          <Input
            label="API Key Name"
            value={keyName}
            onChange={(e) => setKeyName(e.target.value)}
            placeholder="e.g., My First API Key"
            required
          />
          <Button onClick={handleCreate} isLoading={isPending} disabled={!keyName.trim()}>
            Create API Key
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="p-4 rounded-lg bg-muted">
            <p className="text-sm font-medium mb-2">Your API Key (save this now!):</p>
            <code className="block p-2 bg-background rounded text-sm break-all">
              {createdKey || wizardData.apiKey?.key}
            </code>
            <p className="text-xs text-muted-foreground mt-2">
              This key will only be shown once. Make sure to save it securely.
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => {
              navigator.clipboard.writeText(createdKey || wizardData.apiKey?.key || '')
              showToast('API key copied to clipboard', 'success')
            }}
          >
            Copy to Clipboard
          </Button>
        </div>
      )}
    </div>
  )
}

function DataSourceStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as OnboardingWizardData) || { dataSource: { connected: false } }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3 mb-4">
        <ServerIcon className="h-8 w-8 text-primary" />
        <div>
          <h3 className="font-semibold">Connect a Data Source (Optional)</h3>
          <p className="text-sm text-muted-foreground">
            You can skip this step and connect a data source later
          </p>
        </div>
      </div>
      <Card>
        <CardContent className="p-4">
          <p className="text-sm text-muted-foreground mb-4">
            Connect your data sources to enable warehouse Q&A and data analysis features.
          </p>
          <Button
            variant="outline"
            onClick={() => {
              updateData({ dataSource: { connected: true, type: 'postgresql' } })
            }}
          >
            Connect Data Source
          </Button>
        </CardContent>
      </Card>
      <Button
        variant="ghost"
        onClick={() => {
          updateData({ dataSource: { connected: false } })
        }}
      >
        Skip for Now
      </Button>
    </div>
  )
}

function CollectionStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as OnboardingWizardData) || { collection: { created: false } }
  const [collectionName, setCollectionName] = useState('')

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3 mb-4">
        <MagnifyingGlassIcon className="h-8 w-8 text-primary" />
        <div>
          <h3 className="font-semibold">Set Up Semantic Search (Optional)</h3>
          <p className="text-sm text-muted-foreground">
            Create a collection to enable semantic search capabilities
          </p>
        </div>
      </div>
      <Card>
        <CardContent className="p-4 space-y-4">
          <Input
            label="Collection Name"
            value={collectionName}
            onChange={(e) => setCollectionName(e.target.value)}
            placeholder="e.g., Knowledge Base"
          />
          <Button
            variant="outline"
            onClick={() => {
              if (collectionName.trim()) {
                updateData({
                  collection: { created: true, name: collectionName },
                })
                showToast('Collection created', 'success')
              }
            }}
            disabled={!collectionName.trim()}
          >
            Create Collection
          </Button>
        </CardContent>
      </Card>
      <Button
        variant="ghost"
        onClick={() => {
          updateData({ collection: { created: false } })
        }}
      >
        Skip for Now
      </Button>
    </div>
  )
}

function FeaturesStep() {
  const features = [
    {
      icon: MagnifyingGlassIcon,
      title: 'Semantic Search',
      description: 'Search your knowledge base by meaning',
    },
    {
      icon: ServerIcon,
      title: 'Warehouse Q&A',
      description: 'Ask questions about your data',
    },
    {
      icon: SparklesIcon,
      title: 'AI Agents',
      description: 'Build intelligent workflows',
    },
  ]

  return (
    <div className="space-y-4">
      <h3 className="font-semibold mb-4">Explore Key Features</h3>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {features.map((feature, index) => {
          const Icon = feature.icon
          return (
            <Card key={index} hover>
              <CardContent className="p-4 text-center">
                <Icon className="h-8 w-8 mx-auto mb-2 text-primary" />
                <h4 className="font-semibold mb-1">{feature.title}</h4>
                <p className="text-sm text-muted-foreground">{feature.description}</p>
              </CardContent>
            </Card>
          )
        })}
      </div>
      <div className="mt-6 p-4 rounded-lg bg-primary/10">
        <p className="text-sm">
          <strong>Tip:</strong> Check out the{' '}
          <a href="/dashboard/why-neuronip" className="text-primary underline">
            Why NeuronIP
          </a>{' '}
          page to learn more about all available features.
        </p>
      </div>
    </div>
  )
}

function CompleteStep({ data }: WizardStepProps) {
  const wizardData = data as OnboardingWizardData

  return (
    <div className="space-y-4 text-center">
      <div className="flex justify-center mb-4">
        <CheckCircleIcon className="h-16 w-16 text-green-600" />
      </div>
      <h3 className="text-2xl font-bold">You're All Set!</h3>
      <p className="text-muted-foreground">
        You've completed the onboarding. You can now start using NeuronIP.
      </p>
      <Card className="mt-6">
        <CardContent className="p-4 space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm">API Key Created</span>
            <span className="text-sm font-medium">
              {wizardData.apiKey?.created ? '✓' : 'Skip'}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm">Data Source</span>
            <span className="text-sm font-medium">
              {wizardData.dataSource?.connected ? '✓ Connected' : 'Skipped'}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm">Semantic Collection</span>
            <span className="text-sm font-medium">
              {wizardData.collection?.created ? '✓ Created' : 'Skipped'}
            </span>
          </div>
        </CardContent>
      </Card>
      <div className="mt-6">
        <Button onClick={() => (window.location.href = '/dashboard')}>
          <RocketLaunchIcon className="h-4 w-4 mr-2" />
          Go to Dashboard
        </Button>
      </div>
    </div>
  )
}

interface OnboardingWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function OnboardingWizard({ onComplete, onCancel }: OnboardingWizardProps) {
  const handleComplete = async (data: OnboardingWizardData) => {
    // Mark onboarding as complete (could store in localStorage or call API)
    if (typeof window !== 'undefined') {
      localStorage.setItem('onboarding_completed', 'true')
    }
    if (onComplete) onComplete()
  }

  const steps: WizardStep[] = [
    {
      id: 'welcome',
      title: 'Welcome',
      description: 'Get started with NeuronIP',
      component: WelcomeStep,
      canSkip: false,
    },
    {
      id: 'api-key',
      title: 'Create API Key',
      description: 'Set up your first API key',
      component: APIKeyStep,
      canSkip: true,
    },
    {
      id: 'data-source',
      title: 'Connect Data Source',
      description: 'Link your data (optional)',
      component: DataSourceStep,
      canSkip: true,
      isOptional: true,
    },
    {
      id: 'collection',
      title: 'Set Up Semantic Search',
      description: 'Create a collection (optional)',
      component: CollectionStep,
      canSkip: true,
      isOptional: true,
    },
    {
      id: 'features',
      title: 'Explore Features',
      description: 'Learn about key capabilities',
      component: FeaturesStep,
      canSkip: true,
    },
    {
      id: 'complete',
      title: 'Complete',
      description: "You're ready to go!",
      component: CompleteStep,
      canSkip: false,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Welcome to NeuronIP"
      description="Let's get you set up in just a few steps"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

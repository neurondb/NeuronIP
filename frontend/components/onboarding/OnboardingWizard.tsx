'use client'

import {
  SparklesIcon,
  KeyIcon,
  ServerIcon,
  MagnifyingGlassIcon,
  RocketLaunchIcon,
  CheckCircleIcon,
  ArrowRightIcon,
} from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import { useState } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Input from '@/components/ui/Input'
import { showToast } from '@/components/ui/Toast'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { useCreateAPIKey } from '@/lib/api/queries'
import { microcopy } from '@/lib/copy/microcopy'
import { useOnboarding } from '@/lib/hooks/useOnboarding'

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

function WelcomeStep({ nextStep }: WizardStepProps) {
  const { markSkipped } = useOnboarding()

  const handleSkip = () => {
    markSkipped()
    // Navigate directly to dashboard
    if (typeof window !== 'undefined') {
      window.location.href = '/'
    }
  }

  return (
    <div className="space-y-6 text-center">
      <div className="flex justify-center mb-6">
        <div className="relative">
          <SparklesIcon className="h-20 w-20 text-primary" />
          <motion.div
            className="absolute inset-0 rounded-full bg-primary/20 blur-xl"
            animate={{ scale: [1, 1.2, 1], opacity: [0.5, 0.8, 0.5] }}
            transition={{ duration: 2, repeat: Infinity }}
          />
        </div>
      </div>
      <h3 className="text-3xl font-bold">{microcopy.onboarding.welcome.title}</h3>
      <p className="text-lg text-muted-foreground max-w-md mx-auto">
        {microcopy.onboarding.welcome.subtitle}
      </p>
      <p className="text-base text-muted-foreground max-w-lg mx-auto">
        {microcopy.onboarding.welcome.description}
      </p>

      <div className="flex flex-col sm:flex-row gap-3 justify-center mt-8">
        <Button
          onClick={nextStep}
          size="lg"
          className="min-w-[200px]"
        >
          {microcopy.onboarding.welcome.startSetup}
          <ArrowRightIcon className="h-5 w-5 ml-2" />
        </Button>
        <Button
          onClick={handleSkip}
          variant="outline"
          size="lg"
          className="min-w-[200px]"
        >
          {microcopy.onboarding.welcome.skipAndExplore}
        </Button>
      </div>

      <Card className="mt-8 max-w-md mx-auto">
        <CardContent className="p-6">
          <h4 className="font-semibold mb-4 text-left">{microcopy.onboarding.welcome.whatYoullSetUp}</h4>
          <ul className="text-sm text-left space-y-2 text-muted-foreground">
            <li className="flex items-start gap-2">
              <CheckCircleIcon className="h-5 w-5 text-primary flex-shrink-0 mt-0.5" />
              <span>{microcopy.onboarding.welcome.apiKey}</span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircleIcon className="h-5 w-5 text-primary flex-shrink-0 mt-0.5" />
              <span>{microcopy.onboarding.welcome.dataSource}</span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircleIcon className="h-5 w-5 text-primary flex-shrink-0 mt-0.5" />
              <span>{microcopy.onboarding.welcome.semanticSearch}</span>
            </li>
            <li className="flex items-start gap-2">
              <CheckCircleIcon className="h-5 w-5 text-primary flex-shrink-0 mt-0.5" />
              <span>{microcopy.onboarding.welcome.exploreFeatures}</span>
            </li>
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
      showToast('Please give your key a name', 'warning')
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
          showToast('Your API key is ready!', 'success')
        },
        onError: (error: any) => {
          showToast(error?.message || "Couldn't create your key. Try again?", 'error')
        },
      }
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="p-3 rounded-lg bg-primary/10">
          <KeyIcon className="h-8 w-8 text-primary" />
        </div>
        <div>
          <h3 className="text-xl font-semibold">{microcopy.onboarding.apiKey.title}</h3>
          <p className="text-sm text-muted-foreground mt-1">
            {microcopy.onboarding.apiKey.description}
          </p>
        </div>
      </div>
      {!wizardData.apiKey?.created ? (
        <div className="space-y-4">
          <Input
            label={microcopy.onboarding.apiKey.nameLabel}
            value={keyName}
            onChange={(e) => setKeyName(e.target.value)}
            placeholder={microcopy.onboarding.apiKey.namePlaceholder}
            required
          />
          <Button onClick={handleCreate} isLoading={isPending} disabled={!keyName.trim()} size="lg" className="w-full">
            {microcopy.onboarding.apiKey.createButton}
          </Button>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="p-6 rounded-lg bg-muted border border-border">
            <p className="text-sm font-medium mb-3">{microcopy.onboarding.apiKey.created}</p>
            <code className="block p-4 bg-background rounded-lg text-sm break-all border border-border font-mono">
              {createdKey || wizardData.apiKey?.key}
            </code>
            <p className="text-xs text-muted-foreground mt-3">
              {microcopy.onboarding.apiKey.saveNotice}
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => {
              navigator.clipboard.writeText(createdKey || wizardData.apiKey?.key || '')
              showToast(microcopy.onboarding.apiKey.copied, 'success')
            }}
            size="lg"
            className="w-full"
          >
            {microcopy.onboarding.apiKey.copyButton}
          </Button>
        </div>
      )}
    </div>
  )
}

function DataSourceStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as OnboardingWizardData) || { dataSource: { connected: false } }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="p-3 rounded-lg bg-primary/10">
          <ServerIcon className="h-8 w-8 text-primary" />
        </div>
        <div>
          <h3 className="text-xl font-semibold">{microcopy.onboarding.dataSource.title}</h3>
          <p className="text-sm text-muted-foreground mt-1">
            {microcopy.onboarding.dataSource.description}
          </p>
        </div>
      </div>
      <Card>
        <CardContent className="p-6">
          <p className="text-sm text-muted-foreground mb-6">
            Link your databases to ask questions and get insights from your data.
          </p>
          <div className="flex flex-col sm:flex-row gap-3">
            <Button
              variant="outline"
              onClick={() => {
                updateData({ dataSource: { connected: true, type: 'postgresql' } })
                showToast('Data source connected!', 'success')
              }}
              size="lg"
              className="flex-1"
            >
              {microcopy.onboarding.dataSource.connectButton}
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                updateData({ dataSource: { connected: false } })
              }}
              size="lg"
            >
              {microcopy.onboarding.dataSource.skipButton}
            </Button>
          </div>
          <p className="text-xs text-muted-foreground mt-4 text-center">
            {microcopy.onboarding.dataSource.optional}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

function CollectionStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as OnboardingWizardData) || { collection: { created: false } }
  const [collectionName, setCollectionName] = useState('')

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="p-3 rounded-lg bg-primary/10">
          <MagnifyingGlassIcon className="h-8 w-8 text-primary" />
        </div>
        <div>
          <h3 className="text-xl font-semibold">{microcopy.onboarding.semanticSearch.title}</h3>
          <p className="text-sm text-muted-foreground mt-1">
            {microcopy.onboarding.semanticSearch.description}
          </p>
        </div>
      </div>
      <Card>
        <CardContent className="p-6 space-y-4">
          <Input
            label={microcopy.onboarding.semanticSearch.collectionNameLabel}
            value={collectionName}
            onChange={(e) => setCollectionName(e.target.value)}
            placeholder={microcopy.onboarding.semanticSearch.collectionNamePlaceholder}
          />
          <div className="flex flex-col sm:flex-row gap-3">
            <Button
              variant="outline"
              onClick={() => {
                if (collectionName.trim()) {
                  updateData({
                    collection: { created: true, name: collectionName },
                  })
                  showToast('Collection created!', 'success')
                }
              }}
              disabled={!collectionName.trim()}
              size="lg"
              className="flex-1"
            >
              {microcopy.onboarding.semanticSearch.createButton}
            </Button>
            <Button
              variant="ghost"
              onClick={() => {
                updateData({ collection: { created: false } })
              }}
              size="lg"
            >
              {microcopy.onboarding.semanticSearch.skipButton}
            </Button>
          </div>
          <p className="text-xs text-muted-foreground mt-4 text-center">
            {microcopy.onboarding.semanticSearch.optional}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

function FeaturesStep() {
  const features = [
    {
      icon: MagnifyingGlassIcon,
      title: microcopy.onboarding.features.semanticSearch.title,
      description: microcopy.onboarding.features.semanticSearch.description,
    },
    {
      icon: ServerIcon,
      title: microcopy.onboarding.features.warehouse.title,
      description: microcopy.onboarding.features.warehouse.description,
    },
    {
      icon: SparklesIcon,
      title: microcopy.onboarding.features.agents.title,
      description: microcopy.onboarding.features.agents.description,
    },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-xl font-semibold mb-2">{microcopy.onboarding.features.title}</h3>
        <p className="text-sm text-muted-foreground">{microcopy.onboarding.features.description}</p>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {features.map((feature, index) => {
          const Icon = feature.icon
          return (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
            >
              <Card className="h-full hover:border-primary transition-colors">
                <CardContent className="p-6 text-center">
                  <div className="p-3 rounded-lg bg-primary/10 w-fit mx-auto mb-4">
                    <Icon className="h-8 w-8 text-primary" />
                  </div>
                  <h4 className="font-semibold mb-2">{feature.title}</h4>
                  <p className="text-sm text-muted-foreground">{feature.description}</p>
                </CardContent>
              </Card>
            </motion.div>
          )
        })}
      </div>
      <div className="mt-6 p-4 rounded-lg bg-primary/10 border border-primary/20">
        <p className="text-sm">
          <strong>Tip:</strong> Check out the{' '}
          <a href="/why-neuronip" className="text-primary underline hover:text-primary/80">
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
    <div className="space-y-6 text-center">
      <motion.div
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ type: 'spring', duration: 0.5 }}
        className="flex justify-center mb-6"
      >
        <div className="relative">
          <CheckCircleIcon className="h-20 w-20 text-green-600 dark:text-green-500" />
          <motion.div
            className="absolute inset-0 rounded-full bg-green-500/20 blur-xl"
            animate={{ scale: [1, 1.2, 1], opacity: [0.5, 0.8, 0.5] }}
            transition={{ duration: 2, repeat: Infinity }}
          />
        </div>
      </motion.div>
      <h3 className="text-3xl font-bold">{microcopy.onboarding.complete.title}</h3>
      <p className="text-lg text-muted-foreground max-w-md mx-auto">
        {microcopy.onboarding.complete.description}
      </p>
      <Card className="mt-8 max-w-md mx-auto">
        <CardContent className="p-6 space-y-4">
          <div className="flex items-center justify-between py-2 border-b border-border last:border-0">
            <span className="text-sm text-muted-foreground">API Key</span>
            <span className="text-sm font-medium">
              {wizardData.apiKey?.created ? (
                <span className="text-green-600 dark:text-green-500">✓ Created</span>
              ) : (
                <span className="text-muted-foreground">Skipped</span>
              )}
            </span>
          </div>
          <div className="flex items-center justify-between py-2 border-b border-border last:border-0">
            <span className="text-sm text-muted-foreground">Data Source</span>
            <span className="text-sm font-medium">
              {wizardData.dataSource?.connected ? (
                <span className="text-green-600 dark:text-green-500">✓ Connected</span>
              ) : (
                <span className="text-muted-foreground">Skipped</span>
              )}
            </span>
          </div>
          <div className="flex items-center justify-between py-2">
            <span className="text-sm text-muted-foreground">Semantic Collection</span>
            <span className="text-sm font-medium">
              {wizardData.collection?.created ? (
                <span className="text-green-600 dark:text-green-500">✓ Created</span>
              ) : (
                <span className="text-muted-foreground">Skipped</span>
              )}
            </span>
          </div>
        </CardContent>
      </Card>
      <div className="mt-8">
        <Button
          onClick={() => (window.location.href = '/')}
          size="lg"
          className="min-w-[200px]"
        >
          <RocketLaunchIcon className="h-5 w-5 mr-2" />
          {microcopy.onboarding.complete.goToDashboard}
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
  const { markCompleted } = useOnboarding()

  const handleComplete = async (data: OnboardingWizardData) => {
    markCompleted()
    if (onComplete) onComplete()
  }

  const steps: WizardStep[] = [
    {
      id: 'welcome',
      title: microcopy.onboarding.welcome.title,
      description: microcopy.onboarding.welcome.subtitle,
      component: WelcomeStep,
      canSkip: false,
    },
    {
      id: 'api-key',
      title: microcopy.onboarding.apiKey.title,
      description: microcopy.onboarding.apiKey.description,
      component: APIKeyStep,
      canSkip: true,
    },
    {
      id: 'data-source',
      title: microcopy.onboarding.dataSource.title,
      description: microcopy.onboarding.dataSource.optional,
      component: DataSourceStep,
      canSkip: true,
      isOptional: true,
    },
    {
      id: 'collection',
      title: microcopy.onboarding.semanticSearch.title,
      description: microcopy.onboarding.semanticSearch.optional,
      component: CollectionStep,
      canSkip: true,
      isOptional: true,
    },
    {
      id: 'features',
      title: microcopy.onboarding.features.title,
      description: microcopy.onboarding.features.description,
      component: FeaturesStep,
      canSkip: true,
    },
    {
      id: 'complete',
      title: microcopy.onboarding.complete.title,
      description: microcopy.onboarding.complete.description,
      component: CompleteStep,
      canSkip: false,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title={microcopy.onboarding.welcome.title}
      description={microcopy.onboarding.welcome.subtitle}
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

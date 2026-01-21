'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { showToast } from '@/components/ui/Toast'
import HelpText from '@/components/ui/HelpText'
import { useMutation } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'

interface SemanticWizardData {
  name: string
  description: string
  embeddingModel: string
  chunkSize: number
  chunkOverlap: number
}

const embeddingModels = [
  { id: 'text-embedding-ada-002', name: 'OpenAI Ada-002', description: 'Fast and efficient' },
  { id: 'text-embedding-3-small', name: 'OpenAI 3 Small', description: 'Balanced performance' },
  { id: 'text-embedding-3-large', name: 'OpenAI 3 Large', description: 'Highest quality' },
  { id: 'neurondb-default', name: 'NeuronDB Default', description: 'Optimized for NeuronDB' },
]

function WelcomeStep({ data, updateData, nextStep }: WizardStepProps) {
  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Welcome to Semantic Search Setup</h3>
        <p className="text-sm text-muted-foreground">
          Create a collection to start searching your documents semantically using AI.
        </p>
      </div>
      <HelpText
        variant="inline"
        title="What is Semantic Search?"
        content={
          <p>
            Semantic search allows you to find documents by meaning, not just keywords. Powered by
            vector embeddings and AI, it understands context and intent to provide relevant results.
          </p>
        }
        link="/docs/features/semantic-search"
        linkText="Learn more about Semantic Search"
      />
      <div className="pt-4">
        <Button onClick={nextStep} className="w-full sm:w-auto">
          Get Started →
        </Button>
      </div>
    </div>
  )
}

function CollectionInfoStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as SemanticWizardData) || { name: '', description: '' }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Create a collection to organize your documents for semantic search.
      </p>
      <Input
        label="Collection Name"
        value={wizardData.name || ''}
        onChange={(e) => updateData({ name: e.target.value })}
        placeholder="e.g., knowledge_base"
        required
        helperText="A unique name for your collection"
      />
      <Textarea
        label="Description"
        value={wizardData.description || ''}
        onChange={(e) => updateData({ description: e.target.value })}
        placeholder="Describe what documents this collection contains"
        rows={3}
        helperText="Optional description of the collection"
      />
      <HelpText
        variant="inline"
        content={
          <p>
            Collections group related documents together. You can create multiple collections for
            different purposes or data sources.
          </p>
        }
      />
    </div>
  )
}

function EmbeddingConfigStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as SemanticWizardData) || {
    embeddingModel: 'neurondb-default',
    chunkSize: 1000,
    chunkOverlap: 200,
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Configure how documents are embedded and indexed for semantic search.
      </p>
      <div className="space-y-2">
        <label className="text-sm font-medium">Embedding Model</label>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {embeddingModels.map((model) => (
            <Card
              key={model.id}
              hover
              className={
                wizardData.embeddingModel === model.id
                  ? 'ring-2 ring-primary'
                  : ''
              }
              onClick={() => updateData({ embeddingModel: model.id })}
            >
              <CardContent className="p-3">
                <div className="font-medium text-sm">{model.name}</div>
                <div className="text-xs text-muted-foreground mt-1">
                  {model.description}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <Input
          label="Chunk Size"
          value={wizardData.chunkSize?.toString() || '1000'}
          onChange={(e) =>
            updateData({ chunkSize: parseInt(e.target.value) || 1000 })
          }
          type="number"
          helperText="Characters per chunk"
        />
        <Input
          label="Chunk Overlap"
          value={wizardData.chunkOverlap?.toString() || '200'}
          onChange={(e) =>
            updateData({ chunkOverlap: parseInt(e.target.value) || 200 })
          }
          type="number"
          helperText="Overlapping characters"
        />
      </div>
      <HelpText
        variant="inline"
        content={
          <p>
            Documents are split into chunks for indexing. Overlap helps maintain context across
            chunk boundaries. Adjust based on your document size and structure.
          </p>
        }
      />
    </div>
  )
}

function CompleteStep({ data }: WizardStepProps) {
  const wizardData = data as SemanticWizardData
  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Setup Complete!</h3>
        <p className="text-sm text-muted-foreground">
          Your collection "{wizardData.name}" is ready to use.
        </p>
      </div>
      <div className="rounded-lg border p-4 bg-muted/50 space-y-2">
        <p className="text-sm font-medium">What's next?</p>
        <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
          <li>Upload documents to your collection</li>
          <li>Start searching semantically</li>
          <li>Try the RAG (Retrieval-Augmented Generation) feature</li>
          <li>Monitor search performance and analytics</li>
        </ul>
      </div>
      <HelpText
        variant="inline"
        content={
          <p>
            Upload documents and start asking questions. The system will find relevant content based
            on meaning, not just keywords.
          </p>
        }
      />
    </div>
  )
}

interface SemanticSearchSetupWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function SemanticSearchSetupWizard({
  onComplete,
  onCancel,
}: SemanticSearchSetupWizardProps) {
  const createCollection = useMutation({
    mutationFn: async (data: SemanticWizardData) => {
      const response = await apiClient.post('/semantic/collections', {
        name: data.name,
        description: data.description,
        config: {
          embedding_model: data.embeddingModel,
          chunk_size: data.chunkSize,
          chunk_overlap: data.chunkOverlap,
        },
      })
      return response.data
    },
  })

  const handleComplete = async (data: SemanticWizardData) => {
    try {
      await createCollection.mutateAsync(data)
      showToast('Semantic search collection created successfully!', 'success')
      onComplete?.()
    } catch (error: any) {
      showToast(
        error?.message || 'Failed to create collection. Please try again.',
        'error'
      )
    }
  }

  const steps: WizardStep[] = [
    {
      id: 'welcome',
      title: 'Welcome',
      description: 'Introduction to Semantic Search',
      component: WelcomeStep,
      canSkip: false,
    },
    {
      id: 'collection',
      title: 'Collection Information',
      description: 'Name and describe your collection',
      component: CollectionInfoStep,
      validate: async (data) => {
        const wizardData = data as SemanticWizardData
        return !!(wizardData.name && wizardData.name.trim().length > 0)
      },
    },
    {
      id: 'embedding',
      title: 'Embedding Configuration',
      description: 'Configure document embedding settings',
      component: EmbeddingConfigStep,
      validate: async (data) => {
        const wizardData = data as SemanticWizardData
        return !!(
          wizardData.embeddingModel &&
          wizardData.chunkSize > 0 &&
          wizardData.chunkOverlap >= 0
        )
      },
    },
    {
      id: 'complete',
      title: 'Complete',
      description: "You're all set!",
      component: CompleteStep,
      canSkip: false,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Semantic Search Setup Wizard"
      description="Create a collection for semantic search in a few simple steps"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

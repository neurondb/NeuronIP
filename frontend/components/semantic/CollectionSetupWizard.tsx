'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { showToast } from '@/components/ui/Toast'

interface CollectionWizardData {
  name: string
  description: string
  model: string
  chunking: {
    strategy: string
    chunkSize: number
    overlap: number
  }
  documents: {
    uploaded: boolean
    source?: string
  }
}

const embeddingModels = [
  { id: 'text-embedding-ada-002', name: 'OpenAI Ada', description: 'Fast and efficient' },
  { id: 'text-embedding-3-small', name: 'OpenAI Small', description: 'Balanced performance' },
  { id: 'text-embedding-3-large', name: 'OpenAI Large', description: 'Highest quality' },
]

const chunkingStrategies = [
  { id: 'fixed', name: 'Fixed Size', description: 'Split into fixed-size chunks' },
  { id: 'sentence', name: 'Sentence-based', description: 'Split by sentences' },
  { id: 'paragraph', name: 'Paragraph-based', description: 'Split by paragraphs' },
]

function NameStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as CollectionWizardData) || { name: '', description: '' }

  return (
    <div className="space-y-4">
      <div>
        <Input
          label="Collection Name *"
          value={wizardData.name || ''}
          onChange={(e) => updateData({ name: e.target.value })}
          placeholder="e.g., Knowledge Base"
          required
        />
      </div>
      <div>
        <Textarea
          label="Description"
          value={wizardData.description || ''}
          onChange={(e) => updateData({ description: e.target.value })}
          placeholder="Describe what this collection contains..."
          rows={4}
        />
      </div>
    </div>
  )
}

function ModelStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as CollectionWizardData) || { model: '' }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Choose the embedding model for this collection.
      </p>
      <div className="space-y-3">
        {embeddingModels.map((model) => (
          <Card
            key={model.id}
            hover
            className={wizardData.model === model.id ? 'ring-2 ring-primary' : ''}
            onClick={() => updateData({ model: model.id })}
          >
            <CardContent className="p-4">
              <h3 className="font-semibold mb-1">{model.name}</h3>
              <p className="text-sm text-muted-foreground">{model.description}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

function ChunkingStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as CollectionWizardData) || {
    chunking: { strategy: 'fixed', chunkSize: 1000, overlap: 200 },
  }
  const [strategy, setStrategy] = useState(wizardData.chunking?.strategy || 'fixed')
  const [chunkSize, setChunkSize] = useState(wizardData.chunking?.chunkSize || 1000)
  const [overlap, setOverlap] = useState(wizardData.chunking?.overlap || 200)

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium mb-2">Chunking Strategy</label>
        <select
          value={strategy}
          onChange={(e) => {
            setStrategy(e.target.value)
            updateData({
              chunking: { ...wizardData.chunking, strategy: e.target.value },
            })
          }}
          className="w-full px-3 py-2 border rounded-lg"
        >
          {chunkingStrategies.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name} - {s.description}
            </option>
          ))}
        </select>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <Input
          label="Chunk Size (characters)"
          type="number"
          value={chunkSize}
          onChange={(e) => {
            const value = parseInt(e.target.value) || 1000
            setChunkSize(value)
            updateData({
              chunking: { ...wizardData.chunking, chunkSize: value },
            })
          }}
          min={100}
          max={10000}
        />
        <Input
          label="Overlap (characters)"
          type="number"
          value={overlap}
          onChange={(e) => {
            const value = parseInt(e.target.value) || 200
            setOverlap(value)
            updateData({
              chunking: { ...wizardData.chunking, overlap: value },
            })
          }}
          min={0}
          max={1000}
        />
      </div>
    </div>
  )
}

function DocumentsStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as CollectionWizardData) || { documents: { uploaded: false } }
  const [file, setFile] = useState<File | null>(null)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0]
    if (selectedFile) {
      setFile(selectedFile)
      updateData({
        documents: { uploaded: true, source: selectedFile.name },
      })
    }
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Upload initial documents or connect a data source. You can add more documents later.
      </p>
      <div>
        <label className="block text-sm font-medium mb-2">Upload Document</label>
        <input
          type="file"
          onChange={handleFileChange}
          accept=".txt,.pdf,.md,.docx"
          className="w-full px-3 py-2 border rounded-lg"
        />
        {file && (
          <p className="text-sm text-muted-foreground mt-2">Selected: {file.name}</p>
        )}
      </div>
      <Button
        variant="outline"
        onClick={() => {
          updateData({
            documents: { uploaded: false, source: 'data-source' },
          })
        }}
      >
        Connect Data Source Instead
      </Button>
    </div>
  )
}

function ReviewStep({ data }: WizardStepProps) {
  const wizardData = data as CollectionWizardData

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="p-4 space-y-3">
          <div>
            <h4 className="font-semibold mb-1">Name</h4>
            <p className="text-sm">{wizardData.name || 'Not set'}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Embedding Model</h4>
            <p className="text-sm">{wizardData.model || 'Not set'}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Chunking Strategy</h4>
            <p className="text-sm capitalize">{wizardData.chunking?.strategy || 'Not set'}</p>
            <p className="text-xs text-muted-foreground">
              Size: {wizardData.chunking?.chunkSize || 1000} chars, Overlap:{' '}
              {wizardData.chunking?.overlap || 200} chars
            </p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Documents</h4>
            <p className="text-sm">
              {wizardData.documents?.uploaded
                ? `Uploaded: ${wizardData.documents?.source || 'Yes'}`
                : 'None uploaded'}
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

interface CollectionSetupWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function CollectionSetupWizard({
  onComplete,
  onCancel,
}: CollectionSetupWizardProps) {
  const handleComplete = async (data: CollectionWizardData) => {
    // In real implementation, call API to create collection
    showToast('Collection created successfully', 'success')
    if (onComplete) onComplete()
  }

  const validateName = (data: any): boolean => {
    return !!(data as CollectionWizardData).name?.trim()
  }

  const validateModel = (data: any): boolean => {
    return !!(data as CollectionWizardData).model
  }

  const steps: WizardStep[] = [
    {
      id: 'name',
      title: 'Name & Description',
      description: 'Give your collection a name',
      component: NameStep,
      validate: validateName,
    },
    {
      id: 'model',
      title: 'Choose Embedding Model',
      description: 'Select the model for generating embeddings',
      component: ModelStep,
      validate: validateModel,
    },
    {
      id: 'chunking',
      title: 'Configure Chunking',
      description: 'Set up how documents are split',
      component: ChunkingStep,
    },
    {
      id: 'documents',
      title: 'Upload Documents',
      description: 'Add initial documents to the collection',
      component: DocumentsStep,
      canSkip: true,
    },
    {
      id: 'review',
      title: 'Review & Create',
      description: 'Review your collection configuration',
      component: ReviewStep,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Create Semantic Collection"
      description="Set up a new knowledge base collection"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

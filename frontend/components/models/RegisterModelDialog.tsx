'use client'

import { useState } from 'react'
import Modal from '@/components/ui/Modal'
import Button from '@/components/ui/Button'
import { useRegisterModel } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface RegisterModelDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function RegisterModelDialog({ open, onOpenChange }: RegisterModelDialogProps) {
  const [name, setName] = useState('')
  const [modelType, setModelType] = useState('')
  const [endpoint, setEndpoint] = useState('')
  const { mutate: registerModel, isPending } = useRegisterModel()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !modelType) {
      showToast('Name and model type are required', 'warning')
      return
    }

    registerModel(
      {
        name,
        model_type: modelType,
        endpoint: endpoint || undefined,
      },
      {
        onSuccess: () => {
          showToast('Model registered successfully', 'success')
          onOpenChange(false)
          setName('')
          setModelType('')
          setEndpoint('')
        },
        onError: () => {
          showToast('Failed to register model', 'error')
        },
      }
    )
  }

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Register New Model"
      description="Register a new AI model for inference"
      size="md"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="text-sm font-medium mb-2 block">Model Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., GPT-4"
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">Model Type</label>
          <select
            value={modelType}
            onChange={(e) => setModelType(e.target.value)}
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          >
            <option value="">Select type...</option>
            <option value="language-model">Language Model</option>
            <option value="embedding">Embedding Model</option>
            <option value="classifier">Classifier</option>
            <option value="other">Other</option>
          </select>
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">Endpoint (Optional)</label>
          <input
            type="url"
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            placeholder="https://api.example.com/v1/models"
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={isPending}>
            Register Model
          </Button>
        </div>
      </form>
    </Modal>
  )
}

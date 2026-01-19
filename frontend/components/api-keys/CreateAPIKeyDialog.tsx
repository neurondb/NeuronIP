'use client'

import { useState } from 'react'
import Modal from '@/components/ui/Modal'
import Button from '@/components/ui/Button'
import { useCreateAPIKey } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface CreateAPIKeyDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated?: (apiKey: { id: string; key: string }) => void
}

export default function CreateAPIKeyDialog({
  open,
  onOpenChange,
  onCreated,
}: CreateAPIKeyDialogProps) {
  const [name, setName] = useState('')
  const [rateLimit, setRateLimit] = useState(100)
  const { mutate: createAPIKey, isPending } = useCreateAPIKey()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) {
      showToast('Name is required', 'warning')
      return
    }

    createAPIKey(
      {
        name: name.trim(),
        rate_limit: rateLimit || 100,
      },
      {
        onSuccess: (data) => {
          showToast('API key created successfully', 'success')
          onCreated?.(data)
          onOpenChange(false)
          setName('')
          setRateLimit(100)
        },
        onError: () => {
          showToast('Failed to create API key', 'error')
        },
      }
    )
  }

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Create API Key"
      description="Generate a new API key for authentication"
      size="md"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="text-sm font-medium mb-2 block">Key Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., Production API Key"
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">Rate Limit (per hour)</label>
          <input
            type="number"
            value={rateLimit}
            onChange={(e) => setRateLimit(parseInt(e.target.value) || 100)}
            min={1}
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={isPending}>
            Create API Key
          </Button>
        </div>
      </form>
    </Modal>
  )
}

'use client'

import { useState } from 'react'
import Modal from '@/components/ui/Modal'
import Button from '@/components/ui/Button'
import { useCreateAgent } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface CreateAgentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated?: () => void
}

export default function CreateAgentDialog({
  open,
  onOpenChange,
  onCreated,
}: CreateAgentDialogProps) {
  const [name, setName] = useState('')
  const [agentType, setAgentType] = useState('')
  const [description, setDescription] = useState('')
  const createMutation = useCreateAgent()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name || !agentType) {
      showToast('Name and agent type are required', 'warning')
      return
    }

    const agentData: Record<string, unknown> = {
      name,
      agent_type: agentType,
      status: 'draft',
    }

    if (description) {
      agentData.config = {
        description,
      }
    }

    createMutation.mutate(agentData, {
      onSuccess: () => {
        showToast('Agent created successfully', 'success')
        onOpenChange(false)
        setName('')
        setAgentType('')
        setDescription('')
        onCreated?.()
      },
      onError: (error: any) => {
        showToast(
          error?.response?.data?.message || 'Failed to create agent',
          'error'
        )
      },
    })
  }

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Create New Agent"
      description="Create a new AI agent for your workflows"
      size="md"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="text-sm font-medium mb-2 block">
            Agent Name <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., Customer Support Agent"
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">
            Agent Type <span className="text-red-500">*</span>
          </label>
          <select
            value={agentType}
            onChange={(e) => setAgentType(e.target.value)}
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          >
            <option value="">Select type...</option>
            <option value="workflow">Workflow</option>
            <option value="support">Support</option>
            <option value="analytics">Analytics</option>
            <option value="automation">Automation</option>
            <option value="custom">Custom</option>
          </select>
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">
            Description (Optional)
          </label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe the agent's purpose and capabilities..."
            rows={3}
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring resize-none"
          />
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={createMutation.isPending}>
            {createMutation.isPending ? 'Creating...' : 'Create Agent'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}

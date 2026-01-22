'use client'

import { useState } from 'react'
import Modal from '@/components/ui/Modal'
import Input from '@/components/ui/Input'
import Select from '@/components/ui/Select'
import { Textarea } from '@/components/ui/Textarea'
import Button from '@/components/ui/Button'
import { useCreateSupportTicket } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface CreateTicketDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export default function CreateTicketDialog({ open, onOpenChange, onSuccess }: CreateTicketDialogProps) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [priority, setPriority] = useState<'low' | 'medium' | 'high'>('medium')
  const [errors, setErrors] = useState<Record<string, string>>({})

  const { mutate: createTicket, isPending } = useCreateSupportTicket()

  const validate = () => {
    const newErrors: Record<string, string> = {}

    if (!title.trim()) {
      newErrors.title = 'Title is required'
    } else if (title.trim().length < 3) {
      newErrors.title = 'Title must be at least 3 characters'
    }

    if (!description.trim()) {
      newErrors.description = 'Description is required'
    } else if (description.trim().length < 10) {
      newErrors.description = 'Description must be at least 10 characters'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) {
      return
    }

    createTicket(
      {
        title: title.trim(),
        description: description.trim(),
        priority,
        status: 'open',
      },
      {
        onSuccess: (data) => {
          showToast('Ticket created successfully', 'success')
          // Reset form
          setTitle('')
          setDescription('')
          setPriority('medium')
          setErrors({})
          onOpenChange(false)
          onSuccess?.()
        },
        onError: (error: any) => {
          const errorMessage = error?.response?.data?.error?.message || error?.message || 'Failed to create ticket'
          showToast(errorMessage, 'error')
          
          // Set field-specific errors if available
          if (error?.response?.data?.error?.details) {
            const details = error.response.data.error.details
            const fieldErrors: Record<string, string> = {}
            Object.keys(details).forEach((key) => {
              if (key !== 'request_id') {
                fieldErrors[key] = details[key]
              }
            })
            if (Object.keys(fieldErrors).length > 0) {
              setErrors(fieldErrors)
            }
          }
        },
      }
    )
  }

  const handleClose = () => {
    if (!isPending) {
      setTitle('')
      setDescription('')
      setPriority('medium')
      setErrors({})
      onOpenChange(false)
    }
  }

  return (
    <Modal
      open={open}
      onOpenChange={handleClose}
      title="Create Support Ticket"
      description="Create a new support ticket to get help with your issue"
      size="lg"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Title"
          value={title}
          onChange={(e) => {
            setTitle(e.target.value)
            if (errors.title) {
              setErrors((prev) => ({ ...prev, title: '' }))
            }
          }}
          placeholder="Brief description of your issue"
          required
          error={errors.title}
          disabled={isPending}
        />

        <Select
          label="Priority"
          value={priority}
          onChange={(e) => setPriority(e.target.value as 'low' | 'medium' | 'high')}
          options={[
            { value: 'low', label: 'Low' },
            { value: 'medium', label: 'Medium' },
            { value: 'high', label: 'High' },
          ]}
          disabled={isPending}
        />

        <Textarea
          label="Description"
          value={description}
          onChange={(e) => {
            setDescription(e.target.value)
            if (errors.description) {
              setErrors((prev) => ({ ...prev, description: '' }))
            }
          }}
          placeholder="Provide detailed information about your issue..."
          required
          rows={6}
          error={errors.description}
          disabled={isPending}
        />

        <div className="flex justify-end gap-2 pt-4">
          <Button type="button" variant="outline" onClick={handleClose} disabled={isPending}>
            Cancel
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? 'Creating...' : 'Create Ticket'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}

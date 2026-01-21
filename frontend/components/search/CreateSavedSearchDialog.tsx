'use client'

import { useState } from 'react'
import Modal from '@/components/ui/Modal'
import Button from '@/components/ui/Button'
import { showToast } from '@/components/ui/Toast'
import { createSavedSearch } from '@/lib/api/saved-searches'

interface CreateSavedSearchDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated?: () => void
  initialQuery?: string
}

export default function CreateSavedSearchDialog({
  open,
  onOpenChange,
  onCreated,
  initialQuery = '',
}: CreateSavedSearchDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [query, setQuery] = useState(initialQuery)
  const [tags, setTags] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [isPending, setIsPending] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !query.trim()) {
      showToast('Name and query are required', 'warning')
      return
    }

    setIsPending(true)
    try {
      await createSavedSearch({
        name: name.trim(),
        description: description.trim() || undefined,
        query: query.trim(),
        tags: tags
          .split(',')
          .map((t) => t.trim())
          .filter((t) => t.length > 0),
        is_public: isPublic,
      })
      showToast('Saved search created successfully', 'success')
      onCreated?.()
      onOpenChange(false)
      setName('')
      setDescription('')
      setQuery('')
      setTags('')
      setIsPublic(false)
    } catch (error: any) {
      showToast(error?.response?.data?.message || 'Failed to create saved search', 'error')
    } finally {
      setIsPending(false)
    }
  }

  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Create Saved Search"
      description="Save a query for quick access later"
      size="lg"
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="text-sm font-medium mb-2 block">Name *</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="My Query"
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            required
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Brief description of this query"
            rows={2}
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring resize-none"
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">Query *</label>
          <textarea
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="SELECT * FROM table WHERE..."
            rows={4}
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-ring resize-none"
            required
          />
        </div>

        <div>
          <label className="text-sm font-medium mb-2 block">Tags (comma-separated)</label>
          <input
            type="text"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="analytics, reports, monthly"
            className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="isPublic"
            checked={isPublic}
            onChange={(e) => setIsPublic(e.target.checked)}
            className="rounded border-border"
          />
          <label htmlFor="isPublic" className="text-sm font-medium cursor-pointer">
            Make this search public (visible to all users)
          </label>
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" disabled={isPending}>
            {isPending ? 'Creating...' : 'Create Saved Search'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}

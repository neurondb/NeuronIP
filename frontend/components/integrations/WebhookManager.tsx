'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'
import { showToast } from '@/components/ui/Toast'
import { PlusIcon, TrashIcon } from '@heroicons/react/24/outline'

export default function WebhookManager() {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const queryClient = useQueryClient()

  const { data: webhooks, isLoading } = useQuery({
    queryKey: ['webhooks'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.webhooks)
      return response.data
    },
  })

  const deleteWebhook = useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.delete(API_ENDPOINTS.webhook(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
      showToast('Webhook deleted successfully', 'success')
    },
  })

  const triggerWebhook = useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.post(API_ENDPOINTS.webhookTrigger(id), {})
      return response.data
    },
    onSuccess: () => {
      showToast('Webhook triggered successfully', 'success')
    },
  })

  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const webhookData = {
      name: formData.get('name'),
      url: formData.get('url'),
      events: (formData.get('events') as string).split(',').map((e) => e.trim()),
      enabled: formData.get('enabled') === 'on',
    }

    try {
      await apiClient.post(API_ENDPOINTS.webhooks, webhookData)
      showToast('Webhook created successfully', 'success')
      setShowCreateForm(false)
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
    } catch (error: any) {
      showToast('Failed to create webhook', 'error')
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Webhooks</CardTitle>
            <CardDescription>Manage webhook endpoints for event notifications</CardDescription>
          </div>
          <Button onClick={() => setShowCreateForm(!showCreateForm)} size="sm">
            <PlusIcon className="h-4 w-4 mr-1" />
            New Webhook
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {showCreateForm && (
          <form onSubmit={handleCreate} className="p-4 rounded-lg border border-border space-y-3">
            <div>
              <label className="text-sm font-medium mb-2 block">Name</label>
              <input
                type="text"
                name="name"
                required
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">URL</label>
              <input
                type="url"
                name="url"
                required
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Events (comma-separated)</label>
              <input
                type="text"
                name="events"
                placeholder="query.completed, workflow.started, *"
                required
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div className="flex items-center gap-2">
              <input type="checkbox" id="enabled" name="enabled" defaultChecked />
              <label htmlFor="enabled" className="text-sm font-medium">
                Enabled
              </label>
            </div>
            <div className="flex gap-2 justify-end">
              <Button type="button" variant="outline" onClick={() => setShowCreateForm(false)}>
                Cancel
              </Button>
              <Button type="submit">Create</Button>
            </div>
          </form>
        )}

        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading webhooks...</div>
        ) : !webhooks || (Array.isArray(webhooks) && webhooks.length === 0) ? (
          <div className="text-center py-8 text-muted-foreground">No webhooks configured</div>
        ) : (
          <div className="space-y-2">
            {(Array.isArray(webhooks) ? webhooks : []).map((webhook: any) => (
              <div key={webhook.id} className="p-4 rounded-lg border border-border">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="font-medium">{webhook.name}</div>
                    <div className="text-sm text-muted-foreground">{webhook.url}</div>
                    <div className="text-xs text-muted-foreground mt-1">
                      Events: {webhook.events?.join(', ') || 'none'}
                    </div>
                    {webhook.last_triggered_at && (
                      <div className="text-xs text-muted-foreground">
                        Last triggered: {new Date(webhook.last_triggered_at).toLocaleString()}
                      </div>
                    )}
                  </div>
                  <div className="flex gap-2">
                    <Button
                      onClick={() => triggerWebhook.mutate(webhook.id)}
                      variant="outline"
                      size="sm"
                      disabled={triggerWebhook.isPending}
                    >
                      Test
                    </Button>
                    <Button
                      onClick={() => deleteWebhook.mutate(webhook.id)}
                      variant="ghost"
                      size="sm"
                      disabled={deleteWebhook.isPending}
                    >
                      <TrashIcon className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

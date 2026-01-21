'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { API_ENDPOINTS } from '@/lib/utils/constants'
import { showToast } from '@/components/ui/Toast'
import { PlusIcon, TrashIcon, PencilIcon } from '@heroicons/react/24/outline'

export default function PolicyManager() {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [editingPolicy, setEditingPolicy] = useState<any>(null)
  const queryClient = useQueryClient()

  const { data: policies, isLoading } = useQuery({
    queryKey: ['compliance-policies'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.compliancePolicies)
      return response.data
    },
  })

  const deletePolicy = useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.delete(API_ENDPOINTS.compliancePolicy(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['compliance-policies'] })
      showToast('Policy deleted successfully', 'success')
    },
  })

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const policyData = {
      policy_name: formData.get('name'),
      policy_type: formData.get('type'),
      policy_text: formData.get('text'),
      enabled: formData.get('enabled') === 'on',
      rules: [],
    }

    try {
      if (editingPolicy) {
        await apiClient.put(API_ENDPOINTS.compliancePolicy(editingPolicy.id), policyData)
        showToast('Policy updated successfully', 'success')
      } else {
        await apiClient.post(API_ENDPOINTS.compliancePolicies, policyData)
        showToast('Policy created successfully', 'success')
      }
      setShowCreateForm(false)
      setEditingPolicy(null)
      queryClient.invalidateQueries({ queryKey: ['compliance-policies'] })
    } catch (error: any) {
      showToast('Failed to save policy', 'error')
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Compliance Policies</CardTitle>
            <CardDescription>Manage compliance policies and rules</CardDescription>
          </div>
          <Button onClick={() => setShowCreateForm(!showCreateForm)} size="sm">
            <PlusIcon className="h-4 w-4 mr-1" />
            New Policy
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {(showCreateForm || editingPolicy) && (
          <form onSubmit={handleSubmit} className="p-4 rounded-lg border border-border space-y-3">
            <div>
              <label className="text-sm font-medium mb-2 block">Policy Name *</label>
              <input
                type="text"
                name="name"
                defaultValue={editingPolicy?.policy_name}
                required
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Policy Type *</label>
              <select
                name="type"
                defaultValue={editingPolicy?.policy_type}
                required
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="">Select type...</option>
                <option value="data_privacy">Data Privacy</option>
                <option value="data_retention">Data Retention</option>
                <option value="access_control">Access Control</option>
                <option value="data_quality">Data Quality</option>
                <option value="custom">Custom</option>
              </select>
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Policy Text *</label>
              <textarea
                name="text"
                defaultValue={editingPolicy?.policy_text}
                rows={6}
                required
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring resize-none"
                placeholder="Enter policy description and rules..."
              />
            </div>
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="enabled"
                name="enabled"
                defaultChecked={editingPolicy?.enabled !== false}
                className="rounded border-border"
              />
              <label htmlFor="enabled" className="text-sm font-medium">
                Enabled
              </label>
            </div>
            <div className="flex gap-2 justify-end">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setShowCreateForm(false)
                  setEditingPolicy(null)
                }}
              >
                Cancel
              </Button>
              <Button type="submit">{editingPolicy ? 'Update' : 'Create'}</Button>
            </div>
          </form>
        )}

        {isLoading ? (
          <div className="text-center py-8 text-muted-foreground">Loading policies...</div>
        ) : !policies || (Array.isArray(policies) && policies.length === 0) ? (
          <div className="text-center py-8 text-muted-foreground">No policies configured</div>
        ) : (
          <div className="space-y-2">
            {(Array.isArray(policies) ? policies : []).map((policy: any) => (
              <div key={policy.id} className="p-4 rounded-lg border border-border">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium">{policy.policy_name}</h3>
                      <span className="text-xs px-2 py-1 rounded bg-primary/10 text-primary">
                        {policy.policy_type}
                      </span>
                      {policy.enabled ? (
                        <span className="text-xs px-2 py-1 rounded bg-green-500/10 text-green-600">
                          Enabled
                        </span>
                      ) : (
                        <span className="text-xs px-2 py-1 rounded bg-gray-500/10 text-gray-600">
                          Disabled
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground mt-2">{policy.policy_text}</p>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      onClick={() => {
                        setEditingPolicy(policy)
                        setShowCreateForm(true)
                      }}
                      variant="ghost"
                      size="sm"
                    >
                      <PencilIcon className="h-4 w-4" />
                    </Button>
                    <Button
                      onClick={() => deletePolicy.mutate(policy.id)}
                      variant="ghost"
                      size="sm"
                      disabled={deletePolicy.isPending}
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

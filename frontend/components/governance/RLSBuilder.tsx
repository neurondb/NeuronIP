'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { Textarea } from '@/components/ui/Textarea'
import Input from '@/components/ui/Input'
import { showToast } from '@/components/ui/Toast'

interface RLSPolicy {
  id: string
  connector: string
  schema: string
  table: string
  policy_name: string
  filter_expression: string
  user_roles: string[]
}

export default function RLSBuilder() {
  const [selectedPolicy, setSelectedPolicy] = useState<RLSPolicy | null>(null)
  const [formData, setFormData] = useState({
    connector: '',
    schema: '',
    table: '',
    policy_name: '',
    filter_expression: '',
    user_roles: '',
  })
  const queryClient = useQueryClient()

  // Fetch RLS policies
  const { data: policies, isLoading } = useQuery({
    queryKey: ['rls-policies'],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/governance/rls/policies')
      return response.data as RLSPolicy[]
    },
  })

  // Create policy mutation
  const createPolicyMutation = useMutation({
    mutationFn: async (policy: Partial<RLSPolicy>) => {
      return apiClient.post('/api/v1/governance/rls/policies', policy)
    },
    onSuccess: () => {
      showToast('RLS policy created', 'success')
      queryClient.invalidateQueries({ queryKey: ['rls-policies'] })
      setFormData({
        connector: '',
        schema: '',
        table: '',
        policy_name: '',
        filter_expression: '',
        user_roles: '',
      })
    },
  })

  const handleSubmit = () => {
    if (!formData.connector || !formData.schema || !formData.table || !formData.filter_expression) {
      showToast('Please fill in all required fields', 'error')
      return
    }

    createPolicyMutation.mutate({
      connector: formData.connector,
      schema: formData.schema,
      table: formData.table,
      policy_name: formData.policy_name || `policy_${Date.now()}`,
      filter_expression: formData.filter_expression,
      user_roles: formData.user_roles ? formData.user_roles.split(',').map(r => r.trim()) : [],
    })
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <Card>
        <CardHeader>
          <CardTitle>RLS Policies</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading...</div>
          ) : !policies || policies.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">No RLS policies</div>
          ) : (
            <div className="space-y-3">
              {policies.map((policy) => (
                <div
                  key={policy.id}
                  className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                    selectedPolicy?.id === policy.id
                      ? 'border-primary bg-primary/5'
                      : 'hover:bg-muted/50'
                  }`}
                  onClick={() => setSelectedPolicy(policy)}
                >
                  <h4 className="font-semibold">{policy.policy_name}</h4>
                  <p className="text-sm text-muted-foreground">
                    {policy.connector}.{policy.schema}.{policy.table}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {policy.filter_expression}
                  </p>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Create RLS Policy</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <label className="text-sm font-medium">Connector</label>
            <Input
              value={formData.connector}
              onChange={(e) => setFormData({ ...formData, connector: e.target.value })}
              placeholder="connector_id"
              className="mt-1"
            />
          </div>

          <div>
            <label className="text-sm font-medium">Schema</label>
            <Input
              value={formData.schema}
              onChange={(e) => setFormData({ ...formData, schema: e.target.value })}
              placeholder="public"
              className="mt-1"
            />
          </div>

          <div>
            <label className="text-sm font-medium">Table</label>
            <Input
              value={formData.table}
              onChange={(e) => setFormData({ ...formData, table: e.target.value })}
              placeholder="users"
              className="mt-1"
            />
          </div>

          <div>
            <label className="text-sm font-medium">Policy Name</label>
            <Input
              value={formData.policy_name}
              onChange={(e) => setFormData({ ...formData, policy_name: e.target.value })}
              placeholder="user_data_access"
              className="mt-1"
            />
          </div>

          <div>
            <label className="text-sm font-medium">Filter Expression</label>
            <Textarea
              value={formData.filter_expression}
              onChange={(e) => setFormData({ ...formData, filter_expression: e.target.value })}
              placeholder="user_id = current_user_id() OR role = 'admin'"
              className="mt-1 font-mono text-sm"
              rows={4}
            />
            <p className="text-xs text-muted-foreground mt-1">
              SQL expression that filters rows based on user context
            </p>
          </div>

          <div>
            <label className="text-sm font-medium">User Roles (comma-separated)</label>
            <Input
              value={formData.user_roles}
              onChange={(e) => setFormData({ ...formData, user_roles: e.target.value })}
              placeholder="admin, analyst, viewer"
              className="mt-1"
            />
          </div>

          <Button
            onClick={handleSubmit}
            disabled={createPolicyMutation.isPending}
            className="w-full"
          >
            {createPolicyMutation.isPending ? 'Creating...' : 'Create Policy'}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}

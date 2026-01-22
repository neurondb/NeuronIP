'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { showToast } from '@/components/ui/Toast'
import { Badge } from '@/components/ui/Badge'

interface ResourceQuota {
  id: string
  workspace_id?: string
  user_id?: string
  resource_type: 'queries' | 'agents' | 'storage' | 'compute' | 'api_calls'
  max_limit: number
  current_usage: number
  period: 'hour' | 'day' | 'month'
  reset_at: string
  enabled: boolean
}

interface ResourceQuotasProps {
  workspaceId?: string
  userId?: string
}

export default function ResourceQuotas({ workspaceId, userId }: ResourceQuotasProps) {
  const [quotas, setQuotas] = useState<ResourceQuota[]>([])
  const [loading, setLoading] = useState(false)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newQuota, setNewQuota] = useState({
    resource_type: 'queries' as const,
    max_limit: 1000,
    period: 'day' as const,
  })

  useEffect(() => {
    fetchQuotas()
  }, [workspaceId, userId])

  const fetchQuotas = async () => {
    try {
      const params = new URLSearchParams()
      if (workspaceId) params.append('workspace_id', workspaceId)
      if (userId) params.append('user_id', userId)

      const response = await fetch(`/api/v1/admin/resource-quotas?${params.toString()}`)
      if (!response.ok) throw new Error('Failed to fetch quotas')
      const data = await response.json()
      setQuotas(data)
    } catch (error) {
      console.error('Error fetching quotas:', error)
    }
  }

  const handleCreateQuota = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/admin/resource-quotas', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          workspace_id: workspaceId || null,
          user_id: userId || null,
          ...newQuota,
        }),
      })

      if (!response.ok) throw new Error('Failed to create quota')
      
      showToast('Resource quota created', 'success')
      setShowCreateForm(false)
      fetchQuotas()
    } catch (error: any) {
      showToast(error.message || 'Failed to create quota', 'error')
    } finally {
      setLoading(false)
    }
  }

  const getUsagePercentage = (quota: ResourceQuota) => {
    if (quota.max_limit === 0) return 0
    return Math.min((quota.current_usage / quota.max_limit) * 100, 100)
  }

  const getUsageColor = (percentage: number) => {
    if (percentage >= 90) return 'error'
    if (percentage >= 70) return 'warning'
    return 'success'
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Resource Quotas</CardTitle>
              <CardDescription>
                Monitor and manage resource usage limits
                {workspaceId && ` for workspace ${workspaceId.substring(0, 8)}...`}
                {userId && ` for user ${userId}`}
              </CardDescription>
            </div>
            <Button onClick={() => setShowCreateForm(!showCreateForm)} size="sm">
              {showCreateForm ? 'Cancel' : 'Add Quota'}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {showCreateForm && (
            <div className="mb-6 p-4 border rounded-lg space-y-3">
              <h4 className="font-medium">Create Resource Quota</h4>
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="text-sm font-medium mb-1 block">Resource Type</label>
                  <select
                    value={newQuota.resource_type}
                    onChange={(e) => setNewQuota({ ...newQuota, resource_type: e.target.value as any })}
                    className="w-full rounded-lg border border-border bg-background px-4 py-2"
                  >
                    <option value="queries">Queries</option>
                    <option value="agents">Agents</option>
                    <option value="storage">Storage</option>
                    <option value="compute">Compute</option>
                    <option value="api_calls">API Calls</option>
                  </select>
                </div>
                <div>
                  <label className="text-sm font-medium mb-1 block">Max Limit</label>
                  <Input
                    type="number"
                    value={newQuota.max_limit}
                    onChange={(e) => setNewQuota({ ...newQuota, max_limit: parseInt(e.target.value) || 0 })}
                    placeholder="1000"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-1 block">Period</label>
                  <select
                    value={newQuota.period}
                    onChange={(e) => setNewQuota({ ...newQuota, period: e.target.value as any })}
                    className="w-full rounded-lg border border-border bg-background px-4 py-2"
                  >
                    <option value="hour">Hour</option>
                    <option value="day">Day</option>
                    <option value="month">Month</option>
                  </select>
                </div>
              </div>
              <Button onClick={handleCreateQuota} disabled={loading}>
                Create Quota
              </Button>
            </div>
          )}

          {quotas.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No resource quotas configured
            </div>
          ) : (
            <div className="space-y-4">
              {quotas.map((quota) => {
                const usagePercentage = getUsagePercentage(quota)
                const usageColor = getUsageColor(usagePercentage)

                return (
                  <motion.div
                    key={quota.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="border rounded-lg p-4"
                  >
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-3">
                        <Badge variant={usageColor}>{quota.resource_type}</Badge>
                        <span className="text-sm text-muted-foreground">
                          {quota.current_usage.toLocaleString()} / {quota.max_limit.toLocaleString()} {quota.period}ly
                        </span>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        Resets: {new Date(quota.reset_at).toLocaleString()}
                      </div>
                    </div>
                    <div className="w-full bg-muted rounded-full h-2 mt-2">
                      <div
                        className={`h-2 rounded-full ${
                          usageColor === 'error'
                            ? 'bg-red-500'
                            : usageColor === 'warning'
                            ? 'bg-yellow-500'
                            : 'bg-green-500'
                        }`}
                        style={{ width: `${Math.min(usagePercentage, 100)}%` }}
                      />
                    </div>
                    <div className="mt-2 text-xs text-muted-foreground">
                      {usagePercentage.toFixed(1)}% used
                    </div>
                  </motion.div>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

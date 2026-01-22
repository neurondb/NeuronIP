'use client'

import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { Textarea } from '@/components/ui/Textarea'
import { Badge } from '@/components/ui/Badge'
import { showToast } from '@/components/ui/Toast'
import { formatDistanceToNow } from 'date-fns'

interface ApprovalItem {
  approval_id: string
  metric_id: string
  metric_name: string
  metric_display_name: string
  approver_id: string
  approval_status: string
  comments?: string
  requested_at: string
  updated_at: string
  metric_status: string
  primary_owner_id?: string
}

interface MetricOwner {
  owner_id: string
  owner_type: string
  created_at: string
  updated_at: string
}

export default function MetricApprovalWorkflow() {
  const [selectedApproval, setSelectedApproval] = useState<ApprovalItem | null>(null)
  const [comments, setComments] = useState('')
  const [showOwnership, setShowOwnership] = useState(false)
  const queryClient = useQueryClient()

  // Fetch approval queue
  const { data: approvalQueue, isLoading } = useQuery({
    queryKey: ['metric-approval-queue'],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/metrics/approvals/queue')
      return response.data as ApprovalItem[]
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  // Fetch metric owners
  const { data: metricOwners } = useQuery({
    queryKey: ['metric-owners', selectedApproval?.metric_id],
    queryFn: async () => {
      if (!selectedApproval) return []
      const response = await apiClient.get(`/api/v1/metrics/${selectedApproval.metric_id}/owners`)
      return response.data as MetricOwner[]
    },
    enabled: !!selectedApproval && showOwnership,
  })

  // Approve mutation
  const approveMutation = useMutation({
    mutationFn: async ({ approvalId, approverId }: { approvalId: string; approverId: string }) => {
      return apiClient.post(`/api/v1/metrics/approvals/${approvalId}/approve`, {
        approver_id: approverId,
        comments: comments || undefined,
      })
    },
    onSuccess: () => {
      showToast('Metric approved successfully', 'success')
      queryClient.invalidateQueries({ queryKey: ['metric-approval-queue'] })
      setSelectedApproval(null)
      setComments('')
    },
    onError: (error: any) => {
      showToast(error.response?.data?.message || 'Failed to approve metric', 'error')
    },
  })

  // Reject mutation
  const rejectMutation = useMutation({
    mutationFn: async ({ approvalId, approverId }: { approvalId: string; approverId: string }) => {
      return apiClient.post(`/api/v1/metrics/approvals/${approvalId}/reject`, {
        approver_id: approverId,
        comments: comments || 'No comments provided',
      })
    },
    onSuccess: () => {
      showToast('Metric rejected', 'success')
      queryClient.invalidateQueries({ queryKey: ['metric-approval-queue'] })
      setSelectedApproval(null)
      setComments('')
    },
    onError: (error: any) => {
      showToast(error.response?.data?.message || 'Failed to reject metric', 'error')
    },
  })

  // Request changes mutation
  const requestChangesMutation = useMutation({
    mutationFn: async ({ approvalId, approverId }: { approvalId: string; approverId: string }) => {
      return apiClient.post(`/api/v1/metrics/approvals/${approvalId}/request-changes`, {
        approver_id: approverId,
        comments: comments || 'Changes requested',
      })
    },
    onSuccess: () => {
      showToast('Changes requested', 'success')
      queryClient.invalidateQueries({ queryKey: ['metric-approval-queue'] })
      setSelectedApproval(null)
      setComments('')
    },
    onError: (error: any) => {
      showToast(error.response?.data?.message || 'Failed to request changes', 'error')
    },
  })

  const handleApprove = () => {
    if (!selectedApproval) return
    approveMutation.mutate({
      approvalId: selectedApproval.approval_id,
      approverId: selectedApproval.approver_id,
    })
  }

  const handleReject = () => {
    if (!selectedApproval) return
    if (!comments.trim()) {
      showToast('Please provide a reason for rejection', 'error')
      return
    }
    rejectMutation.mutate({
      approvalId: selectedApproval.approval_id,
      approverId: selectedApproval.approver_id,
    })
  }

  const handleRequestChanges = () => {
    if (!selectedApproval) return
    if (!comments.trim()) {
      showToast('Please describe the changes needed', 'error')
      return
    }
    requestChangesMutation.mutate({
      approvalId: selectedApproval.approval_id,
      approverId: selectedApproval.approver_id,
    })
  }

  const getStatusBadge = (status: string) => {
    const variants: Record<string, 'default' | 'primary' | 'secondary' | 'outline' | 'error' | 'success' | 'info' | 'warning'> = {
      pending: 'default',
      approved: 'success',
      rejected: 'error',
      changes_requested: 'warning',
    }
    return (
      <Badge variant={variants[status] || 'default'} className="capitalize">
        {status.replace('_', ' ')}
      </Badge>
    )
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Approval Queue */}
      <Card>
        <CardHeader>
          <CardTitle>Approval Queue</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading...</div>
          ) : !approvalQueue || approvalQueue.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No pending approvals
            </div>
          ) : (
            <div className="space-y-3">
              {approvalQueue.map((item) => (
                <div
                  key={item.approval_id}
                  className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                    selectedApproval?.approval_id === item.approval_id
                      ? 'border-primary bg-primary/5'
                      : 'hover:bg-muted/50'
                  }`}
                  onClick={() => setSelectedApproval(item)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <h4 className="font-semibold">{item.metric_display_name}</h4>
                      <p className="text-sm text-muted-foreground">{item.metric_name}</p>
                      <div className="flex items-center gap-2 mt-2">
                        {getStatusBadge(item.approval_status)}
                        <span className="text-xs text-muted-foreground">
                          {formatDistanceToNow(new Date(item.requested_at), { addSuffix: true })}
                        </span>
                      </div>
                    </div>
                  </div>
                  {item.primary_owner_id && (
                    <p className="text-xs text-muted-foreground mt-2">
                      Owner: {item.primary_owner_id}
                    </p>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Approval Details */}
      <Card>
        <CardHeader>
          <CardTitle>Approval Details</CardTitle>
        </CardHeader>
        <CardContent>
          {!selectedApproval ? (
            <div className="text-center py-8 text-muted-foreground">
              Select an approval from the queue
            </div>
          ) : (
            <div className="space-y-4">
              <div>
                <h3 className="font-semibold text-lg">{selectedApproval.metric_display_name}</h3>
                <p className="text-sm text-muted-foreground">{selectedApproval.metric_name}</p>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Status:</span>
                  {getStatusBadge(selectedApproval.approval_status)}
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Requested:</span>
                  <span className="text-sm text-muted-foreground">
                    {formatDistanceToNow(new Date(selectedApproval.requested_at), { addSuffix: true })}
                  </span>
                </div>
                {selectedApproval.primary_owner_id && (
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Owner:</span>
                    <span className="text-sm text-muted-foreground">
                      {selectedApproval.primary_owner_id}
                    </span>
                  </div>
                )}
              </div>

              {selectedApproval.comments && (
                <div>
                  <span className="text-sm font-medium">Previous Comments:</span>
                  <p className="text-sm text-muted-foreground mt-1 p-2 bg-muted rounded">
                    {selectedApproval.comments}
                  </p>
                </div>
              )}

              <div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowOwnership(!showOwnership)}
                >
                  {showOwnership ? 'Hide' : 'Show'} Ownership
                </Button>
                {showOwnership && metricOwners && (
                  <div className="mt-2 space-y-1">
                    {metricOwners.map((owner, idx) => (
                      <div key={idx} className="text-sm">
                        <span className="font-medium">{owner.owner_id}</span>
                        <Badge variant="outline" className="ml-2 text-xs">
                          {owner.owner_type}
                        </Badge>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <div>
                <label className="text-sm font-medium">Comments</label>
                <Textarea
                  value={comments}
                  onChange={(e) => setComments(e.target.value)}
                  placeholder="Add your comments or feedback..."
                  className="mt-1"
                  rows={4}
                />
              </div>

              <div className="flex gap-2 pt-4">
                <Button
                  onClick={handleApprove}
                  disabled={approveMutation.isPending}
                  className="flex-1"
                >
                  {approveMutation.isPending ? 'Approving...' : 'Approve'}
                </Button>
                <Button
                  onClick={handleRequestChanges}
                  variant="outline"
                  disabled={requestChangesMutation.isPending}
                  className="flex-1"
                >
                  {requestChangesMutation.isPending ? 'Requesting...' : 'Request Changes'}
                </Button>
                <Button
                  onClick={handleReject}
                  variant="destructive"
                  disabled={rejectMutation.isPending}
                  className="flex-1"
                >
                  {rejectMutation.isPending ? 'Rejecting...' : 'Reject'}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

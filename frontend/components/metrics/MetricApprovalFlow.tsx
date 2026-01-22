'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { showToast } from '@/components/ui/Toast'
import { Badge } from '@/components/ui/Badge'
import { Textarea } from '@/components/ui/Textarea'

interface MetricApproval {
  id: string
  metric_id: string
  approver_id: string
  status: 'pending' | 'approved' | 'rejected' | 'changes_requested'
  comments?: string
  approved_at?: string
  created_at: string
}

interface MetricApprovalFlowProps {
  metricId: string
  currentUserId: string
  onApprovalChange?: () => void
}

export default function MetricApprovalFlow({
  metricId,
  currentUserId,
  onApprovalChange,
}: MetricApprovalFlowProps) {
  const [approvals, setApprovals] = useState<MetricApproval[]>([])
  const [loading, setLoading] = useState(false)
  const [comments, setComments] = useState('')
  const [action, setAction] = useState<'approve' | 'reject' | 'request_changes' | null>(null)

  useEffect(() => {
    fetchApprovals()
  }, [metricId])

  const fetchApprovals = async () => {
    try {
      const response = await fetch(`/api/v1/metrics/${metricId}/approvals`)
      if (!response.ok) throw new Error('Failed to fetch approvals')
      const data = await response.json()
      setApprovals(data)
    } catch (error) {
      console.error('Error fetching approvals:', error)
    }
  }

  const handleAction = async (approvalId: string, actionType: 'approve' | 'reject' | 'request_changes') => {
    setLoading(true)
    try {
      const endpoint = actionType === 'approve' 
        ? `/api/v1/metrics/approvals/${approvalId}/approve`
        : actionType === 'reject'
        ? `/api/v1/metrics/approvals/${approvalId}/reject`
        : `/api/v1/metrics/approvals/${approvalId}/request-changes`

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ comments }),
      })

      if (!response.ok) throw new Error('Failed to update approval')
      
      showToast(`Metric ${actionType === 'approve' ? 'approved' : actionType === 'reject' ? 'rejected' : 'changes requested'}`, 'success')
      setComments('')
      setAction(null)
      fetchApprovals()
      onApprovalChange?.()
    } catch (error: any) {
      showToast(error.message || 'Failed to update approval', 'error')
    } finally {
      setLoading(false)
    }
  }

  const requestApproval = async (approverId: string) => {
    setLoading(true)
    try {
      const response = await fetch(`/api/v1/metrics/${metricId}/approvals`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ approver_id: approverId, comments }),
      })

      if (!response.ok) throw new Error('Failed to request approval')
      
      showToast('Approval requested', 'success')
      setComments('')
      fetchApprovals()
      onApprovalChange?.()
    } catch (error: any) {
      showToast(error.message || 'Failed to request approval', 'error')
    } finally {
      setLoading(false)
    }
  }

  const getStatusBadge = (status: string) => {
    const variants = {
      pending: 'warning',
      approved: 'success',
      rejected: 'error',
      changes_requested: 'info',
    } as const
    return <Badge variant={variants[status as keyof typeof variants] || 'default'}>{status}</Badge>
  }

  const pendingApproval = approvals.find(a => a.status === 'pending')

  return (
    <Card>
      <CardHeader>
        <CardTitle>Approval Workflow</CardTitle>
        <CardDescription>Track and manage metric approval status</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {pendingApproval && (
          <div className="border rounded-lg p-4 bg-yellow-50 dark:bg-yellow-900/20">
            <div className="flex items-center justify-between mb-2">
              <div>
                <h4 className="font-medium">Pending Approval</h4>
                <p className="text-sm text-muted-foreground">
                  Requested by {pendingApproval.approver_id}
                </p>
              </div>
              {getStatusBadge(pendingApproval.status)}
            </div>
            
            {pendingApproval.approver_id === currentUserId && (
              <div className="mt-4 space-y-2">
                <Textarea
                  placeholder="Add comments..."
                  value={comments}
                  onChange={(e) => setComments(e.target.value)}
                  rows={3}
                />
                <div className="flex gap-2">
                  <Button
                    onClick={() => handleAction(pendingApproval.id, 'approve')}
                    disabled={loading}
                    variant="primary"
                  >
                    Approve
                  </Button>
                  <Button
                    onClick={() => handleAction(pendingApproval.id, 'request_changes')}
                    disabled={loading}
                    variant="secondary"
                  >
                    Request Changes
                  </Button>
                  <Button
                    onClick={() => handleAction(pendingApproval.id, 'reject')}
                    disabled={loading}
                    variant="destructive"
                  >
                    Reject
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}

        <div className="space-y-2">
          <h4 className="font-medium">Approval History</h4>
          {approvals.length === 0 ? (
            <p className="text-sm text-muted-foreground">No approvals yet</p>
          ) : (
            <div className="space-y-2">
              {approvals.map((approval) => (
                <motion.div
                  key={approval.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="border rounded-lg p-3"
                >
                  <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">{approval.approver_id}</span>
                      {getStatusBadge(approval.status)}
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {new Date(approval.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  {approval.comments && (
                    <p className="text-sm text-muted-foreground mt-1">{approval.comments}</p>
                  )}
                </motion.div>
              ))}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

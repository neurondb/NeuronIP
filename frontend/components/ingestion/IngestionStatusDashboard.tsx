'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { showToast } from '@/components/ui/Toast'
import { formatDistanceToNow } from 'date-fns'

interface IngestionStatus {
  data_source_id: string
  total_jobs: number
  running_jobs: number
  completed_jobs: number
  failed_jobs: number
  pending_jobs: number
  total_rows_processed: number
  last_sync_at?: string
  last_sync_status: string
  cdc_enabled: boolean
  cdc_lag?: number
  recent_jobs: IngestionJob[]
}

interface IngestionJob {
  id: string
  data_source_id: string
  job_type: string
  status: string
  error_message?: string
  rows_processed: number
  started_at?: string
  completed_at?: string
  created_at: string
}

interface IngestionFailure {
  job_id: string
  data_source_id: string
  job_type: string
  error_message: string
  failed_at: string
  retry_count: number
}

interface IngestionStatusDashboardProps {
  dataSourceId: string
}

export default function IngestionStatusDashboard({ dataSourceId }: IngestionStatusDashboardProps) {
  const [activeTab, setActiveTab] = useState<'status' | 'failures'>('status')
  const queryClient = useQueryClient()

  // Fetch ingestion status
  const { data: status, isLoading: statusLoading } = useQuery({
    queryKey: ['ingestion-status', dataSourceId],
    queryFn: async () => {
      const response = await apiClient.get(`/api/v1/ingestion/data-sources/${dataSourceId}/status`)
      return response.data as IngestionStatus
    },
    refetchInterval: 10000, // Refresh every 10 seconds
  })

  // Fetch failures
  const { data: failures, isLoading: failuresLoading } = useQuery({
    queryKey: ['ingestion-failures', dataSourceId],
    queryFn: async () => {
      const response = await apiClient.get(`/api/v1/ingestion/failures?data_source_id=${dataSourceId}`)
      return response.data as IngestionFailure[]
    },
    enabled: activeTab === 'failures',
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  // Retry mutation
  const retryMutation = useMutation({
    mutationFn: async (jobId: string) => {
      return apiClient.post(`/api/v1/ingestion/jobs/${jobId}/retry`)
    },
    onSuccess: () => {
      showToast('Job retry initiated', 'success')
      queryClient.invalidateQueries({ queryKey: ['ingestion-status', dataSourceId] })
      queryClient.invalidateQueries({ queryKey: ['ingestion-failures', dataSourceId] })
    },
    onError: (error: any) => {
      showToast(error.response?.data?.message || 'Failed to retry job', 'error')
    },
  })

  const getStatusBadge = (status: string) => {
    const variants: Record<string, 'default' | 'success' | 'error' | 'warning'> = {
      pending: 'default',
      running: 'warning',
      completed: 'success',
      failed: 'error',
    }
    return (
      <Badge variant={variants[status] || 'default'} className="capitalize">
        {status}
      </Badge>
    )
  }

  const formatCDCLag = (lagMs?: number) => {
    if (!lagMs) return 'N/A'
    const seconds = Math.floor(lagMs / 1000)
    if (seconds < 60) return `${seconds}s`
    const minutes = Math.floor(seconds / 60)
    if (minutes < 60) return `${minutes}m`
    const hours = Math.floor(minutes / 60)
    return `${hours}h`
  }

  if (statusLoading) {
    return <div className="text-center py-8 text-muted-foreground">Loading...</div>
  }

  if (!status) {
    return <div className="text-center py-8 text-muted-foreground">No status available</div>
  }

  return (
    <div className="space-y-6">
      {/* Status Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Jobs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{status.total_jobs}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Running</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">{status.running_jobs}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Completed</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{status.completed_jobs}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Failed</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{status.failed_jobs}</div>
          </CardContent>
        </Card>
      </div>

      {/* Additional Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Rows Processed</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {status.total_rows_processed.toLocaleString()}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Last Sync</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm">
              {status.last_sync_at ? (
                <>
                  <div>{formatDistanceToNow(new Date(status.last_sync_at), { addSuffix: true })}</div>
                  {getStatusBadge(status.last_sync_status)}
                </>
              ) : (
                <div className="text-muted-foreground">Never</div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">CDC Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-1">
              <Badge variant={status.cdc_enabled ? 'success' : 'default'}>
                {status.cdc_enabled ? 'Enabled' : 'Disabled'}
              </Badge>
              {status.cdc_enabled && status.cdc_lag !== undefined && (
                <div className="text-xs text-muted-foreground mt-1">
                  Lag: {formatCDCLag(status.cdc_lag)}
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <div className="border-b">
        <div className="flex space-x-4">
          <button
            onClick={() => setActiveTab('status')}
            className={`pb-2 px-1 border-b-2 ${
              activeTab === 'status'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground'
            }`}
          >
            Recent Jobs
          </button>
          <button
            onClick={() => setActiveTab('failures')}
            className={`pb-2 px-1 border-b-2 ${
              activeTab === 'failures'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground'
            }`}
          >
            Failure Queue ({status.failed_jobs})
          </button>
        </div>
      </div>

      {/* Content */}
      {activeTab === 'status' && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Jobs</CardTitle>
          </CardHeader>
          <CardContent>
            {!status.recent_jobs || status.recent_jobs.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">No recent jobs</div>
            ) : (
              <div className="space-y-3">
                {status.recent_jobs.map((job) => (
                  <div key={job.id} className="p-4 border rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <h4 className="font-semibold">{job.job_type}</h4>
                          {getStatusBadge(job.status)}
                        </div>
                        <div className="mt-2 space-y-1 text-sm text-muted-foreground">
                          <div>Rows: {job.rows_processed.toLocaleString()}</div>
                          {job.started_at && (
                            <div>
                              Started: {formatDistanceToNow(new Date(job.started_at), { addSuffix: true })}
                            </div>
                          )}
                          {job.completed_at && (
                            <div>
                              Completed: {formatDistanceToNow(new Date(job.completed_at), { addSuffix: true })}
                            </div>
                          )}
                        </div>
                        {job.error_message && (
                          <div className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 rounded text-sm text-red-600 dark:text-red-400">
                            {job.error_message}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {activeTab === 'failures' && (
        <Card>
          <CardHeader>
            <CardTitle>Failure Queue</CardTitle>
          </CardHeader>
          <CardContent>
            {failuresLoading ? (
              <div className="text-center py-8 text-muted-foreground">Loading...</div>
            ) : !failures || failures.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">No failures</div>
            ) : (
              <div className="space-y-3">
                {failures.map((failure) => (
                  <div key={failure.job_id} className="p-4 border border-red-200 dark:border-red-800 rounded-lg">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <h4 className="font-semibold">{failure.job_type}</h4>
                          <Badge variant="error">Failed</Badge>
                          {failure.retry_count > 0 && (
                            <Badge variant="outline">Retries: {failure.retry_count}</Badge>
                          )}
                        </div>
                        <div className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 rounded text-sm text-red-600 dark:text-red-400">
                          {failure.error_message}
                        </div>
                        <div className="mt-2 text-xs text-muted-foreground">
                          Failed: {formatDistanceToNow(new Date(failure.failed_at), { addSuffix: true })}
                        </div>
                      </div>
                      <Button
                        onClick={() => retryMutation.mutate(failure.job_id)}
                        disabled={retryMutation.isPending}
                        size="sm"
                        variant="outline"
                      >
                        {retryMutation.isPending ? 'Retrying...' : 'Retry'}
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}

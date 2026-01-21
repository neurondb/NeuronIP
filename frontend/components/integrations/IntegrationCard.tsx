'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useDeleteIntegration, useTestIntegration } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { TrashIcon, CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline'

interface IntegrationCardProps {
  integration: {
    id: string
    name: string
    integration_type: string
    enabled: boolean
    status: string
    status_message?: string
    last_sync_at?: string
  }
  onEdit?: (id: string) => void
  onTest?: (id: string) => void
  onRefresh?: () => void
}

export default function IntegrationCard({ integration, onEdit, onTest, onRefresh }: IntegrationCardProps) {
  const deleteIntegration = useDeleteIntegration()
  const testIntegration = useTestIntegration()

  const handleDelete = async () => {
    if (!confirm(`Are you sure you want to delete ${integration.name}?`)) {
      return
    }

    try {
      await deleteIntegration.mutateAsync(integration.id)
      showToast('Integration deleted successfully', 'success')
      onRefresh?.()
    } catch (error: any) {
      showToast('Failed to delete integration', 'error')
    }
  }

  const handleTest = async () => {
    try {
      await testIntegration.mutateAsync(integration.id)
      showToast('Integration test passed', 'success')
      onRefresh?.()
    } catch (error: any) {
      showToast('Integration test failed', 'error')
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-green-500'
      case 'error':
        return 'bg-red-500'
      case 'warning':
        return 'bg-yellow-500'
      case 'disabled':
        return 'bg-gray-400'
      default:
        return 'bg-gray-400'
    }
  }

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'slack':
        return 'Slack'
      case 'teams':
        return 'Microsoft Teams'
      case 'email':
        return 'Email'
      case 'webhook':
        return 'Webhook'
      case 'helpdesk':
        return 'Helpdesk'
      default:
        return type
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={`h-3 w-3 rounded-full ${getStatusColor(integration.status)}`} />
            <div>
              <CardTitle className="text-lg">{integration.name}</CardTitle>
              <CardDescription>{getTypeLabel(integration.integration_type)}</CardDescription>
            </div>
          </div>
          <div className="flex items-center gap-2">
            {integration.enabled ? (
              <CheckCircleIcon className="h-5 w-5 text-green-500" />
            ) : (
              <XCircleIcon className="h-5 w-5 text-gray-400" />
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {integration.status_message && (
          <div className="text-sm text-muted-foreground">{integration.status_message}</div>
        )}
        {integration.last_sync_at && (
          <div className="text-xs text-muted-foreground">
            Last sync: {new Date(integration.last_sync_at).toLocaleString()}
          </div>
        )}
        <div className="flex gap-2">
          {onTest && (
            <Button
              onClick={handleTest}
              variant="outline"
              size="sm"
              disabled={testIntegration.isPending}
            >
              Test
            </Button>
          )}
          {onEdit && (
            <Button onClick={() => onEdit(integration.id)} variant="outline" size="sm">
              Configure
            </Button>
          )}
          <Button
            onClick={handleDelete}
            variant="ghost"
            size="sm"
            disabled={deleteIntegration.isPending}
          >
            <TrashIcon className="h-4 w-4 text-destructive" />
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

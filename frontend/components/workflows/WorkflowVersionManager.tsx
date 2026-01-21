'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import {
  useWorkflowVersions,
  useCreateWorkflowVersion,
  useWorkflow,
} from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface WorkflowVersionManagerProps {
  workflowId: string
}

export default function WorkflowVersionManager({ workflowId }: WorkflowVersionManagerProps) {
  const [newVersion, setNewVersion] = useState('')
  const [changes, setChanges] = useState('')
  const { data: versions, isLoading } = useWorkflowVersions(workflowId)
  const { data: workflow } = useWorkflow(workflowId)
  const { mutate: createVersion, isPending } = useCreateWorkflowVersion(workflowId)

  const handleCreateVersion = () => {
    if (!newVersion.trim()) {
      showToast('Please enter a version number', 'warning')
      return
    }

    createVersion(
      {
        version: newVersion,
        changes: changes ? JSON.parse(changes) : {},
      },
      {
        onSuccess: () => {
          showToast('Version created successfully', 'success')
          setNewVersion('')
          setChanges('')
        },
        onError: () => {
          showToast('Failed to create version', 'error')
        },
      }
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Version Management</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Current Version */}
        {workflow && (
          <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded">
            <div className="text-sm font-medium text-blue-900">Current Version</div>
            <div className="text-lg font-bold text-blue-700">{workflow.version || '1.0.0'}</div>
            <div className="text-xs text-blue-600 mt-1">
              Last updated: {workflow.updated_at ? new Date(workflow.updated_at).toLocaleString() : 'N/A'}
            </div>
          </div>
        )}

        {/* Create New Version */}
        <div className="mb-6 p-4 border border-border rounded-lg">
          <h3 className="font-semibold mb-3">Create New Version</h3>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium mb-1 block">Version Number</label>
              <input
                type="text"
                value={newVersion}
                onChange={(e) => setNewVersion(e.target.value)}
                placeholder="e.g., 1.1.0"
                className="w-full rounded border border-border px-3 py-2"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-1 block">Changes (JSON)</label>
              <textarea
                value={changes}
                onChange={(e) => setChanges(e.target.value)}
                placeholder='{"description": "Added new step", "steps": [...]}'
                className="w-full rounded border border-border px-3 py-2 font-mono text-sm"
                rows={4}
              />
            </div>
            <Button onClick={handleCreateVersion} disabled={isPending || !newVersion.trim()}>
              {isPending ? 'Creating...' : 'Create Version'}
            </Button>
          </div>
        </div>

        {/* Version History */}
        <div>
          <h3 className="font-semibold mb-3">Version History</h3>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading versions...</div>
          ) : versions?.versions?.length > 0 ? (
            <div className="space-y-2">
              {versions.versions.map((version: any) => (
                <div
                  key={version.id}
                  className="p-3 border border-border rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-semibold">v{version.version}</span>
                    <span className="text-xs text-muted-foreground">
                      {new Date(version.created_at).toLocaleString()}
                    </span>
                  </div>
                  {version.changes && (
                    <div className="text-sm text-muted-foreground">
                      {typeof version.changes === 'string'
                        ? version.changes
                        : JSON.stringify(version.changes, null, 2)}
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className="text-sm text-muted-foreground">No versions available</div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

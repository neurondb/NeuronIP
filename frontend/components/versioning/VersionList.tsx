'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useVersions, useRollbackVersion } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { ClockIcon, ArrowPathIcon } from '@heroicons/react/24/outline'
import { format } from 'date-fns'

interface VersionListProps {
  resourceType: string
  resourceId: string
}

export default function VersionList({ resourceType, resourceId }: VersionListProps) {
  const { data: versionsData, isLoading } = useVersions(resourceType, resourceId)
  const rollbackMutation = useRollbackVersion()

  const versions = versionsData?.versions || []

  const handleRollback = async (versionId: string) => {
    if (!confirm('Are you sure you want to rollback to this version?')) return

    try {
      await rollbackMutation.mutateAsync(versionId)
      showToast('Version rolled back successfully', 'success')
    } catch (error: any) {
      showToast('Failed to rollback version', 'error')
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Loading versions...</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Versions for {resourceType}/{resourceId.slice(0, 8)}...</CardTitle>
      </CardHeader>
      <CardContent>
        {versions.length === 0 ? (
          <p className="text-center text-muted-foreground py-8">No versions found</p>
        ) : (
          <div className="space-y-3">
            {versions.map((version: any) => (
              <div
                key={version.id}
                className={`p-4 border rounded-lg ${
                  version.is_current ? 'border-primary bg-primary/5' : 'border-border'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <ClockIcon className="h-5 w-5 text-muted-foreground" />
                      <span className="font-semibold">v{version.version_number}</span>
                      {version.is_current && (
                        <span className="text-xs px-2 py-0.5 bg-primary/10 text-primary rounded">Current</span>
                      )}
                    </div>
                    <p className="text-sm text-muted-foreground mt-1">
                      Created {format(new Date(version.created_at), 'PPpp')}
                    </p>
                    {version.created_by && (
                      <p className="text-xs text-muted-foreground mt-1">By {version.created_by}</p>
                    )}
                    {version.version_data && Object.keys(version.version_data).length > 0 && (
                      <div className="mt-2 text-xs text-muted-foreground">
                        {Object.keys(version.version_data).length} data fields
                      </div>
                    )}
                  </div>
                  {!version.is_current && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleRollback(version.id)}
                      disabled={rollbackMutation.isPending}
                    >
                      <ArrowPathIcon className="h-4 w-4 mr-1" />
                      Rollback
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
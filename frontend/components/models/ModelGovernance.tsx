'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Textarea } from '@/components/ui/Textarea'
import { showToast } from '@/components/ui/Toast'
import { formatDistanceToNow } from 'date-fns'

interface Model {
  id: string
  model_name: string
  version: string
  provider: string
  model_id: string
  status: string
  approved_by?: string
  approved_at?: string
  created_by: string
  created_at: string
  updated_at: string
}

interface PromptTemplate {
  id: string
  name: string
  version: string
  template_text: string
  variables: string[]
  status: string
  approved_by?: string
  approved_at?: string
  created_by: string
  created_at: string
}

interface ModelGovernanceProps {
  workspaceId?: string
}

export default function ModelGovernance({ workspaceId }: ModelGovernanceProps) {
  const [activeTab, setActiveTab] = useState<'models' | 'prompts'>('models')
  const [selectedModel, setSelectedModel] = useState<Model | null>(null)
  const [selectedPrompt, setSelectedPrompt] = useState<PromptTemplate | null>(null)
  const [showRollback, setShowRollback] = useState(false)
  const [targetVersion, setTargetVersion] = useState('')
  const queryClient = useQueryClient()

  // Fetch models
  const { data: models, isLoading: modelsLoading } = useQuery({
    queryKey: ['models', workspaceId],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/models', {
        params: workspaceId ? { workspace_id: workspaceId } : {},
      })
      return response.data as Model[]
    },
  })

  // Fetch prompts
  const { data: prompts, isLoading: promptsLoading } = useQuery({
    queryKey: ['prompt-templates'],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/prompts')
      return response.data as PromptTemplate[]
    },
    enabled: activeTab === 'prompts',
  })

  // Fetch model versions
  const { data: modelVersions } = useQuery({
    queryKey: ['model-versions', selectedModel?.model_name],
    queryFn: async () => {
      if (!selectedModel) return []
      const response = await apiClient.get(`/api/v1/models/${selectedModel.id}/versions`)
      return response.data as Model[]
    },
    enabled: !!selectedModel && showRollback,
  })

  // Fetch prompt versions
  const { data: promptVersions } = useQuery({
    queryKey: ['prompt-versions', selectedPrompt?.name],
    queryFn: async () => {
      if (!selectedPrompt) return []
      const response = await apiClient.get(`/api/v1/prompts/${selectedPrompt.name}/versions`)
      return response.data as PromptTemplate[]
    },
    enabled: !!selectedPrompt && showRollback,
  })

  // Approve model mutation
  const approveModelMutation = useMutation({
    mutationFn: async ({ modelId, approverId }: { modelId: string; approverId: string }) => {
      return apiClient.post(`/api/v1/models/${modelId}/approve`, { approver_id: approverId })
    },
    onSuccess: () => {
      showToast('Model approved', 'success')
      queryClient.invalidateQueries({ queryKey: ['models'] })
    },
  })

  // Rollback model mutation
  const rollbackModelMutation = useMutation({
    mutationFn: async ({ modelName, targetVersion }: { modelName: string; targetVersion: string }) => {
      return apiClient.post(`/api/v1/models/${modelName}/rollback`, { target_version: targetVersion })
    },
    onSuccess: () => {
      showToast('Model rolled back', 'success')
      queryClient.invalidateQueries({ queryKey: ['models'] })
      setShowRollback(false)
      setTargetVersion('')
    },
  })

  // Approve prompt mutation
  const approvePromptMutation = useMutation({
    mutationFn: async ({ promptId, approverId }: { promptId: string; approverId: string }) => {
      return apiClient.post(`/api/v1/prompts/${promptId}/approve`, { approver_id: approverId })
    },
    onSuccess: () => {
      showToast('Prompt approved', 'success')
      queryClient.invalidateQueries({ queryKey: ['prompt-templates'] })
    },
  })

  // Rollback prompt mutation
  const rollbackPromptMutation = useMutation({
    mutationFn: async ({ promptName, targetVersion }: { promptName: string; targetVersion: string }) => {
      return apiClient.post(`/api/v1/prompts/${promptName}/rollback`, { target_version: targetVersion })
    },
    onSuccess: () => {
      showToast('Prompt rolled back', 'success')
      queryClient.invalidateQueries({ queryKey: ['prompt-templates'] })
      setShowRollback(false)
      setTargetVersion('')
    },
  })

  const getStatusBadge = (status: string) => {
    const variants: Record<string, 'default' | 'success' | 'error' | 'warning'> = {
      draft: 'default',
      pending_approval: 'warning',
      approved: 'success',
      deprecated: 'error',
    }
    return (
      <Badge variant={variants[status] || 'default'} className="capitalize">
        {status.replace('_', ' ')}
      </Badge>
    )
  }

  const handleRollback = () => {
    if (activeTab === 'models' && selectedModel) {
      if (!targetVersion) {
        showToast('Please select a target version', 'error')
        return
      }
      rollbackModelMutation.mutate({
        modelName: selectedModel.model_name,
        targetVersion,
      })
    } else if (activeTab === 'prompts' && selectedPrompt) {
      if (!targetVersion) {
        showToast('Please select a target version', 'error')
        return
      }
      rollbackPromptMutation.mutate({
        promptName: selectedPrompt.name,
        targetVersion,
      })
    }
  }

  return (
    <div className="space-y-6">
      {/* Tabs */}
      <div className="border-b">
        <div className="flex space-x-4">
          <button
            onClick={() => {
              setActiveTab('models')
              setSelectedModel(null)
              setShowRollback(false)
            }}
            className={`pb-2 px-1 border-b-2 ${
              activeTab === 'models'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground'
            }`}
          >
            Models
          </button>
          <button
            onClick={() => {
              setActiveTab('prompts')
              setSelectedPrompt(null)
              setShowRollback(false)
            }}
            className={`pb-2 px-1 border-b-2 ${
              activeTab === 'prompts'
                ? 'border-primary text-primary'
                : 'border-transparent text-muted-foreground'
            }`}
          >
            Prompt Templates
          </button>
        </div>
      </div>

      {/* Models Tab */}
      {activeTab === 'models' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Model Registry</CardTitle>
            </CardHeader>
            <CardContent>
              {modelsLoading ? (
                <div className="text-center py-8 text-muted-foreground">Loading...</div>
              ) : !models || models.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">No models registered</div>
              ) : (
                <div className="space-y-3">
                  {models.map((model) => (
                    <div
                      key={model.id}
                      className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                        selectedModel?.id === model.id
                          ? 'border-primary bg-primary/5'
                          : 'hover:bg-muted/50'
                      }`}
                      onClick={() => {
                        setSelectedModel(model)
                        setShowRollback(false)
                      }}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h4 className="font-semibold">{model.model_name}</h4>
                          <p className="text-sm text-muted-foreground">
                            {model.provider} / {model.model_id}
                          </p>
                          <div className="flex items-center gap-2 mt-2">
                            {getStatusBadge(model.status)}
                            <span className="text-xs text-muted-foreground">v{model.version}</span>
                          </div>
                        </div>
                      </div>
                      {model.approved_by && (
                        <p className="text-xs text-muted-foreground mt-2">
                          Approved by {model.approved_by}{' '}
                          {model.approved_at &&
                            formatDistanceToNow(new Date(model.approved_at), { addSuffix: true })}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Model Details</CardTitle>
            </CardHeader>
            <CardContent>
              {!selectedModel ? (
                <div className="text-center py-8 text-muted-foreground">
                  Select a model to view details
                </div>
              ) : (
                <div className="space-y-4">
                  <div>
                    <h3 className="font-semibold text-lg">{selectedModel.model_name}</h3>
                    <p className="text-sm text-muted-foreground">
                      {selectedModel.provider} / {selectedModel.model_id}
                    </p>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Version:</span>
                      <span className="text-sm">{selectedModel.version}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Status:</span>
                      {getStatusBadge(selectedModel.status)}
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Created:</span>
                      <span className="text-sm text-muted-foreground">
                        {formatDistanceToNow(new Date(selectedModel.created_at), { addSuffix: true })}
                      </span>
                    </div>
                    {selectedModel.approved_by && (
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium">Approved by:</span>
                        <span className="text-sm text-muted-foreground">
                          {selectedModel.approved_by}
                        </span>
                      </div>
                    )}
                  </div>

                  <div className="flex gap-2 pt-4">
                    {selectedModel.status === 'pending_approval' && (
                      <Button
                        onClick={() =>
                          approveModelMutation.mutate({
                            modelId: selectedModel.id,
                            approverId: 'current-user', // In production, get from auth
                          })
                        }
                        disabled={approveModelMutation.isPending}
                        className="flex-1"
                      >
                        {approveModelMutation.isPending ? 'Approving...' : 'Approve'}
                      </Button>
                    )}
                    <Button
                      onClick={() => setShowRollback(!showRollback)}
                      variant="outline"
                      className="flex-1"
                    >
                      {showRollback ? 'Cancel' : 'Rollback'}
                    </Button>
                  </div>

                  {showRollback && (
                    <div className="space-y-2 pt-4 border-t">
                      <label className="text-sm font-medium">Target Version</label>
                      <select
                        value={targetVersion}
                        onChange={(e) => setTargetVersion(e.target.value)}
                        className="w-full p-2 border rounded"
                      >
                        <option value="">Select version...</option>
                        {modelVersions
                          ?.filter((v) => v.version !== selectedModel.version)
                          .map((v) => (
                            <option key={v.id} value={v.version}>
                              {v.version} ({v.status})
                            </option>
                          ))}
                      </select>
                      <Button
                        onClick={handleRollback}
                        disabled={rollbackModelMutation.isPending || !targetVersion}
                        variant="destructive"
                        className="w-full"
                      >
                        {rollbackModelMutation.isPending ? 'Rolling back...' : 'Rollback to Version'}
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Prompts Tab */}
      {activeTab === 'prompts' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Prompt Templates</CardTitle>
            </CardHeader>
            <CardContent>
              {promptsLoading ? (
                <div className="text-center py-8 text-muted-foreground">Loading...</div>
              ) : !prompts || prompts.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">No prompt templates</div>
              ) : (
                <div className="space-y-3">
                  {prompts.map((prompt) => (
                    <div
                      key={prompt.id}
                      className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                        selectedPrompt?.id === prompt.id
                          ? 'border-primary bg-primary/5'
                          : 'hover:bg-muted/50'
                      }`}
                      onClick={() => {
                        setSelectedPrompt(prompt)
                        setShowRollback(false)
                      }}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h4 className="font-semibold">{prompt.name}</h4>
                          <div className="flex items-center gap-2 mt-2">
                            {getStatusBadge(prompt.status)}
                            <span className="text-xs text-muted-foreground">v{prompt.version}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Prompt Details</CardTitle>
            </CardHeader>
            <CardContent>
              {!selectedPrompt ? (
                <div className="text-center py-8 text-muted-foreground">
                  Select a prompt to view details
                </div>
              ) : (
                <div className="space-y-4">
                  <div>
                    <h3 className="font-semibold text-lg">{selectedPrompt.name}</h3>
                    <p className="text-sm text-muted-foreground">Version {selectedPrompt.version}</p>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Status:</span>
                      {getStatusBadge(selectedPrompt.status)}
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Variables:</span>
                      <span className="text-sm text-muted-foreground">
                        {selectedPrompt.variables.join(', ')}
                      </span>
                    </div>
                  </div>

                  <div>
                    <label className="text-sm font-medium">Template:</label>
                    <Textarea
                      value={selectedPrompt.template_text}
                      readOnly
                      className="mt-1 font-mono text-xs"
                      rows={8}
                    />
                  </div>

                  <div className="flex gap-2 pt-4">
                    {selectedPrompt.status === 'pending_approval' && (
                      <Button
                        onClick={() =>
                          approvePromptMutation.mutate({
                            promptId: selectedPrompt.id,
                            approverId: 'current-user',
                          })
                        }
                        disabled={approvePromptMutation.isPending}
                        className="flex-1"
                      >
                        {approvePromptMutation.isPending ? 'Approving...' : 'Approve'}
                      </Button>
                    )}
                    <Button
                      onClick={() => setShowRollback(!showRollback)}
                      variant="outline"
                      className="flex-1"
                    >
                      {showRollback ? 'Cancel' : 'Rollback'}
                    </Button>
                  </div>

                  {showRollback && (
                    <div className="space-y-2 pt-4 border-t">
                      <label className="text-sm font-medium">Target Version</label>
                      <select
                        value={targetVersion}
                        onChange={(e) => setTargetVersion(e.target.value)}
                        className="w-full p-2 border rounded"
                      >
                        <option value="">Select version...</option>
                        {promptVersions
                          ?.filter((v) => v.version !== selectedPrompt.version)
                          .map((v) => (
                            <option key={v.id} value={v.version}>
                              {v.version} ({v.status})
                            </option>
                          ))}
                      </select>
                      <Button
                        onClick={handleRollback}
                        disabled={rollbackPromptMutation.isPending || !targetVersion}
                        variant="destructive"
                        className="w-full"
                      >
                        {rollbackPromptMutation.isPending ? 'Rolling back...' : 'Rollback to Version'}
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useAgents, useDeleteAgent, useDeployAgent } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { CpuChipIcon, PlayIcon, TrashIcon, Cog6ToothIcon } from '@heroicons/react/24/outline'
import { format } from 'date-fns'

interface AgentListProps {
  onSelectAgent?: (agentId: string) => void
  onCreateNew?: () => void
}

export default function AgentList({ onSelectAgent, onCreateNew }: AgentListProps) {
  const { data: agentsData, isLoading } = useAgents()
  const deleteMutation = useDeleteAgent()
  const deployMutation = useDeployAgent()

  const agents = agentsData?.agents || agentsData || []

  const handleDelete = async (agentId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm('Are you sure you want to delete this agent?')) return

    try {
      await deleteMutation.mutateAsync(agentId)
      showToast('Agent deleted successfully', 'success')
    } catch (error: any) {
      showToast('Failed to delete agent', 'error')
    }
  }

  const handleDeploy = async (agentId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    try {
      await deployMutation.mutateAsync(agentId)
      showToast('Agent deployed successfully', 'success')
    } catch (error: any) {
      showToast('Failed to deploy agent', 'error')
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Loading agents...</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Agents ({agents.length})</CardTitle>
            {onCreateNew && (
              <Button onClick={onCreateNew} size="sm">
                New Agent
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {agents.length === 0 ? (
            <div className="text-center py-12">
              <CpuChipIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4 opacity-50" />
              <p className="text-muted-foreground mb-4">No agents found. Create one to get started.</p>
              {onCreateNew && (
                <Button onClick={onCreateNew} size="md">
                  Create Your First Agent
                </Button>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {agents.map((agent: any) => (
                <motion.div
                  key={agent.id}
                  whileHover={{ y: -2 }}
                  className="border border-border rounded-lg p-4 hover:shadow-md transition-shadow cursor-pointer"
                  onClick={() => onSelectAgent?.(agent.id)}
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <CpuChipIcon className="h-5 w-5 text-purple-600" />
                      <h3 className="font-semibold text-sm">{agent.name || agent.id}</h3>
                    </div>
                    <span
                      className={`text-xs px-2 py-1 rounded ${
                        agent.status === 'active' || agent.status === 'deployed'
                          ? 'bg-green-100 text-green-800'
                          : agent.status === 'draft'
                          ? 'bg-gray-100 text-gray-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {agent.status || 'draft'}
                    </span>
                  </div>

                  {agent.description && (
                    <p className="text-xs text-muted-foreground mb-3 line-clamp-2">{agent.description}</p>
                  )}

                  {agent.created_at && (
                    <p className="text-xs text-muted-foreground mb-3">
                      Created {format(new Date(agent.created_at), 'PP')}
                    </p>
                  )}

                  <div className="flex gap-2 mt-4">
                    {(agent.status === 'draft' || !agent.status) && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={(e) => handleDeploy(agent.id, e)}
                        disabled={deployMutation.isPending}
                      >
                        <PlayIcon className="h-4 w-4 mr-1" />
                        Deploy
                      </Button>
                    )}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onSelectAgent?.(agent.id)}
                    >
                      <Cog6ToothIcon className="h-4 w-4 mr-1" />
                      Configure
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={(e) => handleDelete(agent.id, e)}
                      disabled={deleteMutation.isPending}
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </div>
                </motion.div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
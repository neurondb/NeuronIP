'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useAgent, useAgentPerformance, useDeployAgent } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { PlayIcon, ChartBarIcon } from '@heroicons/react/24/outline'
import { format } from 'date-fns'

interface AgentDetailProps {
  agentId: string
}

export default function AgentDetail({ agentId }: AgentDetailProps) {
  const { data: agent, isLoading } = useAgent(agentId)
  const { data: performance, isLoading: perfLoading } = useAgentPerformance(agentId)
  const deployMutation = useDeployAgent()

  const handleDeploy = async () => {
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
          <p className="text-muted-foreground">Loading agent details...</p>
        </CardContent>
      </Card>
    )
  }

  if (!agent) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Agent not found</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>{agent.name || agent.id}</CardTitle>
            {agent.status !== 'active' && agent.status !== 'deployed' && (
              <Button onClick={handleDeploy} disabled={deployMutation.isPending}>
                <PlayIcon className="h-5 w-5 mr-2" />
                Deploy
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <h4 className="text-sm font-semibold mb-2">Status</h4>
            <span
              className={`inline-block text-xs px-3 py-1 rounded ${
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
            <div>
              <h4 className="text-sm font-semibold mb-2">Description</h4>
              <p className="text-sm text-muted-foreground">{agent.description}</p>
            </div>
          )}

          {agent.config && (
            <div>
              <h4 className="text-sm font-semibold mb-2">Configuration</h4>
              <pre className="text-xs bg-muted p-3 rounded-lg overflow-x-auto">
                {JSON.stringify(agent.config, null, 2)}
              </pre>
            </div>
          )}

          {agent.created_at && (
            <div>
              <h4 className="text-sm font-semibold mb-2">Created</h4>
              <p className="text-sm text-muted-foreground">
                {format(new Date(agent.created_at), 'PPpp')}
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {performance && !perfLoading && (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <ChartBarIcon className="h-5 w-5" />
              <CardTitle>Performance</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {performance.total_executions !== undefined && (
                <div>
                  <div className="text-sm text-muted-foreground">Total Executions</div>
                  <div className="text-2xl font-bold">{performance.total_executions}</div>
                </div>
              )}
              {performance.success_rate !== undefined && (
                <div>
                  <div className="text-sm text-muted-foreground">Success Rate</div>
                  <div className="text-2xl font-bold">
                    {((performance.success_rate || 0) * 100).toFixed(1)}%
                  </div>
                </div>
              )}
              {performance.avg_response_time !== undefined && (
                <div>
                  <div className="text-sm text-muted-foreground">Avg Response</div>
                  <div className="text-2xl font-bold">
                    {(performance.avg_response_time || 0).toFixed(2)}s
                  </div>
                </div>
              )}
              {performance.total_tokens !== undefined && (
                <div>
                  <div className="text-sm text-muted-foreground">Total Tokens</div>
                  <div className="text-2xl font-bold">{performance.total_tokens?.toLocaleString() || 0}</div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
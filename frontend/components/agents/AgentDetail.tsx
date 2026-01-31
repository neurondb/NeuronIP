'use client'

import { PlayIcon, ChartBarIcon } from '@heroicons/react/24/outline'
import { format } from 'date-fns'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { showToast } from '@/components/ui/Toast'
import {
  useAgent,
  useAgentPerformance,
  useAgentRuns,
  useAgentMemory,
  useAgentEvaluations,
  useDeployAgent,
} from '@/lib/api/queries'


interface AgentDetailProps {
  agentId: string
}

export default function AgentDetail({ agentId }: AgentDetailProps) {
  const { data: agent, isLoading } = useAgent(agentId)
  const { data: performance, isLoading: perfLoading } = useAgentPerformance(agentId)
  const { data: runs } = useAgentRuns(agentId, 20)
  const { data: memory } = useAgentMemory(agentId, 50)
  const { data: evaluations } = useAgentEvaluations(agentId, 20)
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

      {Array.isArray(runs) && runs.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Recent Runs</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              {runs.slice(0, 10).map((run: { id: string; task?: string; status?: string; start_time?: string }) => (
                <li key={run.id} className="flex justify-between border-b border-border pb-2 last:border-0">
                  <span className="truncate">{run.task || run.id}</span>
                  <span className="text-muted-foreground">{run.status}</span>
                  {run.start_time && (
                    <span className="text-muted-foreground text-xs">{format(new Date(run.start_time), 'PP')}</span>
                  )}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      {Array.isArray(memory) && memory.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Memory</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              {memory.slice(0, 10).map((entry: { id: string; memory_key: string }) => (
                <li key={entry.id} className="border-b border-border pb-2 last:border-0">
                  <span className="font-medium">{entry.memory_key}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      {Array.isArray(evaluations) && evaluations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Evaluations</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              {evaluations.slice(0, 10).map((ev: { id: string; status?: string; score?: number; started_at?: string }) => (
                <li key={ev.id} className="flex justify-between border-b border-border pb-2 last:border-0">
                  <span>{ev.status}</span>
                  {ev.score != null && <span>Score: {ev.score}</span>}
                  {ev.started_at && (
                    <span className="text-muted-foreground text-xs">{format(new Date(ev.started_at), 'PP')}</span>
                  )}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
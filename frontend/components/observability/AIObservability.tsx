'use client'

import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { formatDistanceToNow } from 'date-fns'

interface RetrievalStats {
  avg_hit_rate: number
  avg_evidence_coverage: number
  avg_similarity: number
  avg_latency_ms: number
  total_retrievals: number
  low_hit_rate_count: number
}

interface HallucinationStats {
  total_signals: number
  avg_confidence: number
  avg_citation_accuracy: number
  avg_evidence_strength: number
  critical_count: number
  high_count: number
  medium_count: number
  low_count: number
}

interface AgentExecutionLog {
  id: string
  agent_id: string
  agent_run_id: string
  step_type: string
  tool_name?: string
  decision?: string
  latency_ms: number
  tokens_used?: number
  cost?: number
  timestamp: string
}

interface HallucinationSignal {
  id: string
  query_id?: string
  agent_run_id?: string
  confidence_score: number
  risk_level: string
  citation_accuracy: number
  evidence_strength: number
  flags: string[]
  created_at: string
}

export default function AIObservability() {
  const [activeTab, setActiveTab] = useState<'retrieval' | 'hallucination' | 'agent-logs' | 'cost'>('retrieval')
  const [timeRange, setTimeRange] = useState('24h')

  // Fetch retrieval stats
  const { data: retrievalStats } = useQuery({
    queryKey: ['retrieval-stats', timeRange],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/observability/retrieval/stats', {
        params: { time_range: timeRange },
      })
      return response.data as RetrievalStats
    },
    refetchInterval: 30000,
  })

  // Fetch hallucination stats
  const { data: hallucinationStats } = useQuery({
    queryKey: ['hallucination-stats', timeRange],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/observability/hallucination/stats', {
        params: { time_range: timeRange },
      })
      return response.data as HallucinationStats
    },
    enabled: activeTab === 'hallucination',
    refetchInterval: 30000,
  })

  // Fetch hallucination signals
  const { data: hallucinationSignals } = useQuery({
    queryKey: ['hallucination-signals', timeRange],
    queryFn: async () => {
      const response = await apiClient.get('/api/v1/observability/hallucination/signals', {
        params: { limit: 50 },
      })
      return response.data as HallucinationSignal[]
    },
    enabled: activeTab === 'hallucination',
    refetchInterval: 30000,
  })

  const getRiskBadge = (riskLevel: string) => {
    const variants: Record<string, 'default' | 'success' | 'error' | 'warning'> = {
      low: 'success',
      medium: 'warning',
      high: 'error',
      critical: 'error',
    }
    return (
      <Badge variant={variants[riskLevel] || 'default'} className="capitalize">
        {riskLevel}
      </Badge>
    )
  }

  return (
    <div className="space-y-6">
      {/* Time Range Selector */}
      <div className="flex items-center gap-2">
        <label className="text-sm font-medium">Time Range:</label>
        <select
          value={timeRange}
          onChange={(e) => setTimeRange(e.target.value)}
          className="p-2 border rounded"
        >
          <option value="1h">Last Hour</option>
          <option value="24h">Last 24 Hours</option>
          <option value="7d">Last 7 Days</option>
        </select>
      </div>

      {/* Tabs */}
      <div className="border-b">
        <div className="flex space-x-4">
          {['retrieval', 'hallucination', 'agent-logs', 'cost'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab as any)}
              className={`pb-2 px-1 border-b-2 capitalize ${
                activeTab === tab
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground'
              }`}
            >
              {tab.replace('-', ' ')}
            </button>
          ))}
        </div>
      </div>

      {/* Retrieval Metrics Tab */}
      {activeTab === 'retrieval' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Avg Hit Rate</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {retrievalStats?.avg_hit_rate
                    ? (retrievalStats.avg_hit_rate * 100).toFixed(1) + '%'
                    : 'N/A'}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Evidence Coverage</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {retrievalStats?.avg_evidence_coverage
                    ? (retrievalStats.avg_evidence_coverage * 100).toFixed(1) + '%'
                    : 'N/A'}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Avg Similarity</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {retrievalStats?.avg_similarity
                    ? (retrievalStats.avg_similarity * 100).toFixed(1) + '%'
                    : 'N/A'}
                </div>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Retrieval Performance</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div className="flex justify-between">
                  <span>Total Retrievals:</span>
                  <span className="font-medium">{retrievalStats?.total_retrievals || 0}</span>
                </div>
                <div className="flex justify-between">
                  <span>Avg Latency:</span>
                  <span className="font-medium">
                    {retrievalStats?.avg_latency_ms
                      ? (retrievalStats.avg_latency_ms / 1000).toFixed(2) + 's'
                      : 'N/A'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Low Hit Rate Count:</span>
                  <span className="font-medium text-yellow-600">
                    {retrievalStats?.low_hit_rate_count || 0}
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Hallucination Detection Tab */}
      {activeTab === 'hallucination' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Critical</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">
                  {hallucinationStats?.critical_count || 0}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">High</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-orange-600">
                  {hallucinationStats?.high_count || 0}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Medium</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-yellow-600">
                  {hallucinationStats?.medium_count || 0}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Low</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  {hallucinationStats?.low_count || 0}
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Avg Confidence</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {hallucinationStats?.avg_confidence
                    ? (hallucinationStats.avg_confidence * 100).toFixed(1) + '%'
                    : 'N/A'}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Citation Accuracy</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {hallucinationStats?.avg_citation_accuracy
                    ? (hallucinationStats.avg_citation_accuracy * 100).toFixed(1) + '%'
                    : 'N/A'}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Evidence Strength</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {hallucinationStats?.avg_evidence_strength
                    ? (hallucinationStats.avg_evidence_strength * 100).toFixed(1) + '%'
                    : 'N/A'}
                </div>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Recent Hallucination Signals</CardTitle>
            </CardHeader>
            <CardContent>
              {!hallucinationSignals || hallucinationSignals.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">No signals</div>
              ) : (
                <div className="space-y-3">
                  {hallucinationSignals.map((signal) => (
                    <div key={signal.id} className="p-4 border rounded-lg">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-2">
                            {getRiskBadge(signal.risk_level)}
                            <span className="text-sm text-muted-foreground">
                              Confidence: {(signal.confidence_score * 100).toFixed(1)}%
                            </span>
                          </div>
                          <div className="mt-2 space-y-1 text-sm">
                            <div>
                              Citation Accuracy: {(signal.citation_accuracy * 100).toFixed(1)}%
                            </div>
                            <div>
                              Evidence Strength: {(signal.evidence_strength * 100).toFixed(1)}%
                            </div>
                            {signal.flags && signal.flags.length > 0 && (
                              <div className="flex gap-1 mt-2">
                                {signal.flags.map((flag, idx) => (
                                  <Badge key={idx} variant="outline" className="text-xs">
                                    {flag}
                                  </Badge>
                                ))}
                              </div>
                            )}
                          </div>
                        </div>
                        <span className="text-xs text-muted-foreground">
                          {formatDistanceToNow(new Date(signal.created_at), { addSuffix: true })}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Agent Logs Tab */}
      {activeTab === 'agent-logs' && (
        <Card>
          <CardHeader>
            <CardTitle>Agent Execution Logs</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8 text-muted-foreground">
              Select an agent to view execution logs
            </div>
          </CardContent>
        </Card>
      )}

      {/* Cost Tab */}
      {activeTab === 'cost' && (
        <Card>
          <CardHeader>
            <CardTitle>Cost Tracking</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8 text-muted-foreground">
              Cost tracking per query and agent run
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

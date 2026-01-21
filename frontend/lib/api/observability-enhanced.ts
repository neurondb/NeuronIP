import apiClient from './client'

export interface LatencyMetrics {
  p50: number
  p95: number
  p99: number
  p999: number
  min: number
  max: number
  avg: number
}

export interface TokenUsage {
  total_tokens: number
  total_cost_usd: number
  avg_tokens_per_request: number
}

export interface EmbeddingCost {
  [model: string]: {
    total_tokens: number
    total_cost_usd: number
    request_count: number
  }
}

export async function getLatencyMetrics(
  endpoint: string,
  startTime?: string,
  endTime?: string
): Promise<LatencyMetrics> {
  const params = new URLSearchParams({ endpoint })
  if (startTime) params.set('start_time', startTime)
  if (endTime) params.set('end_time', endTime)

  const response = await apiClient.get(`/observability/metrics/latency?${params}`)
  return response.data
}

export async function getErrorRate(
  endpoint: string,
  startTime?: string,
  endTime?: string
): Promise<{ error_rate: number }> {
  const params = new URLSearchParams({ endpoint })
  if (startTime) params.set('start_time', startTime)
  if (endTime) params.set('end_time', endTime)

  const response = await apiClient.get(`/observability/metrics/error-rate?${params}`)
  return response.data
}

export async function getTokenUsage(
  userId?: string,
  startTime?: string,
  endTime?: string
): Promise<TokenUsage> {
  const params = new URLSearchParams()
  if (userId) params.set('user_id', userId)
  if (startTime) params.set('start_time', startTime)
  if (endTime) params.set('end_time', endTime)

  const response = await apiClient.get(`/observability/metrics/token-usage?${params}`)
  return response.data
}

export async function getEmbeddingCost(
  startTime?: string,
  endTime?: string
): Promise<EmbeddingCost> {
  const params = new URLSearchParams()
  if (startTime) params.set('start_time', startTime)
  if (endTime) params.set('end_time', endTime)

  const response = await apiClient.get(`/observability/metrics/embedding-cost?${params}`)
  return response.data
}

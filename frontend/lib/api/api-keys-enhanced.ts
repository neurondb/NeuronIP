import apiClient from './client'

export interface APIKey {
  id: string
  key?: string // Only shown once on creation
  key_prefix: string
  name?: string
  scopes: string[]
  rate_limit: number
  expires_at?: string
  rotation_enabled: boolean
  next_rotation_at?: string
  tags?: string[]
  created_at: string
}

export interface CreateAPIKeyRequest {
  name: string
  scopes?: string[]
  rate_limit?: number
  expires_at?: string
  rotation_enabled?: boolean
  rotation_interval_days?: number
  tags?: string[]
}

export interface UsageAnalytics {
  total_requests: number
  successful_requests: number
  failed_requests: number
  avg_response_time_ms?: number
  last_used_at?: string
}

export async function createAPIKey(request: CreateAPIKeyRequest): Promise<APIKey> {
  const response = await apiClient.post('/api-keys', request)
  return response.data
}

export async function rotateAPIKey(id: string, rotationType: string = 'manual'): Promise<APIKey> {
  const response = await apiClient.post(`/api-keys/${id}/rotate`, { rotation_type: rotationType })
  return response.data
}

export async function getAPIKeyUsage(
  id: string,
  startDate?: string,
  endDate?: string
): Promise<UsageAnalytics> {
  const params = new URLSearchParams()
  if (startDate) params.set('start_date', startDate)
  if (endDate) params.set('end_date', endDate)

  const response = await apiClient.get(`/api-keys/${id}/usage?${params}`)
  return response.data
}

export async function revokeAPIKey(id: string, reason?: string): Promise<void> {
  await apiClient.post(`/api-keys/${id}/revoke`, { reason })
}

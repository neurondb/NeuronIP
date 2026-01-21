import apiClient from './client'

export interface SavedSearch {
  id: string
  name: string
  description?: string
  query: string
  semantic_query?: string
  schema_id?: string
  sql_filters?: Record<string, any>
  semantic_table?: string
  semantic_column?: string
  limit: number
  threshold: number
  is_public: boolean
  owner_id?: string
  tags?: string[]
  created_at: string
  updated_at: string
}

export async function listSavedSearches(params?: {
  user_id?: string
  public_only?: boolean
}): Promise<SavedSearch[]> {
  const queryParams = new URLSearchParams()
  if (params?.user_id) queryParams.set('user_id', params.user_id)
  if (params?.public_only) queryParams.set('public_only', 'true')

  const response = await apiClient.get(`/warehouse/saved-searches?${queryParams}`)
  return response.data
}

export async function getSavedSearch(id: string): Promise<SavedSearch> {
  const response = await apiClient.get(`/warehouse/saved-searches/${id}`)
  return response.data
}

export async function createSavedSearch(search: Partial<SavedSearch>): Promise<SavedSearch> {
  const response = await apiClient.post('/warehouse/saved-searches', search)
  return response.data
}

export async function updateSavedSearch(id: string, search: Partial<SavedSearch>): Promise<void> {
  await apiClient.put(`/warehouse/saved-searches/${id}`, search)
}

export async function deleteSavedSearch(id: string): Promise<void> {
  await apiClient.delete(`/warehouse/saved-searches/${id}`)
}

export async function executeSavedSearch(id: string): Promise<any> {
  const response = await apiClient.post(`/warehouse/saved-searches/${id}/execute`)
  return response.data
}

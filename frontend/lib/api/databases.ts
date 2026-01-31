import apiClient from './client'

export interface Database {
  id: string
  name: string
  description?: string
  workspace_id?: string
  created_by?: string
  created_at: string
  updated_at: string
  metadata?: Record<string, any>
}

export interface DatabaseColumn {
  id: string
  database_id: string
  name: string
  type: 'text' | 'number' | 'date' | 'select' | 'multiSelect' | 'person' | 'file' | 'checkbox'
  options?: string[]
  order: number
  created_at: string
  updated_at: string
}

export interface DatabaseRow {
  id: string
  database_id: string
  data: Record<string, any>
  created_by?: string
  created_at: string
  updated_at: string
}

export interface ViewPreferences {
  view_type: 'table' | 'kanban' | 'calendar' | 'gallery' | 'list'
  filters?: Array<Record<string, any>>
  sort?: Record<string, any>
}

export interface CreateDatabaseRequest {
  name: string
  description?: string
  workspace_id?: string
  columns: Array<{
    name: string
    type: string
    options?: string[]
    order?: number
  }>
  metadata?: Record<string, any>
}

export interface UpdateRowRequest {
  data: Record<string, any>
}

export const databasesApi = {
  getDatabase: async (databaseId: string): Promise<{ database: Database; columns: DatabaseColumn[]; rows: DatabaseRow[] }> => {
    const response = await apiClient.get(`/databases/${databaseId}`)
    return response.data
  },

  createDatabase: async (data: CreateDatabaseRequest): Promise<{ database: Database; columns: DatabaseColumn[] }> => {
    const response = await apiClient.post('/databases', data)
    return response.data
  },

  updateRow: async (databaseId: string, rowId: string, data: UpdateRowRequest): Promise<{ row: DatabaseRow }> => {
    const response = await apiClient.patch(`/databases/${databaseId}/rows/${rowId}`, data)
    return response.data
  },

  createRow: async (databaseId: string, data: UpdateRowRequest): Promise<{ row: DatabaseRow }> => {
    const response = await apiClient.post(`/databases/${databaseId}/rows`, data)
    return response.data
  },

  deleteRow: async (databaseId: string, rowId: string): Promise<void> => {
    await apiClient.delete(`/databases/${databaseId}/rows/${rowId}`)
  },

  updateViewPreferences: async (databaseId: string, preferences: ViewPreferences): Promise<void> => {
    await apiClient.patch(`/databases/${databaseId}/view-preferences`, preferences)
  },

  getViewPreferences: async (databaseId: string): Promise<{ preferences: ViewPreferences }> => {
    const response = await apiClient.get(`/databases/${databaseId}/view-preferences`)
    return response.data
  },
}

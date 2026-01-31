import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

import { showToast } from '@/components/ui/Toast'

import { databasesApi, Database, DatabaseColumn, DatabaseRow, ViewPreferences, CreateDatabaseRequest, UpdateRowRequest } from '../api/databases'

export function useDatabase(databaseId: string | null) {
  return useQuery({
    queryKey: ['database', databaseId],
    queryFn: () => databasesApi.getDatabase(databaseId!),
    enabled: !!databaseId,
    staleTime: 30000, // 30 seconds
  })
}

export function useCreateDatabase() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateDatabaseRequest) => databasesApi.createDatabase(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['databases'] })
      showToast('Database created successfully', 'success')
    },
    onError: (error: any) => {
      showToast(error?.response?.data?.message || 'Failed to create database', 'error')
    },
  })
}

export function useUpdateRow(databaseId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ rowId, data }: { rowId: string; data: UpdateRowRequest }) =>
      databasesApi.updateRow(databaseId, rowId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['database', databaseId] })
      showToast('Row updated successfully', 'success')
    },
    onError: (error: any) => {
      showToast(error?.response?.data?.message || 'Failed to update row', 'error')
    },
  })
}

export function useCreateRow(databaseId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateRowRequest) => databasesApi.createRow(databaseId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['database', databaseId] })
      showToast('Row created successfully', 'success')
    },
    onError: (error: any) => {
      showToast(error?.response?.data?.message || 'Failed to create row', 'error')
    },
  })
}

export function useDeleteRow(databaseId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (rowId: string) => databasesApi.deleteRow(databaseId, rowId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['database', databaseId] })
      showToast('Row deleted successfully', 'success')
    },
    onError: (error: any) => {
      showToast(error?.response?.data?.message || 'Failed to delete row', 'error')
    },
  })
}

export function useUpdateViewPreferences(databaseId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (preferences: ViewPreferences) => databasesApi.updateViewPreferences(databaseId, preferences),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['database', databaseId, 'view-preferences'] })
    },
    onError: (error: any) => {
      showToast(error?.response?.data?.message || 'Failed to update view preferences', 'error')
    },
  })
}

export function useViewPreferences(databaseId: string) {
  return useQuery({
    queryKey: ['database', databaseId, 'view-preferences'],
    queryFn: () => databasesApi.getViewPreferences(databaseId),
    staleTime: 60000, // 1 minute
  })
}

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  databasesApi,
  Database,
  CreateDatabaseRequest,
  UpdateRowRequest,
  ViewPreferences,
} from '@/lib/api/databases'

export function useDatabase(databaseId: string) {
  return useQuery({
    queryKey: ['database', databaseId],
    queryFn: () => databasesApi.getDatabase(databaseId),
    enabled: !!databaseId,
  })
}

export function useCreateDatabase() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateDatabaseRequest) => databasesApi.createDatabase(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['databases'] })
    },
  })
}

export function useUpdateRow() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      databaseId,
      rowId,
      data,
    }: {
      databaseId: string
      rowId: string
      data: UpdateRowRequest
    }) => databasesApi.updateRow(databaseId, rowId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['database', variables.databaseId] })
    },
  })
}

export function useCreateRow() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      databaseId,
      data,
    }: {
      databaseId: string
      data: UpdateRowRequest
    }) => databasesApi.createRow(databaseId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['database', variables.databaseId] })
    },
  })
}

export function useDeleteRow() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      databaseId,
      rowId,
    }: {
      databaseId: string
      rowId: string
    }) => databasesApi.deleteRow(databaseId, rowId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['database', variables.databaseId] })
    },
  })
}

export function useUpdateViewPreferences() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      databaseId,
      preferences,
    }: {
      databaseId: string
      preferences: ViewPreferences
    }) => databasesApi.updateViewPreferences(databaseId, preferences),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['database', variables.databaseId] })
    },
  })
}

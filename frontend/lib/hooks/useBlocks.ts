import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { blocksApi, Block, CreateBlockRequest, UpdateBlockRequest } from '@/lib/api/blocks'

export function useBlocks(pageId: string) {
  return useQuery({
    queryKey: ['blocks', pageId],
    queryFn: () => blocksApi.getBlocks(pageId),
    enabled: !!pageId,
  })
}

export function useCreateBlock() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateBlockRequest) => blocksApi.createBlock(data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['blocks', variables.page_id] })
    },
  })
}

export function useUpdateBlock() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ blockId, data }: { blockId: string; data: UpdateBlockRequest }) =>
      blocksApi.updateBlock(blockId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['blocks'] })
    },
  })
}

export function useDeleteBlock() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ blockId, pageId }: { blockId: string; pageId: string }) =>
      blocksApi.deleteBlock(blockId),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['blocks', variables.pageId] })
    },
  })
}

export function useReorderBlocks() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ pageId, blockIds }: { pageId: string; blockIds: string[] }) =>
      blocksApi.reorderBlocks(pageId, blockIds),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['blocks', variables.pageId] })
    },
    onError: (error: any) => {
      console.error('Failed to reorder blocks:', error)
    },
  })
}

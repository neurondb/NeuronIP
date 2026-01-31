import apiClient from './client'

export interface Block {
  id: string
  type: string
  content: string | Record<string, any>
  order: number
  parent_id?: string
  page_id: string
  created_at: string
  updated_at: string
  metadata?: Record<string, any>
}

export interface CreateBlockRequest {
  type: string
  content: string | Record<string, any>
  order?: number
  parent_id?: string
  page_id: string
  metadata?: Record<string, any>
}

export interface UpdateBlockRequest {
  content?: string | Record<string, any>
  order?: number
  metadata?: Record<string, any>
}

export const blocksApi = {
  // Get all blocks for a page
  getBlocks: async (pageId: string): Promise<Block[]> => {
    const response = await apiClient.get(`/blocks?page_id=${pageId}`)
    return response.data.blocks || []
  },

  // Create a new block
  createBlock: async (data: CreateBlockRequest): Promise<Block> => {
    const response = await apiClient.post('/blocks', data)
    return response.data.block
  },

  // Update a block
  updateBlock: async (blockId: string, data: UpdateBlockRequest): Promise<Block> => {
    const response = await apiClient.patch(`/blocks/${blockId}`, data)
    return response.data.block
  },

  // Delete a block
  deleteBlock: async (blockId: string): Promise<void> => {
    await apiClient.delete(`/blocks/${blockId}`)
  },

  // Reorder blocks
  reorderBlocks: async (pageId: string, blockIds: string[]): Promise<void> => {
    await apiClient.post(`/blocks/reorder`, { page_id: pageId, block_ids: blockIds })
  },
}

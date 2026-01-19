import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from './client'
import { API_ENDPOINTS, QUERY_KEYS } from '../utils/constants'

// Types
interface SemanticSearchRequest {
  query: string
  collection_id?: string
  top_k?: number
}

interface SemanticRAGRequest {
  query: string
  collection_id?: string
  context_limit?: number
}

interface WarehouseQueryRequest {
  query: string
  schema_id?: string
  semantic_query?: string
  sql_filters?: Record<string, unknown>
}

interface ComplianceCheckRequest {
  entity_type: string
  entity_id: string
  entity_content: string
}

// Semantic Search Queries
export function useSemanticSearch() {
  return useMutation({
    mutationFn: async (data: SemanticSearchRequest) => {
      const response = await apiClient.post(API_ENDPOINTS.semanticSearch, data)
      return response.data
    },
  })
}

export function useSemanticRAG() {
  return useMutation({
    mutationFn: async (data: SemanticRAGRequest) => {
      const response = await apiClient.post(API_ENDPOINTS.semanticRAG, data)
      return response.data
    },
  })
}

export function useSemanticCollection(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.semanticCollection(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.semanticCollection(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useCreateDocument() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: FormData) => {
      const response = await apiClient.post(API_ENDPOINTS.semanticDocuments, data, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.semanticDocuments] })
    },
  })
}

// Warehouse Queries
export function useWarehouseQuery() {
  return useMutation({
    mutationFn: async (data: WarehouseQueryRequest) => {
      const response = await apiClient.post(API_ENDPOINTS.warehouseQuery, data)
      return response.data
    },
  })
}

export function useWarehouseSchemas() {
  return useQuery({
    queryKey: [QUERY_KEYS.warehouseSchemas],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.warehouseSchemas)
      return response.data
    },
  })
}

export function useWarehouseSchema(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.warehouseSchema(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.warehouseSchema(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

// Workflow Queries
export function useExecuteWorkflow(workflowId: string) {
  return useMutation({
    mutationFn: async (input: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.workflowExecute(workflowId), { input })
      return response.data
    },
  })
}

export function useWorkflow(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.workflow(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflow(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

// Compliance Queries
export function useComplianceCheck() {
  return useMutation({
    mutationFn: async (data: ComplianceCheckRequest) => {
      const response = await apiClient.post(API_ENDPOINTS.complianceCheck, data)
      return response.data
    },
  })
}

export function useComplianceAnomalies(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.complianceAnomalies, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.complianceAnomalies, { params: filters })
      return response.data
    },
  })
}

// Analytics Queries
export function useSearchAnalytics(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.analyticsSearch, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.analyticsSearch, { params: filters })
      return response.data
    },
  })
}

export function useWarehouseAnalytics(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.analyticsWarehouse, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.analyticsWarehouse, { params: filters })
      return response.data
    },
  })
}

export function useWorkflowAnalytics(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.analyticsWorkflows, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.analyticsWorkflows, { params: filters })
      return response.data
    },
  })
}

export function useComplianceAnalytics(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.analyticsCompliance, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.analyticsCompliance, { params: filters })
      return response.data
    },
  })
}

// Models Queries
export function useModels() {
  return useQuery({
    queryKey: [QUERY_KEYS.models],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.models)
      return response.data
    },
  })
}

export function useModel(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.model(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.model(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useRegisterModel() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.models, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.models] })
    },
  })
}

export function useInferModel(modelId: string) {
  return useMutation({
    mutationFn: async (input: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.modelInfer(modelId), { input })
      return response.data
    },
  })
}

// Knowledge Graph Queries
export function useExtractEntities() {
  return useMutation({
    mutationFn: async (data: { text: string; entity_types?: string[] }) => {
      const response = await apiClient.post(API_ENDPOINTS.knowledgeGraphExtractEntities, data)
      return response.data
    },
  })
}

export function useSearchEntities() {
  return useMutation({
    mutationFn: async (data: { query: string; entity_types?: string[] }) => {
      const response = await apiClient.post(API_ENDPOINTS.knowledgeGraphSearchEntities, data)
      return response.data
    },
  })
}

export function useTraverseGraph() {
  return useMutation({
    mutationFn: async (data: { start_entity_id: string; end_entity_id?: string; max_depth?: number }) => {
      const response = await apiClient.post(API_ENDPOINTS.knowledgeGraphTraverse, data)
      return response.data
    },
  })
}

export function useSearchGlossary() {
  return useMutation({
    mutationFn: async (data: { query: string; limit?: number }) => {
      const response = await apiClient.post(API_ENDPOINTS.knowledgeGraphSearchGlossary, data)
      return response.data
    },
  })
}

// Alerts Queries
export function useAlerts(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: ['alerts', filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.alerts, { params: filters })
      return response.data
    },
  })
}

export function useAlertRules() {
  return useQuery({
    queryKey: ['alert-rules'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.alertRules)
      return response.data
    },
  })
}

export function useCreateAlertRule() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.alertRules, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alert-rules'] })
    },
  })
}

export function useResolveAlert() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.post(API_ENDPOINTS.alertResolve(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] })
    },
  })
}

// API Keys Queries
export function useAPIKeys() {
  return useQuery({
    queryKey: ['api-keys'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.apiKeys)
      return response.data
    },
  })
}

export function useCreateAPIKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: { name: string; rate_limit?: number }) => {
      const response = await apiClient.post(API_ENDPOINTS.apiKeys, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
    },
  })
}

export function useDeleteAPIKey() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.delete(API_ENDPOINTS.apiKey(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
    },
  })
}

// Users Queries
export function useUsers(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: ['users', filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.users, { params: filters })
      return response.data
    },
  })
}

export function useUser(id: string, enabled = true) {
  return useQuery({
    queryKey: ['user', id],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.user(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

// Support Queries
export function useSupportTickets(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: ['support-tickets', filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.supportTickets, { params: filters })
      return response.data
    },
  })
}

export function useCreateSupportTicket() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.supportTickets, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['support-tickets'] })
    },
  })
}

export function useSupportTicket(id: string, enabled = true) {
  return useQuery({
    queryKey: ['support-ticket', id],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.supportTicket(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

// Health Check
export function useHealthCheck() {
  return useQuery({
    queryKey: ['health'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.health)
      return response.data
    },
    refetchInterval: 30000, // Check every 30 seconds
  })
}

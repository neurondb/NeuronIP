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
export function useWorkflows() {
  return useQuery({
    queryKey: [QUERY_KEYS.workflows],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflows)
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

export function useCreateWorkflow() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.workflows, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.workflows] })
    },
  })
}

export function useUpdateWorkflow(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.put(API_ENDPOINTS.workflow(id), data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.workflow(id) })
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.workflows] })
    },
  })
}

export function useDeleteWorkflow(id: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.delete(API_ENDPOINTS.workflow(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.workflows] })
    },
  })
}

export function useExecuteWorkflow(workflowId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (input: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.workflowExecute(workflowId), { input })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.workflowExecutionStatus(workflowId) })
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.workflowMonitoring(workflowId) })
    },
  })
}

export function useCreateWorkflowVersion(workflowId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: { version: string; changes: Record<string, unknown> }) => {
      const response = await apiClient.post(API_ENDPOINTS.workflowVersions(workflowId), data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.workflowVersions(workflowId) })
    },
  })
}

export function useWorkflowVersions(workflowId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.workflowVersions(workflowId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflowVersions(workflowId))
      return response.data
    },
    enabled: enabled && !!workflowId,
  })
}

export function useScheduleWorkflow(workflowId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (schedule: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.workflowSchedule(workflowId), schedule)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.workflowSchedules(workflowId) })
    },
  })
}

export function useWorkflowSchedules(workflowId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.workflowSchedules(workflowId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflowSchedules(workflowId))
      return response.data
    },
    enabled: enabled && !!workflowId,
  })
}

export function useCancelWorkflowSchedule(workflowId: string, scheduleId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async () => {
      const response = await apiClient.post(API_ENDPOINTS.workflowCancelSchedule(workflowId, scheduleId))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.workflowSchedules(workflowId) })
    },
  })
}

export function useRecoverWorkflowExecution(executionId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data?: { retry_from_step?: string }) => {
      const response = await apiClient.post(API_ENDPOINTS.workflowExecutionRecover(executionId), data || {})
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.workflowExecutionStatus(executionId) })
    },
  })
}

export function useWorkflowExecutionStatus(executionId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.workflowExecutionStatus(executionId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflowExecutionStatus(executionId))
      return response.data
    },
    enabled: enabled && !!executionId,
    refetchInterval: 2000, // Poll every 2 seconds for running executions
  })
}

export function useWorkflowExecutionLogs(executionId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.workflowExecutionLogs(executionId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflowExecutionLogs(executionId))
      return response.data
    },
    enabled: enabled && !!executionId,
  })
}

export function useWorkflowExecutionMetrics(executionId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.workflowExecutionMetrics(executionId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflowExecutionMetrics(executionId))
      return response.data
    },
    enabled: enabled && !!executionId,
  })
}

export function useWorkflowExecutionDecisions(executionId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.workflowExecutionDecisions(executionId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflowExecutionDecisions(executionId))
      return response.data
    },
    enabled: enabled && !!executionId,
  })
}

export function useWorkflowMonitoring(workflowId: string, timeRange = '24h', enabled = true) {
  return useQuery({
    queryKey: [...QUERY_KEYS.workflowMonitoring(workflowId), timeRange],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflowMonitoring(workflowId), {
        params: { time_range: timeRange },
      })
      return response.data
    },
    enabled: enabled && !!workflowId,
  })
}

export function useWorkflowVersion(workflowId: string, versionId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.workflowVersion(workflowId, versionId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.workflowVersion(workflowId, versionId))
      return response.data
    },
    enabled: enabled && !!workflowId && !!versionId,
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

// Metrics Queries
export function useMetrics(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.metrics, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.metrics, { params: filters })
      return response.data
    },
  })
}

export function useMetric(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.metric(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.metric(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useCreateMetric() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.metrics, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.metrics] })
    },
  })
}

export function useUpdateMetric() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Record<string, unknown> }) => {
      const response = await apiClient.put(API_ENDPOINTS.metric(id), data)
      return response.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.metrics] })
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.metric(variables.id) })
    },
  })
}

export function useDeleteMetric() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.delete(API_ENDPOINTS.metric(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.metrics] })
    },
  })
}

export function useSearchMetrics() {
  return useMutation({
    mutationFn: async (data: { query: string }) => {
      const response = await apiClient.post(API_ENDPOINTS.metricSearch, data)
      return response.data
    },
  })
}

export function useDiscoverMetrics() {
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.metricDiscover, data)
      return response.data
    },
  })
}

export function useCalculateMetric() {
  return useMutation({
    mutationFn: async ({ id, filters }: { id: string; filters?: Record<string, unknown> }) => {
      const response = await apiClient.post(API_ENDPOINTS.metricCalculate(id), { filters })
      return response.data
    },
  })
}

export function useMetricLineage(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.metricLineage(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.metricLineage(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useAddMetricLineage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Record<string, unknown> }) => {
      const response = await apiClient.post(API_ENDPOINTS.metricLineage(id), data)
      return response.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.metricLineage(variables.id) })
    },
  })
}

export function useApproveMetric() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, approvedBy }: { id: string; approvedBy: string }) => {
      const response = await apiClient.post(API_ENDPOINTS.metricApprove(id), { approved_by: approvedBy })
      return response.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.metrics] })
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.metric(variables.id) })
    },
  })
}

// Integrations Queries
export function useIntegrations(type?: string) {
  return useQuery({
    queryKey: ['integrations', type],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.integrations, {
        params: type ? { type } : {},
      })
      return response.data
    },
  })
}

export function useIntegration(id: string, enabled = true) {
  return useQuery({
    queryKey: ['integration', id],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.integration(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useCreateIntegration() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.integrations, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['integrations'] })
    },
  })
}

export function useUpdateIntegration() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Record<string, unknown> }) => {
      const response = await apiClient.put(API_ENDPOINTS.integration(id), data)
      return response.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['integrations'] })
      queryClient.invalidateQueries({ queryKey: ['integration', variables.id] })
    },
  })
}

export function useDeleteIntegration() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.delete(API_ENDPOINTS.integration(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['integrations'] })
    },
  })
}

export function useTestIntegration() {
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.post(API_ENDPOINTS.integrationTest(id))
      return response.data
    },
  })
}

export function useIntegrationHealth() {
  return useQuery({
    queryKey: ['integration-health'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.integrationHealth)
      return response.data
    },
    refetchInterval: 30000,
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

// Lineage Queries
export function useLineage(resourceType: string, resourceId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.lineage(resourceType, resourceId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.lineage(resourceType, resourceId))
      return response.data
    },
    enabled: enabled && !!resourceType && !!resourceId,
  })
}

export function useLineageGraph(enabled = true) {
  return useQuery({
    queryKey: [QUERY_KEYS.lineageGraph],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.lineageGraph)
      return response.data
    },
    enabled,
  })
}

export function useLineageImpact(resourceId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.lineageImpact(resourceId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.lineageImpact(resourceId))
      return response.data
    },
    enabled: enabled && !!resourceId,
  })
}

export function useTrackTransformation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: {
      source_id: string
      target_id: string
      edge_type: string
      transformation?: Record<string, unknown>
    }) => {
      const response = await apiClient.post(API_ENDPOINTS.lineageTrack, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.lineageGraph] })
    },
  })
}

// Column Lineage Queries
export function useColumnLineage(
  connectorId: string,
  schemaName: string,
  tableName: string,
  columnName: string,
  enabled = true
) {
  return useQuery({
    queryKey: QUERY_KEYS.columnLineage(connectorId, schemaName, tableName, columnName),
    queryFn: async () => {
      const response = await apiClient.get(
        API_ENDPOINTS.columnLineage(connectorId, schemaName, tableName, columnName)
      )
      return response.data
    },
    enabled: enabled && !!connectorId && !!schemaName && !!tableName && !!columnName,
  })
}

export function useTrackColumnLineage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: {
      source_connector_id?: string
      source_schema_name: string
      source_table_name: string
      source_column_name: string
      target_connector_id?: string
      target_schema_name: string
      target_table_name: string
      target_column_name: string
      edge_type: string
      transformation?: Record<string, unknown>
    }) => {
      const response = await apiClient.post(API_ENDPOINTS.columnLineageTrack, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.columnLineageGraph] })
    },
  })
}

// Data Quality Queries
export function useQualityDashboard() {
  return useQuery({
    queryKey: [QUERY_KEYS.dataQualityDashboard],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.dataQualityDashboard)
      return response.data
    },
  })
}

export function useQualityTrends(level: string, days: number) {
  return useQuery({
    queryKey: QUERY_KEYS.dataQualityTrends(level, days),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.dataQualityTrends, {
        params: { level, days },
      })
      return response.data
    },
  })
}

// Catalog Queries
export function useCatalogDatasets(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.catalogDatasets, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.catalogDatasets, { params: filters })
      return response.data
    },
  })
}

export function useCatalogDataset(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.catalogDataset(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.catalogDataset(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useCatalogSearch() {
  return useMutation({
    mutationFn: async (query: string) => {
      const response = await apiClient.get(API_ENDPOINTS.catalogSearch, { params: { query } })
      return response.data
    },
  })
}

export function useCatalogOwners() {
  return useQuery({
    queryKey: ['catalog-owners'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.catalogOwners)
      return response.data
    },
  })
}

export function useCatalogDiscover() {
  return useMutation({
    mutationFn: async (tags: string[]) => {
      const response = await apiClient.post(API_ENDPOINTS.catalogDiscover, { tags })
      return response.data
    },
  })
}

// Versioning Queries
export function useVersions(resourceType: string, resourceId: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.versions(resourceType, resourceId),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.versions(resourceType, resourceId))
      return response.data
    },
    enabled: enabled && !!resourceType && !!resourceId,
  })
}

export function useVersion(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.version(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.version(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useVersionHistory(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.versionHistory(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.versionHistory(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useCreateVersion() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: {
      resource_type: string
      resource_id: string
      version_number: string
      version_data: Record<string, unknown>
    }) => {
      const response = await apiClient.post(API_ENDPOINTS.versionCreate, data)
      return response.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: QUERY_KEYS.versions(variables.resource_type, variables.resource_id),
      })
    },
  })
}

export function useRollbackVersion() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.post(API_ENDPOINTS.versionRollback(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['versions'] })
    },
  })
}

// Audit Queries
export function useAuditEvents(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.auditEvents, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.auditEvents, { params: filters })
      return response.data
    },
  })
}

export function useAuditActivity(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.auditActivity, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.auditActivity, { params: filters })
      return response.data
    },
  })
}

export function useComplianceTrail(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.auditComplianceTrail, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.auditComplianceTrail, { params: filters })
      return response.data
    },
  })
}

export function useSearchAuditEvents() {
  return useMutation({
    mutationFn: async (data: { query: string; limit?: number }) => {
      const response = await apiClient.post(API_ENDPOINTS.auditSearch, data)
      return response.data
    },
  })
}

// Billing Queries
export function useBillingUsage(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.billingUsage, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.billingUsage, { params: filters })
      return response.data
    },
  })
}

export function useBillingMetrics(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: [QUERY_KEYS.billingMetrics, filters],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.billingMetrics, { params: filters })
      return response.data
    },
  })
}

export function useBillingDashboard() {
  return useQuery({
    queryKey: [QUERY_KEYS.billingDashboard],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.billingDashboard)
      return response.data
    },
  })
}

export function useTrackUsage() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: {
      metric_type: string
      metric_name: string
      value: number
      user_id?: string
      timestamp?: string
    }) => {
      const response = await apiClient.post(API_ENDPOINTS.billingTrack, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.billingUsage] })
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.billingDashboard] })
    },
  })
}

// Agents Queries
export function useAgents() {
  return useQuery({
    queryKey: [QUERY_KEYS.agents],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.agents)
      return response.data
    },
  })
}

export function useAgent(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.agent(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.agent(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useAgentPerformance(id: string, enabled = true) {
  return useQuery({
    queryKey: QUERY_KEYS.agentPerformance(id),
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.agentPerformance(id))
      return response.data
    },
    enabled: enabled && !!id,
  })
}

export function useCreateAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (data: Record<string, unknown>) => {
      const response = await apiClient.post(API_ENDPOINTS.agents, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.agents] })
    },
  })
}

export function useUpdateAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Record<string, unknown> }) => {
      const response = await apiClient.put(API_ENDPOINTS.agent(id), data)
      return response.data
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.agents] })
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.agent(variables.id) })
    },
  })
}

export function useDeleteAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.delete(API_ENDPOINTS.agent(id))
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.agents] })
    },
  })
}

export function useDeployAgent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const response = await apiClient.post(API_ENDPOINTS.agentDeploy(id))
      return response.data
    },
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.agents] })
      queryClient.invalidateQueries({ queryKey: QUERY_KEYS.agent(id) })
    },
  })
}

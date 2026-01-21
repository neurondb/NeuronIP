export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'

export const API_ENDPOINTS = {
  // Health
  health: '/health',

  // Semantic
  semanticSearch: '/semantic/search',
  semanticRAG: '/semantic/rag',
  semanticDocuments: '/semantic/documents',
  semanticCollection: (id: string) => `/semantic/collections/${id}`,

  // Warehouse
  warehouseQuery: '/warehouse/query',
  warehouseQueryById: (id: string) => `/warehouse/queries/${id}`,
  warehouseQueryHistory: '/warehouse/queries/history',
  warehouseOptimize: '/warehouse/optimize',
  warehouseSchemas: '/warehouse/schemas',
  warehouseSchema: (id: string) => `/warehouse/schemas/${id}`,

  // Workflows
  workflows: '/workflows',
  workflow: (id: string) => `/workflows/${id}`,
  workflowExecute: (id: string) => `/workflows/${id}/execute`,
  workflowVersions: (id: string) => `/workflows/${id}/versions`,
  workflowVersion: (id: string, versionId: string) => `/workflows/${id}/versions/${versionId}`,
  workflowSchedule: (id: string) => `/workflows/${id}/schedule`,
  workflowSchedules: (id: string) => `/workflows/${id}/schedules`,
  workflowCancelSchedule: (id: string, scheduleId: string) => `/workflows/${id}/schedules/${scheduleId}/cancel`,
  workflowExecutionStatus: (id: string) => `/workflows/executions/${id}/status`,
  workflowExecutionRecover: (id: string) => `/workflows/executions/${id}/recover`,
  workflowExecutionLogs: (id: string) => `/workflows/executions/${id}/logs`,
  workflowExecutionMetrics: (id: string) => `/workflows/executions/${id}/metrics`,
  workflowExecutionDecisions: (id: string) => `/workflows/executions/${id}/decisions`,
  workflowMonitoring: (id: string) => `/workflows/${id}/monitoring`,

  // Compliance
  complianceCheck: '/compliance/check',
  complianceAnomalies: '/compliance/anomalies',
  compliancePolicies: '/compliance/policies',
  compliancePolicy: (id: string) => `/compliance/policies/${id}`,
  complianceReport: '/compliance/report',

  // Analytics
  analyticsSearch: '/analytics/search',
  analyticsWarehouse: '/analytics/warehouse',
  analyticsWorkflows: '/analytics/workflows',
  analyticsCompliance: '/analytics/compliance',
  analyticsRetrievalQuality: '/analytics/retrieval-quality',

  // Models
  models: '/models',
  model: (id: string) => `/models/${id}`,
  modelInfer: (id: string) => `/models/${id}/infer`,

  // Knowledge Graph
  knowledgeGraphExtractEntities: '/knowledge-graph/entities/extract',
  knowledgeGraphEntity: (id: string) => `/knowledge-graph/entities/${id}`,
  knowledgeGraphEntityLinks: (id: string) => `/knowledge-graph/entities/${id}/links`,
  knowledgeGraphSearchEntities: '/knowledge-graph/entities/search',
  knowledgeGraphLinkEntities: '/knowledge-graph/entities/link',
  knowledgeGraphTraverse: '/knowledge-graph/traverse',
  knowledgeGraphEntityTypes: '/knowledge-graph/entity-types',
  knowledgeGraphGlossary: '/knowledge-graph/glossary',
  knowledgeGraphGlossaryTerm: (id: string) => `/knowledge-graph/glossary/${id}`,
  knowledgeGraphSearchGlossary: '/knowledge-graph/glossary/search',

  // Alerts
  alertsCheck: '/alerts/check',
  alerts: '/alerts',
  alertResolve: (id: string) => `/alerts/${id}/resolve`,
  alertRules: '/alerts/rules',
  alertRule: (id: string) => `/alerts/rules/${id}`,

  // API Keys
  apiKeys: '/api-keys',
  apiKey: (id: string) => `/api-keys/${id}`,

  // Users
  users: '/users',
  user: (id: string) => `/users/${id}`,

  // Support
  supportTickets: '/support/tickets',
  supportTicket: (id: string) => `/support/tickets/${id}`,
  supportTicketConversations: (id: string) => `/support/tickets/${id}/conversations`,
  supportTicketSimilarCases: (id: string) => `/support/tickets/${id}/similar-cases`,

  // Integrations
  integrations: '/integrations',
  integration: (id: string) => `/integrations/${id}`,
  integrationTest: (id: string) => `/integrations/${id}/test`,
  integrationHealth: '/integrations/health',
  integrationsHelpdeskSync: '/integrations/helpdesk/sync',
  webhooks: '/webhooks',
  webhook: (id: string) => `/webhooks/${id}`,
  webhookTrigger: (id: string) => `/webhooks/${id}/trigger`,

  // Observability
  observabilityQueryPerformance: '/observability/queries/performance',
  observabilityLogs: '/observability/logs',
  observabilityLogsStream: '/observability/logs/stream',
  observabilityMetrics: '/observability/metrics',
  observabilityRealtime: '/observability/realtime',
  observabilityBenchmark: '/observability/benchmark',
  observabilityCostBreakdown: '/observability/cost/breakdown',
  observabilityAgentLogs: '/observability/agent-logs',
  observabilityWorkflowLogs: '/observability/workflow-logs',

  // Metrics
  metrics: '/metrics',
  metric: (id: string) => `/metrics/${id}`,
  metricSearch: '/metrics/search',
  metricDiscover: '/metrics/discover',
  metricCalculate: (id: string) => `/metrics/${id}/calculate`,
  metricLineage: (id: string) => `/metrics/${id}/lineage`,
  metricApprove: (id: string) => `/metrics/${id}/approve`,

  // Lineage
  lineage: (resourceType: string, resourceId: string) => `/lineage/${resourceType}/${resourceId}`,
  lineageTrack: '/lineage/track',
  lineageImpact: (resourceId: string) => `/lineage/impact/${resourceId}`,
  lineageGraph: '/lineage/graph',
  columnLineage: (connectorId: string, schemaName: string, tableName: string, columnName: string) =>
    `/lineage/columns/${connectorId}/${schemaName}/${tableName}/${columnName}`,
  columnLineageTrack: '/lineage/columns/track',
  columnLineageNodes: '/lineage/columns/nodes',

  // Data Quality
  dataQualityRules: '/data-quality/rules',
  dataQualityRule: (id: string) => `/data-quality/rules/${id}`,
  dataQualityExecute: (id: string) => `/data-quality/rules/${id}/execute`,
  dataQualityDashboard: '/data-quality/dashboard',
  dataQualityTrends: '/data-quality/trends',

  // Catalog
  catalogDatasets: '/catalog/datasets',
  catalogDataset: (id: string) => `/catalog/datasets/${id}`,
  catalogSearch: '/catalog/search',
  catalogOwners: '/catalog/owners',
  catalogDiscover: '/catalog/discover',

  // Versioning
  versions: (resourceType: string, resourceId: string) => `/versions/${resourceType}/${resourceId}`,
  versionCreate: '/versions/create',
  version: (id: string) => `/versions/${id}`,
  versionRollback: (id: string) => `/versions/${id}/rollback`,
  versionHistory: (id: string) => `/versions/${id}/history`,

  // Audit
  auditEvents: '/audit/events',
  auditActivity: '/audit/activity',
  auditComplianceTrail: '/audit/compliance-trail',
  auditSearch: '/audit/search',

  // Billing
  billingUsage: '/billing/usage',
  billingMetrics: '/billing/metrics',
  billingDashboard: '/billing/dashboard',
  billingTrack: '/billing/track',

  // Agents
  agents: '/agents',
  agent: (id: string) => `/agents/${id}`,
  agentPerformance: (id: string) => `/agents/${id}/performance`,
  agentDeploy: (id: string) => `/agents/${id}/deploy`,
} as const

export const QUERY_KEYS = {
  // Semantic
  semanticSearch: 'semantic-search',
  semanticCollection: (id: string) => ['semantic-collection', id],
  semanticDocuments: 'semantic-documents',

  // Warehouse
  warehouseQueries: 'warehouse-queries',
  warehouseQuery: (id: string) => ['warehouse-query', id],
  warehouseSchemas: 'warehouse-schemas',
  warehouseSchema: (id: string) => ['warehouse-schema', id],

  // Workflows
  workflows: 'workflows',
  workflow: (id: string) => ['workflow', id],
  workflowVersions: (id: string) => ['workflow-versions', id],
  workflowVersion: (id: string, versionId: string) => ['workflow-version', id, versionId],
  workflowSchedules: (id: string) => ['workflow-schedules', id],
  workflowExecutionStatus: (id: string) => ['workflow-execution-status', id],
  workflowExecutionLogs: (id: string) => ['workflow-execution-logs', id],
  workflowExecutionMetrics: (id: string) => ['workflow-execution-metrics', id],
  workflowExecutionDecisions: (id: string) => ['workflow-execution-decisions', id],
  workflowMonitoring: (id: string) => ['workflow-monitoring', id],
  workflowExecutions: (id: string) => ['workflow-executions', id],

  // Compliance
  complianceAnomalies: 'compliance-anomalies',

  // Analytics
  analyticsSearch: 'analytics-search',
  analyticsWarehouse: 'analytics-warehouse',
  analyticsWorkflows: 'analytics-workflows',
  analyticsCompliance: 'analytics-compliance',
  analyticsRetrievalQuality: 'analytics-retrieval-quality',

  // Models
  models: 'models',
  model: (id: string) => ['model', id],

  // Knowledge Graph
  knowledgeGraphEntities: 'knowledge-graph-entities',
  knowledgeGraphEntity: (id: string) => ['knowledge-graph-entity', id],
  knowledgeGraphGlossary: 'knowledge-graph-glossary',

  // Alerts
  alerts: 'alerts',
  alertRules: 'alert-rules',

  // API Keys
  apiKeys: 'api-keys',

  // Users
  users: 'users',
  user: (id: string) => ['user', id],

  // Support
  supportTickets: 'support-tickets',
  supportTicket: (id: string) => ['support-ticket', id],

  // Metrics
  metrics: 'metrics',
  metric: (id: string) => ['metric', id],
  metricLineage: (id: string) => ['metric-lineage', id],

  // Lineage
  lineage: (resourceType: string, resourceId: string) => ['lineage', resourceType, resourceId],
  lineageGraph: 'lineage-graph',
  lineageImpact: (resourceId: string) => ['lineage-impact', resourceId],
  columnLineage: (connectorId: string, schemaName: string, tableName: string, columnName: string) =>
    ['column-lineage', connectorId, schemaName, tableName, columnName],
  columnLineageGraph: 'column-lineage-graph',

  // Data Quality
  dataQualityDashboard: 'data-quality-dashboard',
  dataQualityTrends: (level: string, days: number) => ['data-quality-trends', level, days],

  // Catalog
  catalogDatasets: 'catalog-datasets',
  catalogDataset: (id: string) => ['catalog-dataset', id],

  // Versioning
  versions: (resourceType: string, resourceId: string) => ['versions', resourceType, resourceId],
  version: (id: string) => ['version', id],
  versionHistory: (id: string) => ['version-history', id],

  // Audit
  auditEvents: 'audit-events',
  auditActivity: 'audit-activity',
  auditComplianceTrail: 'audit-compliance-trail',

  // Billing
  billingUsage: 'billing-usage',
  billingMetrics: 'billing-metrics',
  billingDashboard: 'billing-dashboard',

  // Agents
  agents: 'agents',
  agent: (id: string) => ['agent', id],
  agentPerformance: (id: string) => ['agent-performance', id],
} as const

export const STORAGE_KEYS = {
  theme: 'neuronip-theme',
  sidebarCollapsed: 'neuronip-sidebar-collapsed',
  userPreferences: 'neuronip-user-preferences',
} as const

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
  warehouseSchemas: '/warehouse/schemas',
  warehouseSchema: (id: string) => `/warehouse/schemas/${id}`,

  // Workflows
  workflowExecute: (id: string) => `/workflows/${id}/execute`,
  workflow: (id: string) => `/workflows/${id}`,

  // Compliance
  complianceCheck: '/compliance/check',
  complianceAnomalies: '/compliance/anomalies',

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
  integrationsHelpdeskSync: '/integrations/helpdesk/sync',
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
  workflow: (id: string) => ['workflow', id],
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
} as const

export const STORAGE_KEYS = {
  theme: 'neuronip-theme',
  sidebarCollapsed: 'neuronip-sidebar-collapsed',
  userPreferences: 'neuronip-user-preferences',
} as const

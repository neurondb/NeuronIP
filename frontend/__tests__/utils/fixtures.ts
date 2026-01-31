// Test data fixtures

export const mockAgents = [
  {
    id: '1',
    name: 'Data Analyst Agent',
    status: 'active',
    description: 'Analyzes data and generates insights',
    created_at: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    name: 'Support Agent',
    status: 'inactive',
    description: 'Handles customer support queries',
    created_at: '2024-01-02T00:00:00Z',
  },
]

export const mockWorkflows = [
  {
    id: '1',
    name: 'Data Pipeline',
    status: 'active',
    description: 'ETL pipeline for data processing',
    created_at: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    name: 'Report Generator',
    status: 'draft',
    description: 'Generates weekly reports',
    created_at: '2024-01-02T00:00:00Z',
  },
]

export const mockModels = [
  {
    id: '1',
    name: 'GPT-4',
    type: 'llm',
    status: 'active',
    description: 'Large language model',
  },
  {
    id: '2',
    name: 'text-embedding-ada-002',
    type: 'embedding',
    status: 'active',
    description: 'Text embedding model',
  },
]

export const mockDataSources = [
  {
    id: '1',
    name: 'Production DB',
    type: 'postgresql',
    status: 'connected',
    connection_string: 'postgresql://...',
  },
  {
    id: '2',
    name: 'Data Warehouse',
    type: 'snowflake',
    status: 'connected',
    connection_string: 'snowflake://...',
  },
]

export const mockMetrics = [
  {
    id: '1',
    name: 'Monthly Revenue',
    category: 'financial',
    status: 'approved',
    definition: 'Total revenue for the month',
  },
  {
    id: '2',
    name: 'Active Users',
    category: 'product',
    status: 'draft',
    definition: 'Number of active users',
  },
]

export const mockAlerts = [
  {
    id: '1',
    message: 'High error rate detected',
    severity: 'high',
    status: 'active',
    created_at: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    message: 'Low disk space',
    severity: 'medium',
    status: 'resolved',
    created_at: '2024-01-02T00:00:00Z',
  },
]

export const mockSupportTickets = [
  {
    id: '1',
    subject: 'Login issue',
    status: 'open',
    priority: 'high',
    created_at: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    subject: 'Feature request',
    status: 'closed',
    priority: 'low',
    created_at: '2024-01-02T00:00:00Z',
  },
]

export const mockQueryResults = [
  { id: 1, name: 'Item 1', value: 100 },
  { id: 2, name: 'Item 2', value: 200 },
  { id: 3, name: 'Item 3', value: 300 },
]

export const mockKnowledgeGraphEntities = [
  {
    id: '1',
    name: 'John Doe',
    type: 'person',
    properties: { email: 'john@example.com' },
  },
  {
    id: '2',
    name: 'Acme Corp',
    type: 'organization',
    properties: { industry: 'Technology' },
  },
]

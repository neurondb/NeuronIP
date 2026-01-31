import { http, HttpResponse } from 'msw'
import { setupServer } from 'msw/node'

// Note: This file is imported in setup.ts which already sets up the server

// Mock API responses
export const handlers = [
  // Warehouse endpoints
  http.post('/api/v1/warehouse/query', () => {
    return HttpResponse.json({
      results: [
        { id: 1, name: 'Test', value: 100 },
        { id: 2, name: 'Test 2', value: 200 },
      ],
      query_id: 'test-query-id',
    })
  }),

  // Semantic/RAG endpoints
  http.post('/api/v1/rag/query', () => {
    return HttpResponse.json({
      response: 'This is a test response',
      sources: [],
    })
  }),

  // Agents endpoints
  http.get('/api/v1/agents', () => {
    return HttpResponse.json([
      { id: '1', name: 'Agent 1', status: 'active' },
      { id: '2', name: 'Agent 2', status: 'inactive' },
    ])
  }),

  // Models endpoints
  http.get('/api/v1/models', () => {
    return HttpResponse.json([
      { id: '1', name: 'Model 1', type: 'llm' },
      { id: '2', name: 'Model 2', type: 'embedding' },
    ])
  }),

  // Workflows endpoints
  http.get('/api/v1/workflows', () => {
    return HttpResponse.json([
      { id: '1', name: 'Workflow 1', status: 'active' },
      { id: '2', name: 'Workflow 2', status: 'draft' },
    ])
  }),

  // Compliance endpoints
  http.get('/api/v1/compliance/policies', () => {
    return HttpResponse.json([
      { id: '1', name: 'Policy 1', type: 'data_quality' },
    ])
  }),

  // Data sources endpoints
  http.get('/api/v1/data-sources', () => {
    return HttpResponse.json([
      { id: '1', name: 'PostgreSQL', type: 'postgresql' },
      { id: '2', name: 'S3 Bucket', type: 's3' },
    ])
  }),

  // Metrics endpoints
  http.get('/api/v1/metrics', () => {
    return HttpResponse.json([
      { id: '1', name: 'Revenue', category: 'financial' },
      { id: '2', name: 'Users', category: 'product' },
    ])
  }),

  // Alerts endpoints
  http.get('/api/v1/alerts', () => {
    return HttpResponse.json([
      { id: '1', message: 'Alert 1', severity: 'high' },
    ])
  }),

  // Support endpoints
  http.get('/api/v1/support/tickets', () => {
    return HttpResponse.json([
      { id: '1', subject: 'Ticket 1', status: 'open' },
    ])
  }),

  // Knowledge graph endpoints
  http.post('/api/v1/knowledge-graph/entities/search', () => {
    return HttpResponse.json({
      entities: [
        { id: '1', name: 'Entity 1', type: 'person' },
      ],
    })
  }),

  // Auth endpoints
  http.post('/api/v1/auth/login', () => {
    return HttpResponse.json({
      token: 'mock-token',
      user: { id: '1', email: 'test@example.com' },
    })
  }),
]

export const server = setupServer(...handlers)

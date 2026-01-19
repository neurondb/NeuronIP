/**
 * NeuronIP JavaScript SDK - Basic Usage Example
 */

import NeuronIPClient from '@neuronip/sdk'

// Initialize client
const client = new NeuronIPClient(
  'http://localhost:8082/api/v1',
  'your-api-key-here'
)

// Health check
const health = await client.healthCheck()
console.log(`API Status: ${health.status}`)

// Semantic search
const searchResults = await client.semanticSearch('machine learning algorithms', 10)
console.log(`Found ${searchResults.results?.length || 0} results`)

// Warehouse query
const queryResult = await client.warehouseQuery('Show me total sales by region')
console.log(`Query executed: ${queryResult.query_id}`)

// Create ingestion job
const job = await client.createIngestionJob(
  'your-data-source-id',
  'sync',
  { mode: 'full', tables: ['users', 'orders'] }
)
console.log(`Created job: ${job.id}`)

// List ingestion jobs
const jobs = await client.listIngestionJobs(undefined, 10)
console.log(`Found ${jobs.length} jobs`)

// Get audit logs
const auditLogs = await client.getAuditLogs(undefined, 20)
console.log(`Retrieved ${auditLogs.length} audit log entries`)

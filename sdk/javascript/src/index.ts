/**
 * NeuronIP JavaScript/TypeScript SDK
 * Enterprise Intelligence Platform SDK
 */

export class NeuronIPClient {
  private baseUrl: string
  private apiKey?: string

  constructor(baseUrl: string, apiKey?: string) {
    this.baseUrl = baseUrl.replace(/\/$/, '')
    this.apiKey = apiKey
  }

  private async request<T>(
    method: string,
    endpoint: string,
    body?: any
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (this.apiKey) {
      headers['Authorization'] = `Bearer ${this.apiKey}`
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    })

    if (!response.ok) {
      throw new Error(`API request failed: ${response.statusText}`)
    }

    return response.json()
  }

  async healthCheck(): Promise<{ status: string }> {
    return this.request('GET', '/health')
  }

  async semanticSearch(query: string, limit: number = 10): Promise<any> {
    return this.request('POST', '/semantic/search', { query, limit })
  }

  async warehouseQuery(
    query: string,
    schemaId?: string
  ): Promise<any> {
    return this.request('POST', '/warehouse/query', {
      query,
      schema_id: schemaId,
    })
  }

  async createIngestionJob(
    dataSourceId: string,
    jobType: string,
    config: Record<string, any>
  ): Promise<any> {
    return this.request('POST', '/ingestion/jobs', {
      data_source_id: dataSourceId,
      job_type: jobType,
      config,
    })
  }

  async listIngestionJobs(
    dataSourceId?: string,
    limit: number = 100
  ): Promise<any[]> {
    const params = new URLSearchParams({ limit: limit.toString() })
    if (dataSourceId) {
      params.append('data_source_id', dataSourceId)
    }
    return this.request('GET', `/ingestion/jobs?${params}`)
  }

  async getMetric(metricId: string): Promise<any> {
    return this.request('GET', `/metrics/${metricId}`)
  }

  async createMetric(metric: Record<string, any>): Promise<any> {
    return this.request('POST', '/metrics', metric)
  }

  async getAuditLogs(
    filters?: Record<string, any>,
    limit: number = 100
  ): Promise<any[]> {
    const params = new URLSearchParams({ limit: limit.toString() })
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, String(value))
        }
      })
    }
    return this.request('GET', `/audit/events?${params}`)
  }
}

export default NeuronIPClient

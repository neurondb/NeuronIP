// API Types

export interface CheckStatus {
  status: 'healthy' | 'error' | 'warning' | 'unavailable'
  message?: string
}

export interface HealthResponse {
  status: 'ok' | 'unhealthy' | 'warning'
  service: string
  timestamp: string
  checks?: {
    database?: CheckStatus
    mcp?: CheckStatus
    uptime?: CheckStatus
    [key: string]: CheckStatus | undefined
  }
}

export interface SystemStats {
  status: 'healthy' | 'warning' | 'error'
  database: 'connected' | 'disconnected'
  mcp: 'connected' | 'disconnected' | 'unavailable'
  api: 'online' | 'offline'
  uptime?: string
}

// Database connection testing utilities

// Use 127.0.0.1 instead of localhost for better compatibility
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://127.0.0.1:8082/api/v1'

export interface DatabaseConnection {
  host: string
  port: string
  database: string
  user: string
  password: string
  sslMode?: string
}

export interface DatabaseTestResponse {
  success: boolean
  message: string
  latency_ms?: number
  version?: string
}

/**
 * Test a PostgreSQL database connection
 */
export async function testDatabaseConnection(
  connection: DatabaseConnection
): Promise<DatabaseTestResponse> {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 10000) // 10 second timeout

  const url = `${API_BASE_URL}/database/test`
  
  try {
    console.log('Testing database connection to:', url)
    console.log('Connection config:', { 
      host: connection.host, 
      port: connection.port, 
      database: connection.database,
      user: connection.user 
    })

    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include', // Include cookies for CORS
      signal: controller.signal,
      body: JSON.stringify({
        host: connection.host,
        port: connection.port,
        database: connection.database,
        user: connection.user,
        password: connection.password,
        ssl_mode: connection.sslMode || 'disable',
      }),
    })

    clearTimeout(timeoutId)

    console.log('Response status:', response.status, response.statusText)
    console.log('Response headers:', Object.fromEntries(response.headers.entries()))

    if (!response.ok) {
      let errorData
      try {
        errorData = await response.json()
      } catch {
        errorData = { error: { message: `Connection test failed: ${response.status} ${response.statusText}` } }
      }
      const errorMessage = errorData.error?.message || errorData.message || `Connection test failed: ${response.status} ${response.statusText}`
      console.error('Connection test failed:', errorMessage, errorData)
      throw new Error(errorMessage)
    }

    const result = await response.json()
    console.log('Connection test successful:', result)
    return result
  } catch (error: any) {
    clearTimeout(timeoutId)
    
    console.error('Connection test error:', error)
    
    // Handle network errors
    if (error.name === 'AbortError') {
      throw new Error('Connection test timed out. Please check your network connection and try again.')
    }
    if (error.name === 'TypeError' && (error.message.includes('fetch') || error.message.includes('Failed to fetch'))) {
      const apiUrl = API_BASE_URL.replace('/api/v1', '')
      throw new Error(`Failed to connect to API server. Please ensure the API server is running on ${apiUrl}. Check the browser console for more details.`)
    }
    throw error
  }
}

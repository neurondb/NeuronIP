// Authentication utilities for NeuronIP (Cookie-based sessions with API key fallback)

const AUTH_TOKEN_KEY = 'api_token'

// Check if user is authenticated by calling /auth/me endpoint
export async function checkAuth(): Promise<boolean> {
  if (typeof window === 'undefined') return false
  
  try {
    const apiUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082/api/v1'}/auth/me`
    
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout
    
    try {
      const response = await fetch(apiUrl, {
        method: 'GET',
        credentials: 'include', // Send cookies
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
        },
      })
      
      clearTimeout(timeoutId)
      
      if (!response.ok) {
        return false
      }
      
      return true
    } catch (fetchError: any) {
      clearTimeout(timeoutId)
      if (fetchError.name === 'AbortError') {
        console.error('checkAuth: Request timeout - API server may not be responding')
      }
      return false
    }
  } catch (error) {
    console.error('checkAuth: Error checking authentication', error)
    return false
  }
}

// Login with username and password
export async function login(username: string, password: string, database: string = 'neuronip'): Promise<void> {
  const apiUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082/api/v1'}/auth/login`
  
  const response = await fetch(apiUrl, {
    method: 'POST',
    credentials: 'include', // Important for cookies
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password, database }),
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: { message: 'Login failed' } }))
    throw new Error(errorData.error?.message || 'Login failed')
  }

  const data = await response.json()
  
  // Store token for backward compatibility (API key mode)
  if (data.token) {
    localStorage.setItem(AUTH_TOKEN_KEY, data.token)
  }
}

// Register with username and password
export async function register(username: string, password: string, database: string = 'neuronip'): Promise<void> {
  const apiUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082/api/v1'}/auth/register`
  
  const response = await fetch(apiUrl, {
    method: 'POST',
    credentials: 'include', // Important for cookies
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password, database }),
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: { message: 'Registration failed' } }))
    throw new Error(errorData.error?.message || 'Registration failed')
  }

  const data = await response.json()
  
  // Store token for backward compatibility (API key mode)
  if (data.token) {
    localStorage.setItem(AUTH_TOKEN_KEY, data.token)
  }
}

// Logout
export async function logout(): Promise<void> {
  const apiUrl = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082/api/v1'}/auth/logout`
  
  try {
    await fetch(apiUrl, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    })
  } catch (error) {
    console.error('Logout error:', error)
  }
  
  // Clear local storage
  localStorage.removeItem(AUTH_TOKEN_KEY)
  localStorage.removeItem('selected_database')
}

// Legacy API key functions for backwards compatibility
export function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(AUTH_TOKEN_KEY)
}

export function setAuthToken(token: string): void {
  if (typeof window === 'undefined') return
  localStorage.setItem(AUTH_TOKEN_KEY, token)
}

export function removeAuthToken(): void {
  if (typeof window === 'undefined') return
  localStorage.removeItem(AUTH_TOKEN_KEY)
}

export function getAuthHeaders(): Record<string, string> {
  // For cookie-based auth, no headers needed
  // Keep for API key fallback
  const token = getAuthToken()
  if (!token) return {}
  
  return {
    'Authorization': `Bearer ${token}`,
  }
}

// Legacy API key functions for backwards compatibility (deprecated)
export function getAPIKey(): string | null {
  return getAuthToken()
}

export function setAPIKey(key: string): void {
  setAuthToken(key)
}

export function removeAPIKey(): void {
  removeAuthToken()
}

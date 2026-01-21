import axios, { AxiosInstance, AxiosError } from 'axios'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082/api/v1'

export interface ApiError {
  error: string
  message: string
  code?: string
  details?: any
  requestId?: string
}

const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor for auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = typeof window !== 'undefined' ? localStorage.getItem('api_token') : null
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response) {
      const responseData = error.response.data as any
      
      // Parse backend error format: { error: { code: "...", message: "...", details: {...} } }
      let errorMessage = error.message
      let errorCode: string | undefined
      let errorDetails: any = undefined
      
      if (responseData?.error) {
        // Backend structured error format
        const backendError = responseData.error
        errorMessage = backendError.message || errorMessage
        errorCode = backendError.code
        errorDetails = backendError.details
        
        // Extract request ID from details if available
        if (errorDetails?.request_id) {
          // Request ID is already in details, keep it
        }
      } else if (responseData?.message) {
        // Fallback to message field
        errorMessage = responseData.message
        errorCode = responseData.code
      }
      
      const apiError: ApiError = {
        error: error.response.statusText,
        message: errorMessage,
        code: errorCode,
      }
      
      // Add details if available
      if (errorDetails) {
        ;(apiError as any).details = errorDetails
      }
      
      // Add request ID from response header if available
      const requestId = error.response.headers['x-request-id']
      if (requestId) {
        ;(apiError as any).requestId = requestId
      }
      
      return Promise.reject(apiError)
    }
    return Promise.reject(error)
  }
)

export default apiClient

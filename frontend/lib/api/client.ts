import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios'
import { API_BASE_URL } from '../utils/constants'

export const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000,
})

// Request interceptor
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Add auth token if available
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('neuronip-auth-token')
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// Response interceptor
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response) {
      // Handle different error status codes
      switch (error.response.status) {
        case 401:
          // Unauthorized - clear auth and redirect
          if (typeof window !== 'undefined') {
            localStorage.removeItem('neuronip-auth-token')
            window.location.href = '/login'
          }
          break
        case 403:
          console.error('Forbidden: You do not have permission to access this resource')
          break
        case 404:
          console.error('Not Found: The requested resource was not found')
          break
        case 500:
          console.error('Server Error: An internal server error occurred')
          break
        default:
          console.error('API Error:', error.message)
      }
    } else if (error.request) {
      console.error('Network Error: No response received from server')
    } else {
      console.error('Error:', error.message)
    }
    return Promise.reject(error)
  }
)

export default apiClient

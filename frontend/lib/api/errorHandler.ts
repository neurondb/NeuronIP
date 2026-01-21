import { ApiError } from './client'

export interface ErrorRecovery {
  action: string
  description: string
  link?: string
}

export interface FormattedError {
  title: string
  message: string
  code?: string
  recovery?: ErrorRecovery[]
  requestId?: string
  details?: any
}

/**
 * Formats API errors into user-friendly messages with recovery suggestions
 */
export function formatError(error: ApiError | Error | unknown): FormattedError {
  // Handle ApiError from axios interceptor
  if (error && typeof error === 'object' && 'message' in error) {
    const apiError = error as ApiError
    
    // Get error code
    const code = apiError.code || (apiError as any).error
    
    // Get base message
    let message = apiError.message || 'An unexpected error occurred'
    let title = 'Error'
    let recovery: ErrorRecovery[] = []
    
    // Format based on error code
    switch (code) {
      case 'UNAUTHORIZED':
      case '401':
        title = 'Authentication Required'
        message = 'Your session has expired or you are not authenticated.'
        recovery = [
          {
            action: 'Sign in again',
            description: 'Please sign in to continue',
            link: '/login',
          },
        ]
        break
        
      case 'FORBIDDEN':
      case '403':
        title = 'Access Denied'
        message = 'You do not have permission to perform this action.'
        recovery = [
          {
            action: 'Contact administrator',
            description: 'Request access from your administrator',
          },
        ]
        break
        
      case 'NOT_FOUND':
      case '404':
        title = 'Resource Not Found'
        message = 'The requested resource could not be found.'
        recovery = [
          {
            action: 'Go back',
            description: 'Return to the previous page',
          },
        ]
        break
        
      case 'VALIDATION_FAILED':
      case 'BAD_REQUEST':
      case '400':
        title = 'Invalid Input'
        message = apiError.message || 'Please check your input and try again.'
        if ((apiError as any).details) {
          const details = (apiError as any).details
          if (typeof details === 'object') {
            const fieldErrors = Object.entries(details)
              .filter(([key]) => key !== 'request_id')
              .map(([field, msg]) => `${field}: ${msg}`)
              .join(', ')
            if (fieldErrors) {
              message = `Validation errors: ${fieldErrors}`
            }
          }
        }
        recovery = [
          {
            action: 'Review form',
            description: 'Check the highlighted fields and correct any errors',
          },
        ]
        break
        
      case 'TOO_MANY_REQUESTS':
      case '429':
        title = 'Rate Limit Exceeded'
        message = 'You have made too many requests. Please wait a moment before trying again.'
        recovery = [
          {
            action: 'Wait and retry',
            description: 'Please wait a few seconds and try again',
          },
        ]
        break
        
      case 'TIMEOUT':
      case '504':
        title = 'Request Timeout'
        message = 'The request took too long to complete. This may be due to a large query or system load.'
        recovery = [
          {
            action: 'Try again',
            description: 'Retry the operation',
          },
          {
            action: 'Simplify query',
            description: 'Try a simpler or more specific query',
          },
        ]
        break
        
      case 'SERVICE_UNAVAILABLE':
      case '503':
        title = 'Service Unavailable'
        message = 'The service is temporarily unavailable. Please try again later.'
        recovery = [
          {
            action: 'Retry later',
            description: 'Wait a few minutes and try again',
          },
        ]
        break
        
      case 'INTERNAL_SERVER_ERROR':
      case '500':
        title = 'Server Error'
        message = 'An internal server error occurred. Our team has been notified.'
        recovery = [
          {
            action: 'Try again',
            description: 'Retry the operation',
          },
          {
            action: 'Contact support',
            description: 'If the problem persists, contact support',
            link: '/dashboard/support',
          },
        ]
        break
        
      default:
        // Use the error message as-is for unknown errors
        title = 'Error'
        if (apiError.message) {
          message = apiError.message
        }
        recovery = [
          {
            action: 'Try again',
            description: 'Retry the operation',
          },
        ]
    }
    
    return {
      title,
      message,
      code,
      recovery,
      requestId: apiError.requestId || (apiError as any).request_id,
      details: (apiError as any).details,
    }
  }
  
  // Handle generic Error objects
  if (error instanceof Error) {
    return {
      title: 'Error',
      message: error.message || 'An unexpected error occurred',
      recovery: [
        {
          action: 'Try again',
          description: 'Retry the operation',
        },
      ],
    }
  }
  
  // Fallback for unknown error types
  return {
    title: 'Error',
    message: 'An unexpected error occurred',
    recovery: [
      {
        action: 'Try again',
        description: 'Retry the operation',
      },
    ],
  }
}

/**
 * Gets a user-friendly error message from an error
 */
export function getErrorMessage(error: ApiError | Error | unknown): string {
  return formatError(error).message
}

/**
 * Gets recovery suggestions for an error
 */
export function getErrorRecovery(error: ApiError | Error | unknown): ErrorRecovery[] {
  return formatError(error).recovery || []
}

/**
 * Checks if an error is a specific type
 */
export function isErrorCode(error: ApiError | Error | unknown, code: string): boolean {
  if (error && typeof error === 'object' && 'code' in error) {
    return (error as ApiError).code === code
  }
  return false
}

/**
 * Checks if an error is a validation error
 */
export function isValidationError(error: ApiError | Error | unknown): boolean {
  return isErrorCode(error, 'VALIDATION_FAILED') || isErrorCode(error, 'BAD_REQUEST')
}

/**
 * Checks if an error is an authentication error
 */
export function isAuthError(error: ApiError | Error | unknown): boolean {
  return isErrorCode(error, 'UNAUTHORIZED') || isErrorCode(error, 'FORBIDDEN')
}

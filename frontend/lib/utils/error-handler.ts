export interface ErrorInfo {
  message: string
  code?: string
  statusCode?: number
  timestamp: number
  context?: Record<string, any>
}

export class AppError extends Error {
  code?: string
  statusCode?: number
  context?: Record<string, any>
  retryable?: boolean

  constructor(
    message: string,
    options?: {
      code?: string
      statusCode?: number
      context?: Record<string, any>
      retryable?: boolean
    }
  ) {
    super(message)
    this.name = 'AppError'
    this.code = options?.code
    this.statusCode = options?.statusCode
    this.context = options?.context
    this.retryable = options?.retryable ?? false
  }
}

export function createErrorHandler() {
  return {
    handle: (error: unknown): ErrorInfo => {
      const errorInfo: ErrorInfo = {
        message: 'An unexpected error occurred',
        timestamp: Date.now(),
      }

      if (error instanceof AppError) {
        errorInfo.message = error.message
        errorInfo.code = error.code
        errorInfo.statusCode = error.statusCode
        errorInfo.context = error.context
      } else if (error instanceof Error) {
        errorInfo.message = error.message
      } else if (typeof error === 'string') {
        errorInfo.message = error
      }

      // Log to console in development
      if (process.env.NODE_ENV === 'development') {
        console.error('Error:', errorInfo)
      }

      // Send to error tracking service
      if (typeof window !== 'undefined' && (window as any).Sentry) {
        ;(window as any).Sentry.captureException(error, {
          extra: errorInfo.context,
        })
      }

      return errorInfo
    },
  }
}

export function withErrorHandling<T extends (...args: any[]) => Promise<any>>(
  fn: T,
  errorHandler = createErrorHandler()
): T {
  return (async (...args: any[]) => {
    try {
      return await fn(...args)
    } catch (error) {
      const errorInfo = errorHandler.handle(error)
      throw new AppError(errorInfo.message, {
        code: errorInfo.code,
        statusCode: errorInfo.statusCode,
        context: errorInfo.context,
      })
    }
  }) as T
}

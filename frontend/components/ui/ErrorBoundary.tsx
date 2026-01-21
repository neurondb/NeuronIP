'use client'

import { Component, ReactNode } from 'react'
import Link from 'next/link'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './Card'
import Button from './Button'
import { formatError, getErrorRecovery } from '@/lib/api/errorHandler'
import { ApiError } from '@/lib/api/client'
import { ExclamationTriangleIcon, ArrowPathIcon, HomeIcon } from '@heroicons/react/24/outline'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onReset?: () => void
}

interface State {
  hasError: boolean
  error: Error | null
  apiError: ApiError | null
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null, apiError: null }
  }

  static getDerivedStateFromError(error: Error | ApiError): State {
    // Check if it's an API error
    if (error && typeof error === 'object' && 'code' in error) {
      return { hasError: true, error: null, apiError: error as ApiError }
    }
    return { hasError: true, error: error as Error, apiError: null }
  }

  componentDidCatch(error: Error | ApiError, errorInfo: any) {
    console.error('Error caught by boundary:', error, errorInfo)
    
    // Log to error tracking service if available
    if (typeof window !== 'undefined' && (window as any).Sentry) {
      ;(window as any).Sentry.captureException(error, {
        contexts: {
          react: {
            componentStack: errorInfo.componentStack,
          },
        },
      })
    }
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null, apiError: null })
    if (this.props.onReset) {
      this.props.onReset()
    }
  }

  handleReload = () => {
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      // Format error using error handler
      const formattedError = formatError(this.state.apiError || this.state.error)
      const recovery = getErrorRecovery(this.state.apiError || this.state.error)

      return (
        <div className="flex items-center justify-center min-h-screen p-4 bg-background">
          <Card className="max-w-2xl w-full">
            <CardHeader>
              <div className="flex items-center gap-3">
                <ExclamationTriangleIcon className="h-8 w-8 text-destructive" />
                <div>
                  <CardTitle>{formattedError.title}</CardTitle>
                  <CardDescription className="mt-1">
                    {formattedError.message}
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Error details */}
              {(this.state.error || this.state.apiError) && (
                <div className="space-y-2">
                  {formattedError.requestId && (
                    <div className="text-xs text-muted-foreground">
                      Request ID: <code className="bg-muted px-1 py-0.5 rounded">{formattedError.requestId}</code>
                    </div>
                  )}
                  {formattedError.code && (
                    <div className="text-xs text-muted-foreground">
                      Error Code: <code className="bg-muted px-1 py-0.5 rounded">{formattedError.code}</code>
                    </div>
                  )}
                  {process.env.NODE_ENV === 'development' && this.state.error && (
                    <details className="mt-2">
                      <summary className="text-sm text-muted-foreground cursor-pointer">
                        Technical Details
                      </summary>
                      <pre className="mt-2 p-3 rounded-lg bg-muted text-xs overflow-auto max-h-40">
                        {this.state.error.stack || this.state.error.message}
                      </pre>
                    </details>
                  )}
                </div>
              )}

              {/* Recovery actions */}
              {recovery && recovery.length > 0 && (
                <div className="space-y-2">
                  <h4 className="text-sm font-semibold">What you can do:</h4>
                  <div className="space-y-2">
                    {recovery.map((action, index) => (
                      <div key={index} className="flex items-start gap-2">
                        <div className="mt-0.5 h-2 w-2 rounded-full bg-primary" />
                        <div className="flex-1">
                          <p className="text-sm font-medium">{action.action}</p>
                          {action.description && (
                            <p className="text-xs text-muted-foreground">{action.description}</p>
                          )}
                          {action.link && (
                            <Link
                              href={action.link}
                              className="text-xs text-primary hover:underline mt-1 inline-block"
                            >
                              {action.link}
                            </Link>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Action buttons */}
              <div className="flex gap-2 pt-2">
                <Button onClick={this.handleReset} variant="outline">
                  <ArrowPathIcon className="h-4 w-4 mr-2" />
                  Try Again
                </Button>
                <Button onClick={this.handleReload} variant="outline">
                  Reload Page
                </Button>
                <Link href="/dashboard">
                  <Button variant="ghost">
                    <HomeIcon className="h-4 w-4 mr-2" />
                    Go Home
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        </div>
      )
    }

    return this.props.children
  }
}

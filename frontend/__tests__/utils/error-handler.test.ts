import { describe, it, expect, vi } from 'vitest'

import { createErrorHandler, AppError, withErrorHandling } from '@/lib/utils/error-handler'

describe('error-handler', () => {
  it('creates error handler', () => {
    const handler = createErrorHandler()
    expect(handler).toBeDefined()
    expect(handler.handle).toBeDefined()
  })

  it('handles AppError instances', () => {
    const handler = createErrorHandler()
    const error = new AppError('Test error', { code: 'TEST_ERROR', statusCode: 400 })
    const errorInfo = handler.handle(error)
    
    expect(errorInfo.message).toBe('Test error')
    expect(errorInfo.code).toBe('TEST_ERROR')
    expect(errorInfo.statusCode).toBe(400)
  })

  it('handles standard Error instances', () => {
    const handler = createErrorHandler()
    const error = new Error('Standard error')
    const errorInfo = handler.handle(error)
    
    expect(errorInfo.message).toBe('Standard error')
  })

  it('handles string errors', () => {
    const handler = createErrorHandler()
    const errorInfo = handler.handle('String error')
    
    expect(errorInfo.message).toBe('String error')
  })

  it('handles unknown error types', () => {
    const handler = createErrorHandler()
    const errorInfo = handler.handle({ message: 'Object error' })
    
    expect(errorInfo.message).toBe('An unexpected error occurred')
  })

  it('wraps function with error handling', async () => {
    const testFn = async (x: number) => {
      if (x < 0) throw new Error('Negative number')
      return x * 2
    }

    const wrappedFn = withErrorHandling(testFn)
    
    const result = await wrappedFn(5)
    expect(result).toBe(10)

    await expect(wrappedFn(-1)).rejects.toThrow()
  })
})

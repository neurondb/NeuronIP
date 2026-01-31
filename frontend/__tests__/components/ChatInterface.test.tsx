import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'

import ChatInterface from '@/components/semantic/ChatInterface'

import { render, screen, waitFor } from '../utils/test-utils'

// Mock scrollIntoView
if (typeof Element !== 'undefined' && !Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = vi.fn()
}

// Mock the API hook
const mockMutate = vi.fn()
vi.mock('@/lib/api/queries', () => ({
  useSemanticRAG: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
}))

describe('ChatInterface', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders chat interface', () => {
    render(<ChatInterface />)
    
    const input = screen.getByRole('textbox')
    expect(input).toBeInTheDocument()
  })

  it('displays input field', () => {
    render(<ChatInterface />)
    
    const input = screen.getByRole('textbox')
    expect(input).toBeInTheDocument()
  })

  it('allows typing in input', async () => {
    const user = userEvent.setup()
    render(<ChatInterface />)
    
    const input = screen.getByRole('textbox')
    await user.type(input, 'Test message')
    
    expect(input).toHaveValue('Test message')
  })

  it('sends message on submit', async () => {
    const user = userEvent.setup()
    mockMutate.mockClear()
    mockMutate.mockImplementation((data, callbacks) => {
      // Simulate success immediately
      if (callbacks?.onSuccess) {
        // Use setTimeout to simulate async behavior
        setTimeout(() => {
          callbacks.onSuccess({ response: 'Test response' })
        }, 50)
      }
    })

    render(<ChatInterface />)
    
    const input = screen.getByRole('textbox')
    // Form submission can be triggered by Enter key or button click
    // Try to find submit button, or submit form directly
    const form = input.closest('form')
    
    await user.type(input, 'Test query')
    
    if (form) {
      // Submit the form
      await user.type(input, '{Enter}')
    } else {
      // Try to find and click send button
      const sendButton = screen.queryByRole('button', { name: /send|submit/i }) ||
                        screen.queryByLabelText(/send|submit/i)
      if (sendButton) {
        await user.click(sendButton)
      } else {
        // Fallback: submit via Enter key
        await user.type(input, '{Enter}')
      }
    }
    
    // Wait for mutate to be called
    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          query: 'Test query',
        }),
        expect.any(Object)
      )
    }, { timeout: 2000 })
  })

  it('clears input after sending', async () => {
    const user = userEvent.setup()
    mockMutate.mockClear()
    mockMutate.mockImplementation((data, callbacks) => {
      // Simulate success
      if (callbacks?.onSuccess) {
        setTimeout(() => {
          callbacks.onSuccess({ response: 'Test response' })
        }, 50)
      }
    })

    render(<ChatInterface />)
    
    const input = screen.getByRole('textbox')
    const form = input.closest('form')
    
    await user.type(input, 'Test message')
    expect(input).toHaveValue('Test message')
    
    // Submit form via Enter key (form submission)
    if (form) {
      await user.type(input, '{Enter}')
    } else {
      // Try button click
      const sendButton = screen.queryByRole('button', { name: /send|submit/i })
      if (sendButton) {
        await user.click(sendButton)
      } else {
        await user.type(input, '{Enter}')
      }
    }
    
    // Input should be cleared immediately after submitting (component clears it before API call)
    await waitFor(() => {
      expect(input).toHaveValue('')
    }, { timeout: 1000 })
  })

  it('displays messages', async () => {
    const user = userEvent.setup()
    mockMutate.mockClear()
    mockMutate.mockImplementation((data, callbacks) => {
      // Simulate success callback
      setTimeout(() => {
        if (callbacks?.onSuccess) {
          callbacks.onSuccess({ response: 'Test response' })
        }
      }, 50)
    })

    render(<ChatInterface />)
    
    const input = screen.getByRole('textbox')
    const form = input.closest('form')
    
    await user.type(input, 'Test query')
    
    // Submit form
    if (form) {
      await user.type(input, '{Enter}')
    } else {
      const sendButton = screen.queryByRole('button', { name: /send|submit/i })
      if (sendButton) {
        await user.click(sendButton)
      } else {
        await user.type(input, '{Enter}')
      }
    }
    
    // Wait for user message to appear (it's added immediately before API call)
    await waitFor(() => {
      expect(screen.getByText('Test query')).toBeInTheDocument()
    }, { timeout: 2000 })
    
    // Also wait for assistant response
    await waitFor(() => {
      expect(screen.getByText(/Test response|Response received/i)).toBeInTheDocument()
    }, { timeout: 2000 })
  })
})

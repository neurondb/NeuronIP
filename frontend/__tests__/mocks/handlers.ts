import { http, HttpResponse } from 'msw'

// Mock API handlers for testing
export const handlers = [
  // Example: Mock API endpoint
  http.get('/api/health', () => {
    return HttpResponse.json({ status: 'ok' })
  }),

  // Add more mock handlers as needed
  http.get('/api/users', () => {
    return HttpResponse.json([
      { id: 1, name: 'John Doe', email: 'john@example.com' },
      { id: 2, name: 'Jane Smith', email: 'jane@example.com' },
    ])
  }),
]

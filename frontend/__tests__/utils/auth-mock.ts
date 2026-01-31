import { vi } from 'vitest'

// Mock authentication utilities for testing

export const mockUser = {
  id: '1',
  email: 'test@example.com',
  name: 'Test User',
  role: 'admin',
}

export const mockToken = 'mock-jwt-token'

export const mockAuthContext = {
  user: mockUser,
  token: mockToken,
  isAuthenticated: true,
  login: vi.fn(),
  logout: vi.fn(),
  refreshToken: vi.fn(),
}

// Mock localStorage for auth
export const setupAuthMock = () => {
  const localStorageMock = {
    getItem: vi.fn((key: string) => {
      if (key === 'token') return mockToken
      if (key === 'user') return JSON.stringify(mockUser)
      return null
    }),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn(),
  }
  
  Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
    writable: true,
  })
  
  return localStorageMock
}

// Mock session storage
export const setupSessionMock = () => {
  const sessionStorageMock = {
    getItem: vi.fn(),
    setItem: vi.fn(),
    removeItem: vi.fn(),
    clear: vi.fn(),
  }
  
  Object.defineProperty(window, 'sessionStorage', {
    value: sessionStorageMock,
    writable: true,
  })
  
  return sessionStorageMock
}

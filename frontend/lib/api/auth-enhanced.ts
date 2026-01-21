import apiClient from './client'

export interface TOTPSecret {
  secret: string
  qr_code_url: string
}

export interface UserSession {
  id: string
  user_id: string
  ip_address?: string
  user_agent?: string
  expires_at: string
  created_at: string
}

export async function initiateOIDC(provider: string): Promise<{ auth_url: string; state: string }> {
  const response = await apiClient.post(`/auth/oidc/${provider}/initiate`)
  return response.data
}

export async function generateTOTPSecret(userId: string, email: string): Promise<TOTPSecret> {
  const response = await apiClient.post('/auth/2fa/generate', { user_id: userId, email })
  return response.data
}

export async function getUserSessions(): Promise<UserSession[]> {
  const response = await apiClient.get('/auth/sessions')
  return response.data
}

export async function revokeSession(sessionId: string): Promise<void> {
  await apiClient.delete(`/auth/sessions/${sessionId}`)
}

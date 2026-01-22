'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { useAppStore } from '@/lib/store/useAppStore'
import { KeyIcon, EyeIcon, EyeSlashIcon, CheckCircleIcon } from '@heroicons/react/24/outline'
import apiClient from '@/lib/api/client'
import { showToast } from '@/components/ui/Toast'

export default function SecuritySettings() {
  const { notificationsEnabled, setNotificationsEnabled } = useAppStore()
  const [apiKey, setApiKey] = useState('')
  const [showApiKey, setShowApiKey] = useState(false)
  const [testing, setTesting] = useState(false)
  const [isValid, setIsValid] = useState<boolean | null>(null)
  const router = useRouter()

  const [hasToken, setHasToken] = useState(false)

  useEffect(() => {
    // Load current API key (masked) - client-side only
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('api_token')
      if (token) {
        // Show masked version
        const masked = token.length > 8 
          ? `${token.substring(0, 4)}${'•'.repeat(token.length - 8)}${token.substring(token.length - 4)}`
          : '•'.repeat(8)
        setApiKey(masked)
        setIsValid(true)
        setHasToken(true)
      }
    }
  }, [])

  const testApiKey = async (key: string): Promise<boolean> => {
    try {
      const response = await apiClient.get('/warehouse/schemas', {
        headers: {
          Authorization: `Bearer ${key}`,
        },
      })
      return response.status === 200
    } catch (err) {
      return false
    }
  }

  const handleUpdateApiKey = async () => {
    if (!apiKey.trim()) {
      showToast('Please enter an API key', 'error')
      return
    }

    setTesting(true)
    const isValidKey = await testApiKey(apiKey.trim())

    if (!isValidKey) {
      showToast('Invalid API key. Please check your key and try again.', 'error')
      setTesting(false)
      setIsValid(false)
      return
    }

    // Save the token - client-side only
    if (typeof window !== 'undefined') {
      localStorage.setItem('api_token', apiKey.trim())
      setHasToken(true)
    }
    
    // Mask the key for display
    const masked = apiKey.length > 8 
      ? `${apiKey.substring(0, 4)}${'•'.repeat(apiKey.length - 8)}${apiKey.substring(apiKey.length - 4)}`
      : '•'.repeat(8)
    setApiKey(masked)
    setShowApiKey(false)
    setIsValid(true)
    setTesting(false)
    
    showToast('API key updated successfully', 'success')
    router.refresh()
  }

  const handleClearApiKey = () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('api_token')
      setHasToken(false)
    }
    setApiKey('')
    setIsValid(null)
    showToast('API key cleared. You will be redirected to login.', 'info')
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Security & Authentication</CardTitle>
        <CardDescription>Manage security and authentication settings</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* API Key Management */}
        <div className="space-y-4 p-4 bg-muted/50 rounded-lg border border-border">
          <div className="flex items-center gap-2">
            <KeyIcon className="h-5 w-5 text-foreground" />
            <h3 className="text-sm font-semibold">API Key</h3>
            {isValid === true && (
              <CheckCircleIcon className="h-4 w-4 text-green-500" />
            )}
          </div>
          
          <div className="space-y-3">
            <div className="flex gap-2">
              <div className="flex-1">
                <Input
                  label="Current API Key"
                  type={showApiKey ? 'text' : 'password'}
                  value={apiKey}
                  onChange={(e) => {
                    setApiKey(e.target.value)
                    setIsValid(null)
                  }}
                  placeholder="Enter your API key"
                  disabled={testing}
                  helperText={isValid === true ? 'API key is valid and active' : isValid === false ? 'API key is invalid' : 'Enter a new API key to update'}
                />
              </div>
              <div className="flex items-end">
                <Button
                  variant="outline"
                  onClick={() => setShowApiKey(!showApiKey)}
                  disabled={!apiKey || testing}
                  className="h-10"
                >
                  {showApiKey ? (
                    <EyeSlashIcon className="h-4 w-4" />
                  ) : (
                    <EyeIcon className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>

            {testing && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <div className="h-4 w-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                Verifying API key...
              </div>
            )}

            <div className="flex gap-2">
              <Button
                onClick={handleUpdateApiKey}
                disabled={!apiKey.trim() || testing}
                className="flex-1"
              >
                {apiKey && isValid === true ? 'Update API Key' : 'Set API Key'}
              </Button>
              {hasToken && (
                <Button
                  variant="outline"
                  onClick={handleClearApiKey}
                  disabled={testing}
                >
                  Clear
                </Button>
              )}
            </div>
          </div>
        </div>

        <div>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={notificationsEnabled}
              onChange={(e) => setNotificationsEnabled(e.target.checked)}
              className="rounded border-border"
            />
            <span className="text-sm">Enable Two-Factor Authentication</span>
          </label>
        </div>

        <div>
          <label className="text-sm font-medium mb-3 block">Password Policy</label>
          <select className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
            <option>Standard (8+ characters)</option>
            <option>Strong (12+ characters, mixed case)</option>
            <option>Very Strong (16+ characters, special chars)</option>
          </select>
        </div>

        <div>
          <label className="text-sm font-medium mb-3 block">Session Timeout</label>
          <select className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
            <option>15 minutes</option>
            <option>30 minutes</option>
            <option>1 hour</option>
            <option>8 hours</option>
          </select>
        </div>

        <Button variant="outline">Change Password</Button>
      </CardContent>
    </Card>
  )
}

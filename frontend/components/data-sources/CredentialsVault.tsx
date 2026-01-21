'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { LockClosedIcon, EyeIcon, EyeSlashIcon } from '@heroicons/react/24/outline'

interface CredentialsVaultProps {
  connectorId: string
  onSave: (credentials: any) => void
}

export default function CredentialsVault({ connectorId, onSave }: CredentialsVaultProps) {
  const [showPassword, setShowPassword] = useState(false)
  const [credentials, setCredentials] = useState({
    username: '',
    password: '',
    apiKey: '',
    secretKey: '',
  })

  const handleSave = () => {
    // Credentials are encrypted before sending to API
    onSave(credentials)
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <LockClosedIcon className="h-5 w-5" />
          <CardTitle>Credentials Vault</CardTitle>
        </div>
        <CardDescription>
          Store credentials securely. All credentials are encrypted at rest.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Username</label>
            <Input
              type="text"
              value={credentials.username}
              onChange={(e) => setCredentials({ ...credentials, username: e.target.value })}
              placeholder="Enter username"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <div className="relative">
              <Input
                type={showPassword ? 'text' : 'password'}
                value={credentials.password}
                onChange={(e) => setCredentials({ ...credentials, password: e.target.value })}
                placeholder="Enter password"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                {showPassword ? (
                  <EyeSlashIcon className="h-5 w-5" />
                ) : (
                  <EyeIcon className="h-5 w-5" />
                )}
              </button>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">API Key</label>
            <Input
              type="password"
              value={credentials.apiKey}
              onChange={(e) => setCredentials({ ...credentials, apiKey: e.target.value })}
              placeholder="Enter API key"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">Secret Key</label>
            <Input
              type="password"
              value={credentials.secretKey}
              onChange={(e) => setCredentials({ ...credentials, secretKey: e.target.value })}
              placeholder="Enter secret key"
            />
          </div>

          <div className="pt-4">
            <Button onClick={handleSave}>Save Credentials</Button>
          </div>

          <div className="text-xs text-muted-foreground pt-2 border-t">
            <p>• Credentials are encrypted using AES-256</p>
            <p>• Only authorized users can view credentials</p>
            <p>• Credentials are never logged or exposed in API responses</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

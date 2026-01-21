'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useCreateIntegration, useUpdateIntegration, useIntegration } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface IntegrationConfigDialogProps {
  integrationId?: string
  integrationType?: string
  onSuccess?: () => void
  onCancel?: () => void
}

export default function IntegrationConfigDialog({
  integrationId,
  integrationType,
  onSuccess,
  onCancel,
}: IntegrationConfigDialogProps) {
  const [name, setName] = useState('')
  const [type, setType] = useState(integrationType || '')
  const [enabled, setEnabled] = useState(true)
  const [config, setConfig] = useState<Record<string, any>>({})

  const { data: existingIntegration } = useIntegration(integrationId || '', !!integrationId)
  const createIntegration = useCreateIntegration()
  const updateIntegration = useUpdateIntegration()

  useEffect(() => {
    if (existingIntegration) {
      setName(existingIntegration.name || '')
      setType(existingIntegration.integration_type || '')
      setEnabled(existingIntegration.enabled !== false)
      setConfig(existingIntegration.config || {})
    }
  }, [existingIntegration])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!name || !type) {
      showToast('Name and type are required', 'warning')
      return
    }

    const integrationData = {
      name,
      integration_type: type,
      enabled,
      config,
    }

    try {
      if (integrationId) {
        await updateIntegration.mutateAsync({ id: integrationId, data: integrationData })
        showToast('Integration updated successfully', 'success')
      } else {
        await createIntegration.mutateAsync(integrationData)
        showToast('Integration created successfully', 'success')
      }
      onSuccess?.()
    } catch (error: any) {
      showToast(error?.response?.data?.message || 'Failed to save integration', 'error')
    }
  }

  const renderConfigFields = () => {
    switch (type) {
      case 'slack':
        return (
          <>
            <div>
              <label className="text-sm font-medium mb-2 block">Webhook URL</label>
              <input
                type="url"
                value={config.webhook_url || ''}
                onChange={(e) => setConfig({ ...config, webhook_url: e.target.value })}
                placeholder="https://hooks.slack.com/services/..."
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Bot Token (optional)</label>
              <input
                type="password"
                value={config.bot_token || ''}
                onChange={(e) => setConfig({ ...config, bot_token: e.target.value })}
                placeholder="xoxb-..."
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Channel (optional)</label>
              <input
                type="text"
                value={config.channel || ''}
                onChange={(e) => setConfig({ ...config, channel: e.target.value })}
                placeholder="#general"
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
          </>
        )

      case 'teams':
        return (
          <div>
            <label className="text-sm font-medium mb-2 block">Webhook URL</label>
            <input
              type="url"
              value={config.webhook_url || ''}
              onChange={(e) => setConfig({ ...config, webhook_url: e.target.value })}
              placeholder="https://outlook.office.com/webhook/..."
              className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
        )

      case 'email':
        return (
          <>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium mb-2 block">SMTP Host</label>
                <input
                  type="text"
                  value={config.smtp_host || ''}
                  onChange={(e) => setConfig({ ...config, smtp_host: e.target.value })}
                  placeholder="smtp.gmail.com"
                  className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </div>
              <div>
                <label className="text-sm font-medium mb-2 block">SMTP Port</label>
                <input
                  type="number"
                  value={config.smtp_port || 587}
                  onChange={(e) => setConfig({ ...config, smtp_port: parseInt(e.target.value) || 587 })}
                  className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </div>
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">SMTP Username</label>
              <input
                type="text"
                value={config.smtp_username || ''}
                onChange={(e) => setConfig({ ...config, smtp_username: e.target.value })}
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">SMTP Password</label>
              <input
                type="password"
                value={config.smtp_password || ''}
                onChange={(e) => setConfig({ ...config, smtp_password: e.target.value })}
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">From Email</label>
              <input
                type="email"
                value={config.from_email || ''}
                onChange={(e) => setConfig({ ...config, from_email: e.target.value })}
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="use_tls"
                checked={config.use_tls || false}
                onChange={(e) => setConfig({ ...config, use_tls: e.target.checked })}
                className="rounded border-border"
              />
              <label htmlFor="use_tls" className="text-sm font-medium">
                Use TLS
              </label>
            </div>
          </>
        )

      default:
        return (
          <div>
            <label className="text-sm font-medium mb-2 block">Configuration (JSON)</label>
            <textarea
              value={JSON.stringify(config, null, 2)}
              onChange={(e) => {
                try {
                  setConfig(JSON.parse(e.target.value))
                } catch {
                  // Invalid JSON, ignore
                }
              }}
              rows={6}
              className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring resize-none font-mono text-sm"
            />
          </div>
        )
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{integrationId ? 'Edit Integration' : 'Create Integration'}</CardTitle>
        <CardDescription>Configure integration settings</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-2 block">Name *</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="My Slack Integration"
              className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-2 block">Type *</label>
            <select
              value={type}
              onChange={(e) => setType(e.target.value)}
              disabled={!!integrationId}
              className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              required
            >
              <option value="">Select type...</option>
              <option value="slack">Slack</option>
              <option value="teams">Microsoft Teams</option>
              <option value="email">Email</option>
              <option value="webhook">Webhook</option>
              <option value="helpdesk">Helpdesk</option>
            </select>
          </div>

          {type && renderConfigFields()}

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="enabled"
              checked={enabled}
              onChange={(e) => setEnabled(e.target.checked)}
              className="rounded border-border"
            />
            <label htmlFor="enabled" className="text-sm font-medium">
              Enabled
            </label>
          </div>

          <div className="flex gap-2 justify-end">
            {onCancel && (
              <Button type="button" variant="outline" onClick={onCancel}>
                Cancel
              </Button>
            )}
            <Button type="submit" disabled={createIntegration.isPending || updateIntegration.isPending}>
              {integrationId ? 'Update Integration' : 'Create Integration'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

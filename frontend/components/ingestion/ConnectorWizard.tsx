'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { showToast } from '@/components/ui/Toast'
import { Badge } from '@/components/ui/Badge'

interface ConnectorType {
  type_name: string
  display_name: string
  connector_category: string
  config_schema: any
}

interface ConnectorWizardProps {
  onComplete?: (connectorId: string) => void
  onCancel?: () => void
}

export default function ConnectorWizard({ onComplete, onCancel }: ConnectorWizardProps) {
  const [step, setStep] = useState(1)
  const [connectorType, setConnectorType] = useState<string | null>(null)
  const [config, setConfig] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(false)

  // Common SaaS connectors
  const saasConnectors: ConnectorType[] = [
    {
      type_name: 'zendesk',
      display_name: 'Zendesk',
      connector_category: 'saas',
      config_schema: { subdomain: 'string', email: 'string', api_token: 'string' },
    },
    {
      type_name: 'jira',
      display_name: 'Jira',
      connector_category: 'saas',
      config_schema: { url: 'string', email: 'string', api_token: 'string' },
    },
    {
      type_name: 'salesforce',
      display_name: 'Salesforce',
      connector_category: 'saas',
      config_schema: {
        instance_url: 'string',
        client_id: 'string',
        client_secret: 'string',
        username: 'string',
        password: 'string',
      },
    },
    {
      type_name: 'hubspot',
      display_name: 'HubSpot',
      connector_category: 'saas',
      config_schema: { api_key: 'string', portal_id: 'string' },
    },
  ]

  const handleConnectorSelect = (connector: ConnectorType) => {
    setConnectorType(connector.type_name)
    setStep(2)
  }

  const handleConfigChange = (key: string, value: string) => {
    setConfig({ ...config, [key]: value })
  }

  const handleOAuthConnect = async () => {
    if (!connectorType) return

    setLoading(true)
    try {
      // In production, this would initiate OAuth flow
      showToast('OAuth connection would be initiated here', 'info')
    } catch (error: any) {
      showToast(error.message || 'Failed to initiate OAuth', 'error')
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    if (!connectorType) return

    setLoading(true)
    try {
      const response = await fetch('/api/v1/data-sources/connectors', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          connector_type: connectorType,
          config,
        }),
      })

      if (!response.ok) throw new Error('Failed to create connector')

      const data = await response.json()
      showToast('Connector created successfully', 'success')
      onComplete?.(data.id)
    } catch (error: any) {
      showToast(error.message || 'Failed to create connector', 'error')
    } finally {
      setLoading(false)
    }
  }

  const selectedConnector = saasConnectors.find((c) => c.type_name === connectorType)

  return (
    <Card className="max-w-3xl mx-auto">
      <CardHeader>
        <CardTitle>Setup Data Connector</CardTitle>
        <CardDescription>Connect to your data source</CardDescription>
      </CardHeader>
      <CardContent>
        <AnimatePresence mode="wait">
          {step === 1 && (
            <motion.div
              key="step1"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
            >
              <h3 className="text-lg font-medium mb-4">Select Connector Type</h3>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {saasConnectors.map((connector) => (
                  <motion.div
                    key={connector.type_name}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <Button
                      variant="outline"
                      onClick={() => handleConnectorSelect(connector)}
                      className="w-full h-auto py-4 flex flex-col items-center gap-2"
                    >
                      <span className="font-medium">{connector.display_name}</span>
                      <Badge variant="secondary">{connector.connector_category}</Badge>
                    </Button>
                  </motion.div>
                ))}
              </div>
            </motion.div>
          )}

          {step === 2 && selectedConnector && (
            <motion.div
              key="step2"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              className="space-y-4"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-medium">{selectedConnector.display_name} Configuration</h3>
                  <p className="text-sm text-muted-foreground">
                    Enter your connection details or use OAuth
                  </p>
                </div>
                <Button variant="outline" onClick={() => setStep(1)}>
                  Back
                </Button>
              </div>

              <div className="space-y-3">
                {Object.keys(selectedConnector.config_schema).map((key) => (
                  <div key={key}>
                    <label className="text-sm font-medium mb-1 block">
                      {key.replace(/_/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase())}
                    </label>
                    <Input
                      type={key.includes('token') || key.includes('secret') || key.includes('password') ? 'password' : 'text'}
                      value={config[key] || ''}
                      onChange={(e) => handleConfigChange(key, e.target.value)}
                      placeholder={`Enter ${key}`}
                    />
                  </div>
                ))}
              </div>

              <div className="flex gap-2 pt-4">
                <Button variant="secondary" onClick={handleOAuthConnect} disabled={loading}>
                  Connect with OAuth
                </Button>
                <Button onClick={handleSave} disabled={loading || !config}>
                  Save & Test
                </Button>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        <div className="mt-6 flex justify-between pt-4 border-t">
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <div className="flex gap-2">
            <div className="flex items-center gap-1">
              {[1, 2].map((s) => (
                <div
                  key={s}
                  className={`w-2 h-2 rounded-full ${step >= s ? 'bg-primary' : 'bg-muted'}`}
                />
              ))}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

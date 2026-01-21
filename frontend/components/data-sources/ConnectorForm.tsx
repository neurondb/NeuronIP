'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import Select from '@/components/ui/Select'
import Tooltip from '@/components/ui/Tooltip'
import Warning from '@/components/ui/Warning'
import HelpText from '@/components/ui/HelpText'
import { InformationCircleIcon } from '@heroicons/react/24/outline'

interface ConnectorFormProps {
  onSubmit: (data: any) => void
  onCancel: () => void
  initialData?: any
}

export default function ConnectorForm({ onSubmit, onCancel, initialData }: ConnectorFormProps) {
  const [formData, setFormData] = useState({
    name: initialData?.name || '',
    type: initialData?.type || 'postgresql',
    connectionString: initialData?.connectionString || '',
    enabled: initialData?.enabled ?? true,
  })

  const connectorTypes = [
    { value: 'postgresql', label: 'PostgreSQL' },
    { value: 'mysql', label: 'MySQL' },
    { value: 'mongodb', label: 'MongoDB' },
    { value: 'api', label: 'REST API' },
    { value: 's3', label: 'Amazon S3' },
    { value: 'salesforce', label: 'Salesforce' },
    { value: 'zendesk', label: 'Zendesk' },
    { value: 'slack', label: 'Slack' },
  ]

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(formData)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{initialData ? 'Edit Connector' : 'Add Connector'}</CardTitle>
        <CardDescription>Configure a new data source connector</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Warning
            message="Credentials are encrypted and stored securely. Only authorized users can access connection details."
            severity="low"
            title="Security Notice"
          />
          
          <div>
            <div className="flex items-center gap-2 mb-1">
              <label className="block text-sm font-medium">Name</label>
              <Tooltip
                content="A descriptive name for this connector. This helps identify it in your list of data sources."
                variant="info"
              >
                <InformationCircleIcon className="h-4 w-4 text-muted-foreground cursor-help" />
              </Tooltip>
            </div>
            <Input
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              required
              helperText="Choose a name that identifies this data source"
            />
          </div>

          <div>
            <div className="flex items-center gap-2 mb-1">
              <label className="block text-sm font-medium">Type</label>
              <Tooltip
                content="Select the type of data source you want to connect. Each type has specific configuration requirements."
                variant="info"
              >
                <InformationCircleIcon className="h-4 w-4 text-muted-foreground cursor-help" />
              </Tooltip>
            </div>
            <Select
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value })}
              options={connectorTypes}
              required
            />
            <HelpText
              variant="inline"
              content={
                <p className="text-xs">
                  Different connector types support different features. Some may require additional configuration.
                </p>
              }
            />
          </div>

          <div>
            <div className="flex items-center gap-2 mb-1">
              <label className="block text-sm font-medium">Connection String</label>
              <Tooltip
                content="Connection credentials will be encrypted and stored securely. Never share connection strings."
                variant="warning"
              >
                <InformationCircleIcon className="h-4 w-4 text-muted-foreground cursor-help" />
              </Tooltip>
            </div>
            <Input
              type="password"
              value={formData.connectionString}
              onChange={(e) => setFormData({ ...formData, connectionString: e.target.value })}
              placeholder="Connection details stored securely"
              required
            />
            <HelpText
              variant="inline"
              content={
                <p className="text-xs">
                  Format depends on connector type. For databases, use connection strings like: postgresql://user:pass@host:port/dbname
                </p>
              }
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="enabled"
              checked={formData.enabled}
              onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
              className="rounded"
            />
            <label htmlFor="enabled" className="text-sm">Enabled</label>
          </div>

          <div className="flex gap-2 pt-4">
            <Button type="submit">Save</Button>
            <Button type="button" variant="outline" onClick={onCancel}>Cancel</Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}

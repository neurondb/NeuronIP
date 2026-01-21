'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { showToast } from '@/components/ui/Toast'

interface DataSourceWizardData {
  type: string
  name: string
  connection: {
    host: string
    port: string
    database: string
    username: string
    password: string
  }
  schemas: string[]
  schedule: {
    enabled: boolean
    frequency: string
    cron?: string
  }
  transformation: {
    enabled: boolean
    rules: string
  }
}

const dataSourceTypes = [
  { id: 'postgresql', name: 'PostgreSQL', description: 'Connect to PostgreSQL database' },
  { id: 'mysql', name: 'MySQL', description: 'Connect to MySQL database' },
  { id: 's3', name: 'Amazon S3', description: 'Connect to S3 bucket' },
  { id: 'api', name: 'REST API', description: 'Connect to REST API' },
]

function TypeStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as DataSourceWizardData) || { type: '' }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Select the type of data source you want to connect.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {dataSourceTypes.map((type) => (
          <Card
            key={type.id}
            hover
            className={wizardData.type === type.id ? 'ring-2 ring-primary' : ''}
            onClick={() => updateData({ type: type.id, name: type.name })}
          >
            <CardContent className="p-4">
              <h3 className="font-semibold mb-1">{type.name}</h3>
              <p className="text-sm text-muted-foreground">{type.description}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

function ConnectionStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as DataSourceWizardData) || {
    connection: { host: '', port: '', database: '', username: '', password: '' },
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Enter your connection details. These will be securely stored.
      </p>
      <div className="grid grid-cols-2 gap-4">
        <Input
          label="Host"
          value={wizardData.connection?.host || ''}
          onChange={(e) =>
            updateData({
              connection: { ...wizardData.connection, host: e.target.value },
            })
          }
          placeholder="localhost"
          required
        />
        <Input
          label="Port"
          value={wizardData.connection?.port || ''}
          onChange={(e) =>
            updateData({
              connection: { ...wizardData.connection, port: e.target.value },
            })
          }
          placeholder="5432"
          required
        />
      </div>
      <Input
        label="Database"
        value={wizardData.connection?.database || ''}
        onChange={(e) =>
          updateData({
            connection: { ...wizardData.connection, database: e.target.value },
          })
        }
        placeholder="database_name"
        required
      />
      <Input
        label="Username"
        value={wizardData.connection?.username || ''}
        onChange={(e) =>
          updateData({
            connection: { ...wizardData.connection, username: e.target.value },
          })
        }
        placeholder="username"
        required
      />
      <Input
        label="Password"
        type="password"
        value={wizardData.connection?.password || ''}
        onChange={(e) =>
          updateData({
            connection: { ...wizardData.connection, password: e.target.value },
          })
        }
        placeholder="password"
        required
      />
    </div>
  )
}

function TestStep({ data }: WizardStepProps) {
  const [testResult, setTestResult] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')

  const handleTest = () => {
    setTestResult('testing')
    setTimeout(() => {
      setTestResult('success')
    }, 2000)
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Test the connection to ensure it's configured correctly.
      </p>
      <Button onClick={handleTest} disabled={testResult === 'testing'} isLoading={testResult === 'testing'}>
        Test Connection
      </Button>
      {testResult === 'success' && (
        <div className="text-green-600 text-sm">Connection successful!</div>
      )}
    </div>
  )
}

function SchemaStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as DataSourceWizardData) || { schemas: [] }
  const [selectedSchemas, setSelectedSchemas] = useState<string[]>(wizardData.schemas || [])
  const availableSchemas = ['public', 'analytics', 'warehouse', 'staging']

  const toggleSchema = (schema: string) => {
    const newSelection = selectedSchemas.includes(schema)
      ? selectedSchemas.filter((s) => s !== schema)
      : [...selectedSchemas, schema]
    setSelectedSchemas(newSelection)
    updateData({ schemas: newSelection })
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Select which schemas to sync. You can change this later.
      </p>
      <div className="space-y-2">
        {availableSchemas.map((schema) => (
          <div key={schema} className="flex items-center gap-2">
            <input
              type="checkbox"
              id={`schema-${schema}`}
              checked={selectedSchemas.includes(schema)}
              onChange={() => toggleSchema(schema)}
              className="rounded"
            />
            <label htmlFor={`schema-${schema}`} className="text-sm">
              {schema}
            </label>
          </div>
        ))}
      </div>
    </div>
  )
}

function ScheduleStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as DataSourceWizardData) || {
    schedule: { enabled: false, frequency: 'daily' },
  }
  const [enabled, setEnabled] = useState(wizardData.schedule?.enabled || false)
  const [frequency, setFrequency] = useState(wizardData.schedule?.frequency || 'daily')

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="schedule-enabled"
          checked={enabled}
          onChange={(e) => {
            setEnabled(e.target.checked)
            updateData({
              schedule: { ...wizardData.schedule, enabled: e.target.checked },
            })
          }}
          className="rounded"
        />
        <label htmlFor="schedule-enabled" className="text-sm font-medium">
          Enable Scheduled Sync
        </label>
      </div>
      {enabled && (
        <div>
          <label className="block text-sm font-medium mb-2">Sync Frequency</label>
          <select
            value={frequency}
            onChange={(e) => {
              setFrequency(e.target.value)
              updateData({
                schedule: { ...wizardData.schedule, frequency: e.target.value },
              })
            }}
            className="w-full px-3 py-2 border rounded-lg"
          >
            <option value="hourly">Hourly</option>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="custom">Custom (Cron)</option>
          </select>
          {frequency === 'custom' && (
            <Input
              label="Cron Expression"
              value={wizardData.schedule?.cron || ''}
              onChange={(e) =>
                updateData({
                  schedule: { ...wizardData.schedule, cron: e.target.value },
                })
              }
              placeholder="0 0 * * *"
              className="mt-2"
            />
          )}
        </div>
      )}
    </div>
  )
}

function TransformationStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as DataSourceWizardData) || {
    transformation: { enabled: false, rules: '' },
  }
  const [enabled, setEnabled] = useState(wizardData.transformation?.enabled || false)
  const [rules, setRules] = useState(wizardData.transformation?.rules || '')

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          id="transformation-enabled"
          checked={enabled}
          onChange={(e) => {
            setEnabled(e.target.checked)
            updateData({
              transformation: { ...wizardData.transformation, enabled: e.target.checked },
            })
          }}
          className="rounded"
        />
        <label htmlFor="transformation-enabled" className="text-sm font-medium">
          Enable Data Transformation
        </label>
      </div>
      {enabled && (
        <div>
          <label className="block text-sm font-medium mb-2">Transformation Rules (JSON)</label>
          <textarea
            value={rules}
            onChange={(e) => {
              setRules(e.target.value)
              updateData({
                transformation: { ...wizardData.transformation, rules: e.target.value },
              })
            }}
            placeholder='{"transform": "uppercase", "fields": ["name"]}'
            className="w-full px-3 py-2 border rounded-lg font-mono text-sm"
            rows={6}
          />
        </div>
      )}
    </div>
  )
}

function ReviewStep({ data }: WizardStepProps) {
  const wizardData = data as DataSourceWizardData

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="p-4 space-y-3">
          <div>
            <h4 className="font-semibold mb-1">Type</h4>
            <p className="text-sm">{wizardData.type || 'Not set'}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Connection</h4>
            <p className="text-sm">
              {wizardData.connection?.host ? 'Configured' : 'Not configured'}
            </p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Schemas ({wizardData.schemas?.length || 0})</h4>
            <p className="text-sm">{wizardData.schemas?.join(', ') || 'None selected'}</p>
          </div>
          {wizardData.schedule?.enabled && (
            <div>
              <h4 className="font-semibold mb-1">Schedule</h4>
              <p className="text-sm capitalize">{wizardData.schedule?.frequency || 'Not set'}</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

interface DataSourceSetupWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function DataSourceSetupWizard({
  onComplete,
  onCancel,
}: DataSourceSetupWizardProps) {
  const handleComplete = async (data: DataSourceWizardData) => {
    // In real implementation, call API to create data source
    showToast('Data source created successfully', 'success')
    if (onComplete) onComplete()
  }

  const validateType = (data: any): boolean => {
    return !!(data as DataSourceWizardData).type
  }

  const validateConnection = (data: any): boolean => {
    const wizardData = data as DataSourceWizardData
    return !!(
      wizardData.connection?.host &&
      wizardData.connection?.port &&
      wizardData.connection?.database &&
      wizardData.connection?.username
    )
  }

  const steps: WizardStep[] = [
    {
      id: 'type',
      title: 'Choose Data Source Type',
      description: 'Select the type of data source',
      component: TypeStep,
      validate: validateType,
    },
    {
      id: 'connection',
      title: 'Connection Details',
      description: 'Enter connection information',
      component: ConnectionStep,
      validate: validateConnection,
    },
    {
      id: 'test',
      title: 'Test Connection',
      description: 'Verify the connection works',
      component: TestStep,
      canSkip: true,
    },
    {
      id: 'schemas',
      title: 'Select Schemas',
      description: 'Choose which schemas to sync',
      component: SchemaStep,
      canSkip: true,
    },
    {
      id: 'schedule',
      title: 'Sync Schedule',
      description: 'Configure when to sync data',
      component: ScheduleStep,
      canSkip: true,
      isOptional: true,
    },
    {
      id: 'transformation',
      title: 'Data Transformation',
      description: 'Set up transformation rules (optional)',
      component: TransformationStep,
      canSkip: true,
      isOptional: true,
    },
    {
      id: 'review',
      title: 'Review & Create',
      description: 'Review your configuration',
      component: ReviewStep,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Set Up Data Source"
      description="Connect a new data source to NeuronIP"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

'use client'

import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Card, CardContent } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import { Textarea } from '@/components/ui/Textarea'
import { showToast } from '@/components/ui/Toast'
import HelpText from '@/components/ui/HelpText'
import { useMutation } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'

interface WarehouseWizardData {
  name: string
  description: string
  connection: {
    host: string
    port: string
    database: string
    username: string
    password: string
  }
  sampleQueries: string[]
}

function WelcomeStep({ data, updateData, nextStep }: WizardStepProps) {
  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Welcome to Warehouse Setup</h3>
        <p className="text-sm text-muted-foreground">
          Set up your data warehouse connection to start querying with natural language or SQL.
        </p>
      </div>
      <HelpText
        variant="inline"
        title="What is Warehouse Q&A?"
        content={
          <p>
            Warehouse Q&A allows you to ask natural language questions about your data warehouse.
            The system will automatically generate SQL queries, execute them safely, and provide
            visualizations and explanations.
          </p>
        }
        link="/docs/features/warehouse-qa"
        linkText="Learn more about Warehouse Q&A"
      />
      <div className="pt-4">
        <Button onClick={nextStep} className="w-full sm:w-auto">
          Get Started →
        </Button>
      </div>
    </div>
  )
}

function SchemaInfoStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as WarehouseWizardData) || { name: '', description: '' }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        First, let's create a schema to organize your warehouse queries.
      </p>
      <Input
        label="Schema Name"
        value={wizardData.name || ''}
        onChange={(e) => updateData({ name: e.target.value })}
        placeholder="e.g., sales_data"
        required
        helperText="A unique name for your warehouse schema"
      />
      <Textarea
        label="Description"
        value={wizardData.description || ''}
        onChange={(e) => updateData({ description: e.target.value })}
        placeholder="Describe what data this schema contains"
        rows={3}
        helperText="Optional description of the schema"
      />
      <HelpText
        variant="inline"
        content={
          <p>
            Schemas help organize your warehouse queries. You can create multiple schemas for
            different data sources or purposes.
          </p>
        }
      />
    </div>
  )
}

function ConnectionStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as WarehouseWizardData) || {
    connection: { host: '', port: '5432', database: '', username: '', password: '' },
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground mb-4">
        Connect to your data warehouse. Credentials will be securely stored.
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
          value={wizardData.connection?.port || '5432'}
          onChange={(e) =>
            updateData({
              connection: { ...wizardData.connection, port: e.target.value },
            })
          }
          placeholder="5432"
          type="number"
        />
        <Input
          label="Database"
          value={wizardData.connection?.database || ''}
          onChange={(e) =>
            updateData({
              connection: { ...wizardData.connection, database: e.target.value },
            })
          }
          placeholder="mydb"
          required
          className="col-span-2"
        />
        <Input
          label="Username"
          value={wizardData.connection?.username || ''}
          onChange={(e) =>
            updateData({
              connection: { ...wizardData.connection, username: e.target.value },
            })
          }
          placeholder="user"
          required
        />
        <Input
          label="Password"
          value={wizardData.connection?.password || ''}
          onChange={(e) =>
            updateData({
              connection: { ...wizardData.connection, password: e.target.value },
            })
          }
          placeholder="••••••••"
          type="password"
          required
        />
      </div>
      <HelpText
        variant="inline"
        content={
          <p>
            Your credentials are encrypted and stored securely. Only you and authorized users can
            access this connection.
          </p>
        }
      />
    </div>
  )
}

function CompleteStep({ data }: WizardStepProps) {
  const wizardData = data as WarehouseWizardData
  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Setup Complete!</h3>
        <p className="text-sm text-muted-foreground">
          Your warehouse schema "{wizardData.name}" is ready to use.
        </p>
      </div>
      <div className="rounded-lg border p-4 bg-muted/50 space-y-2">
        <p className="text-sm font-medium">What's next?</p>
        <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
          <li>Start querying your warehouse with natural language</li>
          <li>Explore existing tables and schemas</li>
          <li>Create saved queries for quick access</li>
          <li>Set up query governance rules</li>
        </ul>
      </div>
      <HelpText
        variant="inline"
        content={
          <p>
            Try asking a question like "What are the top 10 products by sales?" and watch the
            system generate and execute SQL for you.
          </p>
        }
      />
    </div>
  )
}

interface WarehouseSetupWizardProps {
  onComplete?: () => void
  onCancel?: () => void
}

export default function WarehouseSetupWizard({
  onComplete,
  onCancel,
}: WarehouseSetupWizardProps) {
  const createSchema = useMutation({
    mutationFn: async (data: WarehouseWizardData) => {
      const response = await apiClient.post('/warehouse/schemas', {
        name: data.name,
        description: data.description,
        config: {
          connection: data.connection,
        },
      })
      return response.data
    },
  })

  const handleComplete = async (data: WarehouseWizardData) => {
    try {
      await createSchema.mutateAsync(data)
      showToast('Warehouse schema created successfully!', 'success')
      onComplete?.()
    } catch (error: any) {
      showToast(
        error?.message || 'Failed to create warehouse schema. Please try again.',
        'error'
      )
    }
  }

  const steps: WizardStep[] = [
    {
      id: 'welcome',
      title: 'Welcome',
      description: 'Introduction to Warehouse Q&A',
      component: WelcomeStep,
      canSkip: false,
    },
    {
      id: 'schema',
      title: 'Schema Information',
      description: 'Name and describe your schema',
      component: SchemaInfoStep,
      validate: async (data) => {
        const wizardData = data as WarehouseWizardData
        return !!(wizardData.name && wizardData.name.trim().length > 0)
      },
    },
    {
      id: 'connection',
      title: 'Connection Details',
      description: 'Configure warehouse connection',
      component: ConnectionStep,
      validate: async (data) => {
        const wizardData = data as WarehouseWizardData
        const conn = wizardData.connection
        return !!(
          conn?.host &&
          conn?.database &&
          conn?.username &&
          conn?.password
        )
      },
    },
    {
      id: 'complete',
      title: 'Complete',
      description: "You're all set!",
      component: CompleteStep,
      canSkip: false,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Warehouse Setup Wizard"
      description="Set up your data warehouse connection in a few simple steps"
      onComplete={handleComplete}
      onCancel={onCancel}
      showProgress={true}
    />
  )
}

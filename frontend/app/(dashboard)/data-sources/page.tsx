'use client'

import { PlusIcon } from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import { useState } from 'react'

import ConnectorForm from '@/components/data-sources/ConnectorForm'
import ConnectorList from '@/components/data-sources/ConnectorList'
import CredentialsVault from '@/components/data-sources/CredentialsVault'
import DataSourceSetupWizard from '@/components/data-sources/DataSourceSetupWizard'
import ScheduleEditor from '@/components/data-sources/ScheduleEditor'
import Button from '@/components/ui/Button'
import Modal from '@/components/ui/Modal'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import apiClient from '@/lib/api/client'


type ViewMode = 'list' | 'add' | 'edit' | 'schedule' | 'credentials' | 'wizard'

interface Connector {
  id?: string
  schedule?: string
  metadata?: Record<string, unknown>
  configuration?: Record<string, unknown>
  credentials?: Record<string, unknown>
  [key: string]: unknown
}

export default function DataSourcesPage() {
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [selectedConnector, setSelectedConnector] = useState<Connector | null>(null)
  const [listRefresh, setListRefresh] = useState(0)

  const handleAddConnector = async (data: Record<string, unknown>) => {
    try {
      await apiClient.post('/connectors', data)
      setListRefresh((n) => n + 1)
      setViewMode('list')
    } catch (err: unknown) {
      const message = err && typeof err === 'object' && 'message' in err ? String((err as { message: string }).message) : 'Failed to create connector'
      console.error(message)
    }
  }

  const handleEditConnector = async (data: Record<string, unknown>) => {
    if (!selectedConnector?.id) return
    try {
      await apiClient.put(`/connectors/${selectedConnector.id}`, { ...selectedConnector, ...data })
      setListRefresh((n) => n + 1)
      setViewMode('list')
      setSelectedConnector(null)
    } catch (err: unknown) {
      const message = err && typeof err === 'object' && 'message' in err ? String((err as { message: string }).message) : 'Failed to update connector'
      console.error(message)
    }
  }

  const handleSaveSchedule = async (schedule: string) => {
    if (!selectedConnector?.id) return
    try {
      const { data: connector } = await apiClient.get<Connector>(`/connectors/${selectedConnector.id}`)
      const metadata = { ...(connector.metadata || {}), schedule }
      await apiClient.put(`/connectors/${selectedConnector.id}`, { ...connector, metadata })
      setListRefresh((n) => n + 1)
      setViewMode('list')
      setSelectedConnector(null)
    } catch (err: unknown) {
      const message = err && typeof err === 'object' && 'message' in err ? String((err as { message: string }).message) : 'Failed to save schedule'
      console.error(message)
    }
  }

  const handleSaveCredentials = async (credentials: Record<string, unknown>) => {
    if (!selectedConnector?.id) return
    try {
      const { data: connector } = await apiClient.get<Connector>(`/connectors/${selectedConnector.id}`)
      await apiClient.put(`/connectors/${selectedConnector.id}`, { ...connector, credentials })
      setListRefresh((n) => n + 1)
      setViewMode('list')
      setSelectedConnector(null)
    } catch (err: unknown) {
      const message = err && typeof err === 'object' && 'message' in err ? String((err as { message: string }).message) : 'Failed to save credentials'
      console.error(message)
    }
  }

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Data Sources</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Connectors for PostgreSQL, S3, APIs, SaaS tools. Sync status, schedules, and credentials.
            </p>
          </div>
          {viewMode === 'list' && (
            <Button onClick={() => setViewMode('wizard')}>
              <PlusIcon className="h-4 w-4 mr-2" />
              Add Data Source
            </Button>
          )}
        </div>
      </motion.div>

      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        {viewMode === 'list' && (
          <ConnectorList
            key={listRefresh}
            onEdit={(c) => {
              setSelectedConnector(c as Connector)
              setViewMode('edit')
            }}
            onSchedule={(c) => {
              setSelectedConnector(c as Connector)
              setViewMode('schedule')
            }}
            onCredentials={(c) => {
              setSelectedConnector(c as Connector)
              setViewMode('credentials')
            }}
          />
        )}
        {viewMode === 'wizard' && (
          <Modal
            open={viewMode === 'wizard'}
            onOpenChange={(open) => setViewMode(open ? 'wizard' : 'list')}
            size="xl"
            title="Set Up Data Source"
          >
            <DataSourceSetupWizard
              onComplete={() => setViewMode('list')}
              onCancel={() => setViewMode('list')}
            />
          </Modal>
        )}
        {viewMode === 'add' && (
          <ConnectorForm
            onSubmit={handleAddConnector}
            onCancel={() => setViewMode('list')}
          />
        )}
        {viewMode === 'edit' && (
          <ConnectorForm
            onSubmit={handleEditConnector}
            onCancel={() => setViewMode('list')}
            initialData={selectedConnector}
          />
        )}
        {viewMode === 'schedule' && (
          <ScheduleEditor
            schedule={selectedConnector?.schedule}
            onSave={handleSaveSchedule}
            onCancel={() => setViewMode('list')}
          />
        )}
        {viewMode === 'credentials' && selectedConnector?.id && (
          <CredentialsVault
            connectorId={selectedConnector.id}
            onSave={handleSaveCredentials}
          />
        )}
      </motion.div>
    </motion.div>
  )
}

'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import ConnectorList from '@/components/data-sources/ConnectorList'
import ConnectorForm from '@/components/data-sources/ConnectorForm'
import ScheduleEditor from '@/components/data-sources/ScheduleEditor'
import CredentialsVault from '@/components/data-sources/CredentialsVault'
import Button from '@/components/ui/Button'
import { PlusIcon } from '@heroicons/react/24/outline'

type ViewMode = 'list' | 'add' | 'edit' | 'schedule' | 'credentials'

export default function DataSourcesPage() {
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [selectedConnector, setSelectedConnector] = useState<any>(null)

  const handleAddConnector = (data: any) => {
    // API call to create connector
    console.log('Creating connector:', data)
    setViewMode('list')
  }

  const handleEditConnector = (data: any) => {
    // API call to update connector
    console.log('Updating connector:', data)
    setViewMode('list')
  }

  const handleSaveSchedule = (schedule: string) => {
    // API call to save schedule
    console.log('Saving schedule:', schedule)
    setViewMode('list')
  }

  const handleSaveCredentials = (credentials: any) => {
    // API call to save credentials
    console.log('Saving credentials:', credentials)
    setViewMode('list')
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
            <Button onClick={() => setViewMode('add')}>
              <PlusIcon className="h-4 w-4 mr-2" />
              Add Connector
            </Button>
          )}
        </div>
      </motion.div>

      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        {viewMode === 'list' && <ConnectorList />}
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
        {viewMode === 'credentials' && (
          <CredentialsVault
            connectorId={selectedConnector?.id}
            onSave={handleSaveCredentials}
          />
        )}
      </motion.div>
    </motion.div>
  )
}

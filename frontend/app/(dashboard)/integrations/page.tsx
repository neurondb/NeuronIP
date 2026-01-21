'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import IntegrationCard from '@/components/integrations/IntegrationCard'
import IntegrationConfigDialog from '@/components/integrations/IntegrationConfigDialog'
import IntegrationSetupWizard from '@/components/integrations/IntegrationSetupWizard'
import WebhookManager from '@/components/integrations/WebhookManager'
import { useIntegrations, useIntegrationHealth } from '@/lib/api/queries'
import Button from '@/components/ui/Button'
import Modal from '@/components/ui/Modal'
import { Card, CardContent } from '@/components/ui/Card'
import { PlusIcon } from '@heroicons/react/24/outline'

export default function IntegrationsPage() {
  const [selectedIntegrationId, setSelectedIntegrationId] = useState<string | null>(null)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showWizard, setShowWizard] = useState(false)
  const [createType, setCreateType] = useState<string>('')
  const [activeTab, setActiveTab] = useState<'integrations' | 'webhooks'>('integrations')

  const { data: integrations, isLoading, refetch } = useIntegrations()
  const { data: health } = useIntegrationHealth()

  const handleCreate = (type: string) => {
    setCreateType(type)
    setShowWizard(true)
  }

  const handleEdit = (id: string) => {
    setSelectedIntegrationId(id)
    setShowCreateDialog(true)
  }

  const handleDialogSuccess = () => {
    setShowCreateDialog(false)
    setSelectedIntegrationId(null)
    setCreateType('')
    refetch()
  }

  const handleDialogCancel = () => {
    setShowCreateDialog(false)
    setSelectedIntegrationId(null)
    setCreateType('')
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
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Integrations</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Slack, Teams, CRM, ERP, ticketing, email. Webhooks and triggers.
            </p>
          </div>
          {health && (
            <div className="text-right">
              <div className="text-sm text-muted-foreground">Health Status</div>
              <div className="text-lg font-bold">
                <span className="text-green-600">{health.active || 0}</span> active /{' '}
                <span className="text-red-600">{health.error || 0}</span> errors
              </div>
            </div>
          )}
        </div>
      </motion.div>

      <motion.div variants={slideUp} className="flex-shrink-0">
        <div className="flex gap-2 border-b border-border">
          <Button
            variant={activeTab === 'integrations' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('integrations')}
            size="sm"
          >
            Integrations
          </Button>
          <Button
            variant={activeTab === 'webhooks' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('webhooks')}
            size="sm"
          >
            Webhooks
          </Button>
        </div>
      </motion.div>

      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        {activeTab === 'integrations' && (
          <div className="space-y-4">
            <div className="flex gap-2 flex-wrap">
              <Button onClick={() => handleCreate('slack')} variant="outline" size="sm">
                <PlusIcon className="h-4 w-4 mr-1" />
                Add Slack
              </Button>
              <Button onClick={() => handleCreate('teams')} variant="outline" size="sm">
                <PlusIcon className="h-4 w-4 mr-1" />
                Add Teams
              </Button>
              <Button onClick={() => handleCreate('email')} variant="outline" size="sm">
                <PlusIcon className="h-4 w-4 mr-1" />
                Add Email
              </Button>
              <Button onClick={() => handleCreate('helpdesk')} variant="outline" size="sm">
                <PlusIcon className="h-4 w-4 mr-1" />
                Add Helpdesk
              </Button>
            </div>

            {showCreateDialog && (
              <IntegrationConfigDialog
                integrationId={selectedIntegrationId || undefined}
                integrationType={createType}
                onSuccess={handleDialogSuccess}
                onCancel={handleDialogCancel}
              />
            )}

            {isLoading ? (
              <div className="text-center py-8 text-muted-foreground">Loading integrations...</div>
            ) : !integrations || (Array.isArray(integrations) && integrations.length === 0) ? (
              <Card>
                <CardContent className="text-center py-12">
                  <p className="text-muted-foreground mb-4">No integrations configured. Create one to get started.</p>
                  <div className="flex gap-2 justify-center flex-wrap">
                    <Button onClick={() => handleCreate('slack')} variant="outline" size="sm">
                      <PlusIcon className="h-4 w-4 mr-1" />
                      Add Slack
                    </Button>
                    <Button onClick={() => handleCreate('teams')} variant="outline" size="sm">
                      <PlusIcon className="h-4 w-4 mr-1" />
                      Add Teams
                    </Button>
                    <Button onClick={() => handleCreate('email')} variant="outline" size="sm">
                      <PlusIcon className="h-4 w-4 mr-1" />
                      Add Email
                    </Button>
                    <Button onClick={() => handleCreate('helpdesk')} variant="outline" size="sm">
                      <PlusIcon className="h-4 w-4 mr-1" />
                      Add Helpdesk
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {(Array.isArray(integrations) ? integrations : []).map((integration: any) => (
                  <IntegrationCard
                    key={integration.id}
                    integration={integration}
                    onEdit={handleEdit}
                    onRefresh={refetch}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {activeTab === 'webhooks' && <WebhookManager />}
      </motion.div>
    </motion.div>
  )
}

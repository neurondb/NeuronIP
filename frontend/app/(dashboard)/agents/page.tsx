'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import AgentList from '@/components/agents/AgentList'
import AgentDetail from '@/components/agents/AgentDetail'
import CreateAgentDialog from '@/components/agents/CreateAgentDialog'
import AgentCreationWizard from '@/components/agents/AgentCreationWizard'
import Modal from '@/components/ui/Modal'
import Button from '@/components/ui/Button'

export default function AgentsPage() {
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [showWizard, setShowWizard] = useState(false)

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
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Agent Hub</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Manage and deploy AI agents. Agent performance, memory, and behavior.
            </p>
          </div>
          {selectedAgentId ? (
            <Button variant="outline" onClick={() => setSelectedAgentId(null)}>
              Back to List
            </Button>
          ) : (
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => setCreateDialogOpen(true)}>
                Quick Create
              </Button>
              <Button onClick={() => setShowWizard(true)}>
                Create with Wizard
              </Button>
            </div>
          )}
        </div>
      </motion.div>
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        {selectedAgentId ? (
          <AgentDetail agentId={selectedAgentId} />
        ) : (
          <AgentList
            onSelectAgent={setSelectedAgentId}
            onCreateNew={() => setCreateDialogOpen(true)}
          />
        )}
      </motion.div>

      <CreateAgentDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onCreated={() => {
          setCreateDialogOpen(false)
        }}
      />

      <Modal
        open={showWizard}
        onOpenChange={setShowWizard}
        size="xl"
        title="Create Agent"
      >
        <AgentCreationWizard
          onComplete={() => setShowWizard(false)}
          onCancel={() => setShowWizard(false)}
        />
      </Modal>
    </motion.div>
  )
}

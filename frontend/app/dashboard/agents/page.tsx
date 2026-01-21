'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import AgentList from '@/components/agents/AgentList'
import AgentDetail from '@/components/agents/AgentDetail'
import Button from '@/components/ui/Button'

export default function AgentsPage() {
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null)

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
          {selectedAgentId && (
            <Button variant="outline" onClick={() => setSelectedAgentId(null)}>
              Back to List
            </Button>
          )}
        </div>
      </motion.div>
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        {selectedAgentId ? (
          <AgentDetail agentId={selectedAgentId} />
        ) : (
          <AgentList onSelectAgent={setSelectedAgentId} />
        )}
      </motion.div>
    </motion.div>
  )
}

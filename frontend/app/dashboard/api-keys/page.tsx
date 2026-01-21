'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import APIKeyList from '@/components/api-keys/APIKeyList'
import CreateAPIKeyDialog from '@/components/api-keys/CreateAPIKeyDialog'
import Button from '@/components/ui/Button'
import { PlusIcon } from '@heroicons/react/24/outline'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function APIKeysPage() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  return (
    <>
      <motion.div
        variants={staggerContainer}
        initial="hidden"
        animate="visible"
        className="space-y-3 sm:space-y-4 flex flex-col h-full"
      >
        {/* Page Header */}
        <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl sm:text-3xl font-bold text-foreground">API Keys</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Manage API keys and rate limits
              </p>
            </div>
            <Button onClick={() => setCreateDialogOpen(true)}>
              <PlusIcon className="h-4 w-4 mr-2" />
              Create API Key
            </Button>
          </div>
        </motion.div>

        {/* API Keys List */}
        <motion.div variants={slideUp} className="flex-1 min-h-0">
          <APIKeyList />
        </motion.div>
      </motion.div>

      <CreateAPIKeyDialog open={createDialogOpen} onOpenChange={setCreateDialogOpen} />
    </>
  )
}

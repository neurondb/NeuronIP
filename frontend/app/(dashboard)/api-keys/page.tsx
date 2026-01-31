'use client'

import { PlusIcon } from '@heroicons/react/24/outline'
import { useState } from 'react'

import APIKeyList from '@/components/api-keys/APIKeyList'
import CreateAPIKeyDialog from '@/components/api-keys/CreateAPIKeyDialog'
import PageTemplate from '@/components/layout/PageTemplate'
import Button from '@/components/ui/Button'

export default function APIKeysPage() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  return (
    <>
      <PageTemplate
        title="API Keys"
        description="Manage API keys and rate limits"
        archetype="list-detail"
        actions={
          <Button onClick={() => setCreateDialogOpen(true)}>
            <PlusIcon className="h-4 w-4 mr-2" />
            Create API Key
          </Button>
        }
      >
        <APIKeyList onCreateNew={() => setCreateDialogOpen(true)} />
      </PageTemplate>
      <CreateAPIKeyDialog open={createDialogOpen} onOpenChange={setCreateDialogOpen} />
    </>
  )
}

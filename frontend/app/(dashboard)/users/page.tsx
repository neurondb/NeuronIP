'use client'

import { PlusIcon } from '@heroicons/react/24/outline'
import { useState } from 'react'

import PageTemplate from '@/components/layout/PageTemplate'
import Button from '@/components/ui/Button'
import AddUserDialog from '@/components/users/AddUserDialog'
import UserList from '@/components/users/UserList'

export default function UsersPage() {
  const [_selectedUserId, setSelectedUserId] = useState<string | null>(null)
  const [isAddUserOpen, setIsAddUserOpen] = useState(false)

  const handleAddUser = () => {
    setIsAddUserOpen(true)
  }

  return (
    <PageTemplate
      title="Users"
      description="Manage user accounts and permissions"
      actions={
        <Button onClick={handleAddUser}>
          <PlusIcon className="h-4 w-4 mr-2" />
          Add User
        </Button>
      }
      archetype="list-detail"
    >
      <UserList onSelectUser={setSelectedUserId} onCreateNew={handleAddUser} />
      <AddUserDialog
        open={isAddUserOpen}
        onOpenChange={setIsAddUserOpen}
        onCreated={() => {
          window.location.reload()
        }}
      />
    </PageTemplate>
  )
}

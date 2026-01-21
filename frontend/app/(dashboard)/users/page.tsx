'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import UserList from '@/components/users/UserList'
import Button from '@/components/ui/Button'
import AddUserDialog from '@/components/users/AddUserDialog'
import { PlusIcon } from '@heroicons/react/24/outline'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function UsersPage() {
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)
  const [isAddUserOpen, setIsAddUserOpen] = useState(false)

  const handleAddUser = () => {
    setIsAddUserOpen(true)
  }

  return (
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
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Users</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Manage user accounts and permissions
            </p>
          </div>
          <Button onClick={handleAddUser}>
            <PlusIcon className="h-4 w-4 mr-2" />
            Add User
          </Button>
        </div>
      </motion.div>

      {/* Users List */}
      <motion.div variants={slideUp} className="flex-1 min-h-0">
        <UserList onSelectUser={setSelectedUserId} onCreateNew={handleAddUser} />
      </motion.div>

      {/* Add User Dialog */}
      <AddUserDialog
        open={isAddUserOpen}
        onOpenChange={setIsAddUserOpen}
        onCreated={() => {
          // Refresh user list
          window.location.reload()
        }}
      />
    </motion.div>
  )
}

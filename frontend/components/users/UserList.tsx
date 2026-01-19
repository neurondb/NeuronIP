'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { MagnifyingGlassIcon, UserCircleIcon } from '@heroicons/react/24/outline'
import { useUsers } from '@/lib/api/queries'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'user' | 'viewer'
  status: 'active' | 'inactive'
  created_at: Date
}

export default function UserList({ onSelectUser }: { onSelectUser?: (userId: string) => void }) {
  const [searchQuery, setSearchQuery] = useState('')
  const { data: users, isLoading } = useUsers()

  const userList: User[] = users?.users || [
    {
      id: '1',
      name: 'John Doe',
      email: 'john@example.com',
      role: 'admin',
      status: 'active',
      created_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30),
    },
    {
      id: '2',
      name: 'Jane Smith',
      email: 'jane@example.com',
      role: 'user',
      status: 'active',
      created_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 15),
    },
    {
      id: '3',
      name: 'Bob Wilson',
      email: 'bob@example.com',
      role: 'viewer',
      status: 'inactive',
      created_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 60),
    },
  ]

  const filteredUsers = userList.filter(
    (user) =>
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const getRoleColor = (role: User['role']) => {
    switch (role) {
      case 'admin':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200'
      case 'user':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
      case 'viewer':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
    }
  }

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">Users</CardTitle>
        <CardDescription className="text-xs">Manage user accounts and permissions</CardDescription>
        <div className="relative mt-3">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search users..."
            className={cn(
              'w-full rounded-lg border border-border bg-background py-2 pl-9 pr-3',
              'text-sm placeholder:text-muted-foreground',
              'focus:outline-none focus:ring-2 focus:ring-ring'
            )}
          />
        </div>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto min-h-0">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loading size="md" />
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>User</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredUsers.map((user, index) => (
                <motion.tr
                  key={user.id}
                  variants={slideUp}
                  initial="hidden"
                  animate="visible"
                  transition={{ ...transition, delay: index * 0.03 }}
                  onClick={() => onSelectUser?.(user.id)}
                  className="hover:bg-muted/50 cursor-pointer transition-colors"
                >
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <UserCircleIcon className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <div className="font-medium">{user.name}</div>
                        <div className="text-xs text-muted-foreground">{user.email}</div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className={cn('px-2 py-1 rounded-full text-xs font-medium capitalize', getRoleColor(user.role))}>
                      {user.role}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className={cn(
                      'px-2 py-1 rounded-full text-xs font-medium',
                      user.status === 'active' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                    )}>
                      {user.status}
                    </span>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(user.created_at).toLocaleDateString()}
                  </TableCell>
                </motion.tr>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}

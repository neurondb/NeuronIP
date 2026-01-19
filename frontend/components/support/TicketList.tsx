'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import Button from '@/components/ui/Button'
import { FunnelIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline'
import { slideUp, transition } from '@/lib/animations/variants'
import { cn } from '@/lib/utils/cn'

interface Ticket {
  id: string
  title: string
  status: 'open' | 'in-progress' | 'resolved' | 'closed'
  priority: 'low' | 'medium' | 'high'
  createdAt: Date
  updatedAt: Date
}

interface TicketListProps {
  tickets: Ticket[]
}

export default function TicketList({ tickets }: TicketListProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | null>(null)

  const getStatusColor = (status: Ticket['status']) => {
    switch (status) {
      case 'open':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
      case 'in-progress':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
      case 'resolved':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
      case 'closed':
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
    }
  }

  const getPriorityColor = (priority: Ticket['priority']) => {
    switch (priority) {
      case 'high':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
      case 'low':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
    }
  }

  const filteredTickets = tickets.filter((ticket) => {
    const matchesSearch = ticket.title.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesStatus = !statusFilter || ticket.status === statusFilter
    return matchesSearch && matchesStatus
  })

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <div className="flex items-center justify-between mb-4">
          <div>
            <CardTitle className="text-base sm:text-lg">Support Tickets</CardTitle>
            <CardDescription className="text-xs mt-1">
              {filteredTickets.length} of {tickets.length} tickets
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search tickets..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className={cn(
                  'w-48 rounded-lg border border-border bg-background py-2 pl-9 pr-3',
                  'text-sm placeholder:text-muted-foreground',
                  'focus:outline-none focus:ring-2 focus:ring-ring'
                )}
              />
            </div>
            <Button variant="outline" size="sm">
              <FunnelIcon className="h-4 w-4 mr-2" />
              Filter
            </Button>
          </div>
        </div>
        {/* Status Filter Pills */}
        <div className="flex gap-2 flex-wrap">
          {(['open', 'in-progress', 'resolved', 'closed'] as const).map((status) => (
            <button
              key={status}
              onClick={() => setStatusFilter(statusFilter === status ? null : status)}
              className={cn(
                'px-3 py-1.5 rounded-full text-xs font-medium transition-colors',
                statusFilter === status
                  ? getStatusColor(status)
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              )}
            >
              {status.charAt(0).toUpperCase() + status.slice(1)} ({tickets.filter((t) => t.status === status).length})
            </button>
          ))}
        </div>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto min-h-0">
        {filteredTickets.length === 0 ? (
          <div className="text-center text-muted-foreground py-12">
            <p className="text-sm">No tickets found</p>
            {searchQuery && <p className="text-xs mt-1">Try adjusting your search or filters</p>}
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Priority</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Updated</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTickets.map((ticket, index) => (
                <motion.tr
                  key={ticket.id}
                  variants={slideUp}
                  initial="hidden"
                  animate="visible"
                  transition={{ ...transition, delay: index * 0.03 }}
                  className="hover:bg-muted/50 cursor-pointer transition-colors"
                >
                  <TableCell className="font-medium">{ticket.title}</TableCell>
                  <TableCell>
                    <span className={cn('px-2 py-1 rounded-full text-xs font-medium', getStatusColor(ticket.status))}>
                      {ticket.status}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className={cn('px-2 py-1 rounded-full text-xs font-medium', getPriorityColor(ticket.priority))}>
                      {ticket.priority}
                    </span>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {ticket.createdAt.toLocaleDateString()}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {ticket.updatedAt.toLocaleDateString()}
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

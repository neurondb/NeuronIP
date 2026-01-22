'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import MetricCard from '@/components/dashboard/MetricCard'
import TicketList from '@/components/support/TicketList'
import CreateTicketDialog from '@/components/support/CreateTicketDialog'
import FilterChips from '@/components/filters/FilterChips'
import Button from '@/components/ui/Button'
import { PlusIcon, FunnelIcon } from '@heroicons/react/24/outline'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import { useSupportTickets } from '@/lib/api/queries'

interface Ticket {
  id: string
  title: string
  status: 'open' | 'in-progress' | 'resolved' | 'closed'
  priority: 'low' | 'medium' | 'high'
  createdAt: Date
  updatedAt: Date
}

export default function SupportPage() {
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [filters, setFilters] = useState<Array<{ id: string; label: string; value: string; type: string }>>([])
  const [statusFilter, setStatusFilter] = useState<string | null>(null)

  // Fetch tickets from API
  const { data: ticketsData, isLoading, refetch } = useSupportTickets()
  
  // Transform API data to match Ticket interface
  const rawTickets = ticketsData?.tickets || ticketsData || []
  const tickets: Ticket[] = Array.isArray(rawTickets)
    ? rawTickets.map((ticket: any) => ({
        id: ticket.id || ticket.ticket_id || String(ticket),
        title: ticket.title || ticket.subject || 'Untitled',
        status: (ticket.status || 'open') as Ticket['status'],
        priority: (ticket.priority || 'medium') as Ticket['priority'],
        createdAt: ticket.created_at ? new Date(ticket.created_at) : new Date(),
        updatedAt: ticket.updated_at ? new Date(ticket.updated_at) : new Date(),
      }))
    : []

  // Calculate metrics
  const metrics = [
    {
      title: 'Total Tickets',
      value: tickets.length,
      description: 'All tickets',
      icon: <PlusIcon />,
    },
    {
      title: 'Open',
      value: tickets.filter((t) => t.status === 'open').length,
      description: 'Requires attention',
      icon: <PlusIcon />,
      trend: { value: 2, isPositive: false },
    },
    {
      title: 'In Progress',
      value: tickets.filter((t) => t.status === 'in-progress').length,
      description: 'Being handled',
      icon: <PlusIcon />,
    },
    {
      title: 'Resolved',
      value: tickets.filter((t) => t.status === 'resolved').length,
      description: 'Completed',
      icon: <PlusIcon />,
      trend: { value: 15, isPositive: true },
    },
  ]

  const filteredTickets = statusFilter
    ? tickets.filter((t) => t.status === statusFilter)
    : tickets

  const handleRemoveFilter = (id: string) => {
    setFilters((prev) => prev.filter((f) => f.id !== id))
    if (id.startsWith('status-')) {
      setStatusFilter(null)
    }
  }

  const handleClearFilters = () => {
    setFilters([])
    setStatusFilter(null)
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
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Support Memory Hub</h1>
            <p className="text-sm text-muted-foreground mt-1">
              AI-powered customer support with persistent long-term memory. Every interaction is remembered, every solution is learned.
            </p>
          </div>
          <Button onClick={() => setShowCreateDialog(true)}>
            <PlusIcon className="h-4 w-4 mr-2" />
            New Ticket
          </Button>
        </div>
      </motion.div>

      {/* Metrics Grid */}
      <motion.div
        variants={staggerContainer}
        className="grid grid-cols-2 lg:grid-cols-4 gap-2 sm:gap-3 flex-shrink-0"
      >
        {metrics.map((metric, index) => (
          <motion.div key={metric.title} variants={slideUp} transition={{ delay: index * 0.05 }}>
            <MetricCard {...metric} />
          </motion.div>
        ))}
      </motion.div>

      {/* Filters */}
      {filters.length > 0 && (
        <motion.div variants={slideUp} className="flex-shrink-0">
          <FilterChips filters={filters} onRemove={handleRemoveFilter} onClearAll={handleClearFilters} />
        </motion.div>
      )}

      {/* Tickets List */}
      <motion.div variants={slideUp} className="flex-1 min-h-0">
        <TicketList
          tickets={filteredTickets}
          onCreateNew={() => setShowCreateDialog(true)}
        />
      </motion.div>

      {/* Create Ticket Dialog */}
      <CreateTicketDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onSuccess={() => {
          refetch()
        }}
      />
    </motion.div>
  )
}

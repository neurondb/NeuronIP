'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import MetricCard from '@/components/dashboard/MetricCard'
import TicketList from '@/components/support/TicketList'
import FilterChips from '@/components/filters/FilterChips'
import Button from '@/components/ui/Button'
import { PlusIcon, FunnelIcon } from '@heroicons/react/24/outline'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

interface Ticket {
  id: string
  title: string
  status: 'open' | 'in-progress' | 'resolved' | 'closed'
  priority: 'low' | 'medium' | 'high'
  createdAt: Date
  updatedAt: Date
}

export default function SupportPage() {
  // Mock tickets data - replace with real API call
  const [tickets] = useState<Ticket[]>([
    {
      id: '1',
      title: 'Integration sync issue',
      status: 'open',
      priority: 'high',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 2),
      updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 2),
    },
    {
      id: '2',
      title: 'Customer memory not updating',
      status: 'in-progress',
      priority: 'medium',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24),
      updatedAt: new Date(Date.now() - 1000 * 60 * 30),
    },
    {
      id: '3',
      title: 'API rate limit question',
      status: 'resolved',
      priority: 'low',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3),
      updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 2),
    },
    {
      id: '4',
      title: 'Database connection timeout',
      status: 'open',
      priority: 'high',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 6),
      updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 6),
    },
    {
      id: '5',
      title: 'Feature request: Export to Excel',
      status: 'in-progress',
      priority: 'low',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 5),
      updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 4),
    },
    {
      id: '6',
      title: 'Authentication token expiry issue',
      status: 'open',
      priority: 'medium',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 12),
      updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 12),
    },
  ])

  const [filters, setFilters] = useState<Array<{ id: string; label: string; value: string; type: string }>>([])
  const [statusFilter, setStatusFilter] = useState<string | null>(null)

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
          <Button>
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
        <TicketList tickets={filteredTickets} />
      </motion.div>
    </motion.div>
  )
}

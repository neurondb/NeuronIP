'use client'

import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function IntegrationsPage() {
  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Integrations</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Slack, Teams, CRM, ERP, ticketing, email. Webhooks and triggers.
        </p>
      </motion.div>
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        <p className="text-muted-foreground">Integrations management coming soon...</p>
      </motion.div>
    </motion.div>
  )
}

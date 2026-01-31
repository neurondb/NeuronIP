'use client'

import { motion } from 'framer-motion'

import AuditLogs from '@/components/audit/AuditLogs'
import { getAuditExportUrl } from '@/lib/api/queries'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function AuditPage() {
  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Audit / Activity</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Full history of user actions, agent actions, and model outputs. Compliance trail for trust.
          </p>
        </div>
        <div className="flex gap-2">
          <a
            href={getAuditExportUrl({ format: 'csv' })}
            download
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium hover:bg-muted"
          >
            Export CSV
          </a>
          <a
            href={getAuditExportUrl({ format: 'json' })}
            download
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium hover:bg-muted"
          >
            Export JSON
          </a>
        </div>
      </motion.div>
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        <AuditLogs />
      </motion.div>
    </motion.div>
  )
}

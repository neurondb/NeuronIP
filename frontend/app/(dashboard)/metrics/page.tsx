'use client'

import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function MetricsPage() {
  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Metrics / Semantic Layer</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Define KPIs, dimensions, business terms. Reusable metrics for analytics and agents.
        </p>
      </motion.div>
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        <p className="text-muted-foreground">Metrics and semantic layer management coming soon...</p>
      </motion.div>
    </motion.div>
  )
}

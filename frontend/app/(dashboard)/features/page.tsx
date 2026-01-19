'use client'

import { motion } from 'framer-motion'
import FeatureGrid from '@/components/features/FeatureGrid'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function FeaturesPage() {
  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Features</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Explore all platform capabilities and features
        </p>
      </motion.div>

      {/* Feature Grid */}
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        <FeatureGrid />
      </motion.div>
    </motion.div>
  )
}

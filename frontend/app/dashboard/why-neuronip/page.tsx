'use client'

import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import WhyNeuronIP from '@/components/marketing/WhyNeuronIP'

export default function WhyNeuronIPPage() {
  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        <WhyNeuronIP />
      </motion.div>
    </motion.div>
  )
}

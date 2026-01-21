'use client'

import { motion } from 'framer-motion'
import WorkflowBuilder from '@/components/workflows/WorkflowBuilder'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function WorkflowBuilderPage() {
  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="h-full flex flex-col"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Workflow Builder</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Create and edit workflows with a visual drag-and-drop interface
        </p>
      </motion.div>
      <motion.div variants={slideUp} className="flex-1 min-h-0">
        <WorkflowBuilder />
      </motion.div>
    </motion.div>
  )
}

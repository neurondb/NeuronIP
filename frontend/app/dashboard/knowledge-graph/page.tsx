'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import GraphVisualization from '@/components/knowledge-graph/GraphVisualization'
import EntityList from '@/components/knowledge-graph/EntityList'
import EntityExtractor from '@/components/knowledge-graph/EntityExtractor'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

interface Entity {
  id: string
  name: string
  type: string
  description?: string
}

export default function KnowledgeGraphPage() {
  const [selectedEntityId, setSelectedEntityId] = useState<string | null>(null)

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Knowledge Graph</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Explore entity relationships and graph traversal
        </p>
      </motion.div>

      {/* Main Content Grid */}
      <div className="grid gap-3 sm:gap-4 lg:grid-cols-3 flex-1 min-h-0">
        {/* Graph Visualization - Takes 2 columns */}
        <motion.div variants={slideUp} className="lg:col-span-2 flex flex-col min-h-0">
          <GraphVisualization
            selectedNodeId={selectedEntityId || undefined}
            onNodeClick={setSelectedEntityId}
          />
        </motion.div>

        {/* Entity List - Takes 1 column */}
        <motion.div variants={slideUp} className="flex flex-col min-h-0">
          <EntityList onSelectEntity={(entity) => setSelectedEntityId(entity.id)} />
        </motion.div>
      </div>

      {/* Entity Extractor */}
      <motion.div variants={slideUp} className="flex-shrink-0">
        <EntityExtractor />
      </motion.div>
    </motion.div>
  )
}

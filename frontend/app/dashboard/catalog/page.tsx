'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import DatasetList from '@/components/catalog/DatasetList'
import DatasetDetail from '@/components/catalog/DatasetDetail'
import Button from '@/components/ui/Button'

export default function CatalogPage() {
  const [selectedDatasetId, setSelectedDatasetId] = useState<string | null>(null)

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Data Catalog</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Browse datasets, fields, owners, descriptions. Semantic and relational discovery.
            </p>
          </div>
          {selectedDatasetId && (
            <Button variant="outline" onClick={() => setSelectedDatasetId(null)}>
              Back to List
            </Button>
          )}
        </div>
      </motion.div>
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
        {selectedDatasetId ? (
          <DatasetDetail datasetId={selectedDatasetId} />
        ) : (
          <DatasetList onSelectDataset={setSelectedDatasetId} />
        )}
      </motion.div>
    </motion.div>
  )
}

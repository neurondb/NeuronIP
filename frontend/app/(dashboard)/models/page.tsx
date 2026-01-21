'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import ModelList from '@/components/models/ModelList'
import InferenceInterface from '@/components/models/InferenceInterface'
import RegisterModelDialog from '@/components/models/RegisterModelDialog'
import Button from '@/components/ui/Button'
import { PlusIcon } from '@heroicons/react/24/outline'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function ModelsPage() {
  const [selectedModelId, setSelectedModelId] = useState<string | null>(null)
  const [registerDialogOpen, setRegisterDialogOpen] = useState(false)

  return (
    <>
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
              <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Models</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Manage AI models and run inferences
              </p>
            </div>
            <Button onClick={() => setRegisterDialogOpen(true)}>
              <PlusIcon className="h-4 w-4 mr-2" />
              Register Model
            </Button>
          </div>
        </motion.div>

        {/* Main Content Grid */}
        <div className="grid gap-3 sm:gap-4 lg:grid-cols-2 flex-1 min-h-0">
          {/* Model List */}
          <motion.div variants={slideUp} className="flex flex-col min-h-0">
            <ModelList
              onSelectModel={setSelectedModelId}
              onCreateNew={() => setRegisterDialogOpen(true)}
            />
          </motion.div>

          {/* Inference Interface */}
          <motion.div variants={slideUp} className="flex flex-col min-h-0">
            {selectedModelId ? (
              <InferenceInterface modelId={selectedModelId} />
            ) : (
              <div className="h-full flex items-center justify-center rounded-lg border border-border bg-card">
                <p className="text-sm text-muted-foreground">Select a model to run inference</p>
              </div>
            )}
          </motion.div>
        </div>
      </motion.div>

      <RegisterModelDialog open={registerDialogOpen} onOpenChange={setRegisterDialogOpen} />
    </>
  )
}

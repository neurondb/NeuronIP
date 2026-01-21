'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import VersionList from '@/components/versioning/VersionList'
import Button from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'

export default function VersioningPage() {
  const [resourceType, setResourceType] = useState<string>('model')
  const [resourceId, setResourceId] = useState<string>('')

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Versioning</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Models, embeddings, workflows, metrics versions. Rollback and history.
        </p>
      </motion.div>
      <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto space-y-4">
        <Card>
          <CardHeader>
            <CardTitle>Select Resource</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-2">
              <select
                value={resourceType}
                onChange={(e) => setResourceType(e.target.value)}
                className="rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="model">Model</option>
                <option value="workflow">Workflow</option>
                <option value="metric">Metric</option>
                <option value="embedding">Embedding</option>
              </select>
              <input
                type="text"
                placeholder="Resource ID"
                value={resourceId}
                onChange={(e) => setResourceId(e.target.value)}
                className="flex-1 rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
          </CardContent>
        </Card>

        {resourceId ? (
          <VersionList resourceType={resourceType} resourceId={resourceId} />
        ) : (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              Enter a resource ID to view versions
            </CardContent>
          </Card>
        )}
      </motion.div>
    </motion.div>
  )
}

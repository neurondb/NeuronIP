'use client'

import { ReactNode } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { CpuChipIcon, PlayIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface ModelCardProps {
  id: string
  name: string
  type: string
  status: 'active' | 'inactive' | 'error'
  description?: string
  onInfer?: () => void
}

export default function ModelCard({ id, name, type, status, description, onInfer }: ModelCardProps) {
  const statusColors = {
    active: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    inactive: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200',
    error: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
  }

  return (
    <motion.div variants={slideUp} initial="hidden" animate="visible" transition={transition}>
      <Card hover className="h-full">
        <CardHeader className="flex-shrink-0">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <CpuChipIcon className="h-6 w-6 text-primary" />
              <div>
                <CardTitle className="text-base font-semibold">{name}</CardTitle>
                <CardDescription className="text-xs mt-0.5">{type}</CardDescription>
              </div>
            </div>
            <span className={cn('px-2 py-1 rounded-full text-xs font-medium', statusColors[status])}>
              {status}
            </span>
          </div>
        </CardHeader>
        <CardContent>
          {description && <p className="text-sm text-muted-foreground mb-3">{description}</p>}
          <Button size="sm" onClick={onInfer} className="w-full">
            <PlayIcon className="h-4 w-4 mr-2" />
            Run Inference
          </Button>
        </CardContent>
      </Card>
    </motion.div>
  )
}

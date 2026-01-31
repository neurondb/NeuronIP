'use client'

import { ExclamationTriangleIcon } from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import { ReactNode } from 'react'

import { microcopy } from '@/lib/copy/microcopy'

import Button from './Button'

interface ErrorStateProps {
  title?: string
  message?: string
  onRetry?: () => void
  retryLabel?: string
  className?: string
  icon?: ReactNode
}

export default function ErrorState({
  title = microcopy.errors.generic,
  message,
  onRetry,
  retryLabel = 'Try Again',
  className = '',
  icon,
}: ErrorStateProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className={`flex flex-col items-center justify-center py-12 px-4 text-center ${className}`}
    >
      <motion.div
        initial={{ scale: 0.8, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ delay: 0.1 }}
        className="mb-4 text-destructive"
      >
        {icon || <ExclamationTriangleIcon className="h-12 w-12" />}
      </motion.div>
      <h3 className="text-lg font-semibold mb-2">{title}</h3>
      {message && (
        <p className="text-sm text-muted-foreground max-w-md mb-6">{message}</p>
      )}
      {onRetry && (
        <Button onClick={onRetry} variant="outline" size="lg">
          {retryLabel}
        </Button>
      )}
    </motion.div>
  )
}

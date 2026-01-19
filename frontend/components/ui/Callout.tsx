'use client'

import { useState, ReactNode } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  XMarkIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  InformationCircleIcon,
  LightBulbIcon,
  DocumentTextIcon,
} from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'

export type CalloutType = 'tip' | 'warning' | 'info' | 'success' | 'error' | 'note'

interface CalloutProps {
  type: CalloutType
  title?: string
  icon?: React.ComponentType<{ className?: string }>
  dismissible?: boolean
  onDismiss?: () => void
  children: ReactNode
  className?: string
}

export default function Callout({
  type,
  title,
  icon: CustomIcon,
  dismissible = false,
  onDismiss,
  children,
  className,
}: CalloutProps) {
  const [isDismissed, setIsDismissed] = useState(false)

  const icons = {
    success: CheckCircleIcon,
    error: ExclamationCircleIcon,
    info: InformationCircleIcon,
    warning: ExclamationCircleIcon,
    tip: LightBulbIcon,
    note: DocumentTextIcon,
  }

  const colors = {
    success: 'bg-green-50 text-green-800 border-green-200 dark:bg-green-900/20 dark:text-green-400 dark:border-green-800/50',
    error: 'bg-red-50 text-red-800 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800/50',
    info: 'bg-blue-50 text-blue-800 border-blue-200 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-800/50',
    warning: 'bg-yellow-50 text-yellow-800 border-yellow-200 dark:bg-yellow-900/20 dark:text-yellow-400 dark:border-yellow-800/50',
    tip: 'bg-cyan-50 text-cyan-800 border-cyan-200 dark:bg-cyan-900/20 dark:text-cyan-400 dark:border-cyan-800/50',
    note: 'bg-gray-50 text-gray-800 border-gray-200 dark:bg-gray-900/20 dark:text-gray-400 dark:border-gray-800/50',
  }

  const DefaultIcon = icons[type]
  const Icon = CustomIcon || DefaultIcon

  const handleDismiss = () => {
    setIsDismissed(true)
    onDismiss?.()
  }

  if (isDismissed) return null

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: -10, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        exit={{ opacity: 0, y: -10, scale: 0.98 }}
        transition={{ duration: 0.2 }}
        className={cn(
          'rounded-lg border p-4 shadow-sm',
          colors[type],
          className
        )}
      >
        <div className="flex items-start gap-3">
          <Icon className="h-5 w-5 flex-shrink-0 mt-0.5" />
          <div className="flex-1 min-w-0">
            {title && (
              <h4 className="text-sm font-semibold mb-1.5">{title}</h4>
            )}
            <div className="text-sm [&>p]:my-0 [&>p:last-child]:mb-0 [&>ul]:my-0 [&>ol]:my-0">
              {children}
            </div>
          </div>
          {dismissible && (
            <button
              onClick={handleDismiss}
              className="flex-shrink-0 rounded opacity-70 hover:opacity-100 transition-opacity focus:outline-none focus:ring-2 focus:ring-offset-1 focus:ring-current"
              aria-label="Dismiss"
            >
              <XMarkIcon className="h-4 w-4" />
            </button>
          )}
        </div>
      </motion.div>
    </AnimatePresence>
  )
}

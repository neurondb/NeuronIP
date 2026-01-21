'use client'

import { ReactNode } from 'react'
import { motion } from 'framer-motion'
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import Callout from './Callout'

interface WarningProps {
  title?: string
  message: ReactNode
  severity?: 'low' | 'medium' | 'high'
  dismissible?: boolean
  onDismiss?: () => void
  className?: string
  action?: ReactNode
}

export default function Warning({
  title,
  message,
  severity = 'medium',
  dismissible = false,
  onDismiss,
  className,
  action,
}: WarningProps) {
  const severityConfig = {
    low: {
      type: 'tip' as const,
      icon: ExclamationTriangleIcon,
    },
    medium: {
      type: 'warning' as const,
      icon: ExclamationTriangleIcon,
    },
    high: {
      type: 'error' as const,
      icon: ExclamationTriangleIcon,
    },
  }

  const config = severityConfig[severity]

  return (
    <motion.div
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      className={cn('w-full', className)}
    >
      <Callout
        type={config.type}
        title={title || 'Warning'}
        icon={config.icon}
        dismissible={dismissible}
        onDismiss={onDismiss}
      >
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 min-w-0">{message}</div>
          {action && <div className="flex-shrink-0">{action}</div>}
        </div>
      </Callout>
    </motion.div>
  )
}

'use client'

import { ReactNode } from 'react'
import { motion } from 'framer-motion'
import {
  CheckCircleIcon,
  XCircleIcon,
  ClockIcon,
  ExclamationCircleIcon,
} from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import Tooltip from './Tooltip'

export type StatusType = 'connected' | 'disconnected' | 'connecting' | 'error' | 'warning'

interface StatusBadgeProps {
  status: StatusType
  label: string
  details?: ReactNode
  size?: 'sm' | 'md' | 'lg'
  showIcon?: boolean
  className?: string
  onClick?: () => void
}

const statusConfig: Record<
  StatusType,
  {
    color: string
    icon: typeof CheckCircleIcon
    textColor: string
  }
> = {
  connected: {
    color: 'bg-green-500 dark:bg-green-400',
    icon: CheckCircleIcon,
    textColor: 'text-green-600 dark:text-green-400',
  },
  disconnected: {
    color: 'bg-red-500 dark:bg-red-400',
    icon: XCircleIcon,
    textColor: 'text-red-600 dark:text-red-400',
  },
  connecting: {
    color: 'bg-yellow-500 dark:bg-yellow-400',
    icon: ClockIcon,
    textColor: 'text-yellow-600 dark:text-yellow-400',
  },
  error: {
    color: 'bg-red-500 dark:bg-red-400',
    icon: XCircleIcon,
    textColor: 'text-red-600 dark:text-red-400',
  },
  warning: {
    color: 'bg-yellow-500 dark:bg-yellow-400',
    icon: ExclamationCircleIcon,
    textColor: 'text-yellow-600 dark:text-yellow-400',
  },
}

const sizes = {
  sm: {
    container: 'px-2 py-1 text-xs',
    icon: 'h-3 w-3',
    dot: 'h-2 w-2',
  },
  md: {
    container: 'px-3 py-1.5 text-sm',
    icon: 'h-4 w-4',
    dot: 'h-2.5 w-2.5',
  },
  lg: {
    container: 'px-4 py-2 text-base',
    icon: 'h-5 w-5',
    dot: 'h-3 w-3',
  },
}

export default function StatusBadge({
  status,
  label,
  details,
  size = 'md',
  showIcon = true,
  className,
  onClick,
}: StatusBadgeProps) {
  const config = statusConfig[status]
  const sizeConfig = sizes[size]
  const Icon = config.icon

  const badgeContent = (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      className={cn(
        'inline-flex items-center gap-2 rounded-full font-medium transition-colors',
        config.textColor,
        sizeConfig.container,
        onClick && 'cursor-pointer hover:opacity-80',
        className
      )}
      onClick={onClick}
    >
      {showIcon && status !== 'connecting' && (
        <Icon className={cn(sizeConfig.icon, 'flex-shrink-0')} />
      )}
      {status === 'connecting' && (
        <div
          className={cn(
            'flex-shrink-0 rounded-full border-2 border-current border-t-transparent animate-spin',
            sizeConfig.icon
          )}
        />
      )}
      {!showIcon && (
        <span
          className={cn(
            'flex-shrink-0 rounded-full',
            config.color,
            sizeConfig.dot
          )}
        />
      )}
      <span>{label}</span>
    </motion.div>
  )

  if (details) {
    return (
      <Tooltip content={details} variant="info" side="top">
        {badgeContent}
      </Tooltip>
    )
  }

  return badgeContent
}

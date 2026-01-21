'use client'

import { ReactNode } from 'react'
import * as TooltipPrimitive from '@radix-ui/react-tooltip'
import { cn } from '@/lib/utils/cn'
import { InformationCircleIcon, ExclamationCircleIcon, LightBulbIcon } from '@heroicons/react/24/outline'

export type TooltipVariant = 'info' | 'warning' | 'help'

interface TooltipProps {
  content: ReactNode
  children: React.ReactNode
  variant?: TooltipVariant
  side?: 'top' | 'right' | 'bottom' | 'left'
  align?: 'start' | 'center' | 'end'
  delayDuration?: number
  className?: string
  disabled?: boolean
}

const variantIcons = {
  info: InformationCircleIcon,
  warning: ExclamationCircleIcon,
  help: LightBulbIcon,
}

const variantColors = {
  info: 'bg-blue-50 text-blue-900 border-blue-200 dark:bg-blue-900/30 dark:text-blue-100 dark:border-blue-800',
  warning: 'bg-yellow-50 text-yellow-900 border-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-100 dark:border-yellow-800',
  help: 'bg-cyan-50 text-cyan-900 border-cyan-200 dark:bg-cyan-900/30 dark:text-cyan-100 dark:border-cyan-800',
}

export default function Tooltip({
  content,
  children,
  variant = 'info',
  side = 'top',
  align = 'center',
  delayDuration = 300,
  className,
  disabled = false,
}: TooltipProps) {
  if (disabled) {
    return <>{children}</>
  }

  const Icon = variantIcons[variant]
  const variantColor = variantColors[variant]

  return (
    <TooltipPrimitive.Provider delayDuration={delayDuration}>
      <TooltipPrimitive.Root>
        <TooltipPrimitive.Trigger asChild>{children}</TooltipPrimitive.Trigger>
        <TooltipPrimitive.Portal>
          <TooltipPrimitive.Content
            side={side}
            align={align}
            sideOffset={8}
            className={cn(
              'z-50 max-w-xs rounded-lg border px-3 py-2 text-sm shadow-lg',
              'animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95',
              variantColor,
              className
            )}
          >
            <div className="flex items-start gap-2">
              <Icon className="h-4 w-4 flex-shrink-0 mt-0.5" />
              <div className="flex-1 min-w-0">{content}</div>
            </div>
            <TooltipPrimitive.Arrow className="fill-current" />
          </TooltipPrimitive.Content>
        </TooltipPrimitive.Portal>
      </TooltipPrimitive.Root>
    </TooltipPrimitive.Provider>
  )
}

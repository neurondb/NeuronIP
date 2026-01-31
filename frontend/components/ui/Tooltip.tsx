'use client'

import { InformationCircleIcon, ExclamationCircleIcon, LightBulbIcon } from '@heroicons/react/24/outline'
import * as TooltipPrimitive from '@radix-ui/react-tooltip'
import { ReactNode } from 'react'

import { cn } from '@/lib/utils/cn'


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
  info: 'bg-info/10 text-info border-info/30 dark:bg-info/20 dark:border-info/40',
  warning: 'bg-warning/10 text-warning border-warning/30 dark:bg-warning/20 dark:border-warning/40',
  help: 'bg-muted text-muted-foreground border-border',
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

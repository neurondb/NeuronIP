'use client'

import { ReactNode } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface MetricCardProps {
  title: string
  value: string | number
  description?: string
  icon?: ReactNode
  trend?: {
    value: number
    isPositive: boolean
  }
  className?: string
}

export default function MetricCard({
  title,
  value,
  description,
  icon,
  trend,
  className,
}: MetricCardProps) {
  return (
    <motion.div
      variants={slideUp}
      initial="hidden"
      animate="visible"
      transition={transition}
    >
      <Card hover className={cn('h-full', className)}>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-1.5 px-4 pt-4">
          <CardTitle className="text-xs sm:text-sm font-medium leading-tight">{title}</CardTitle>
          {icon && <div className="h-4 w-4 sm:h-5 sm:w-5 text-muted-foreground flex-shrink-0">{icon}</div>}
        </CardHeader>
        <CardContent className="px-4 pb-4 pt-1">
          <div className="text-xl sm:text-2xl font-bold leading-tight">{value}</div>
          {description && (
            <p className="text-xs text-muted-foreground mt-1.5 leading-tight">{description}</p>
          )}
          {trend && (
            <div className={cn(
              'text-xs mt-1.5 leading-tight',
              trend.isPositive ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
            )}>
              {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}% from last period
            </div>
          )}
        </CardContent>
      </Card>
    </motion.div>
  )
}

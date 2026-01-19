'use client'

import { ReactNode } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { cn } from '@/lib/utils/cn'

interface ChartContainerProps {
  title?: string
  description?: string
  children: ReactNode
  className?: string
}

export default function ChartContainer({
  title,
  description,
  children,
  className,
}: ChartContainerProps) {
  return (
    <Card className={cn('h-full flex flex-col', className)}>
      {(title || description) && (
        <CardHeader className="flex-shrink-0">
          {title && <CardTitle className="text-base sm:text-lg">{title}</CardTitle>}
          {description && <CardDescription className="text-xs">{description}</CardDescription>}
        </CardHeader>
      )}
      <CardContent className="flex-1 min-h-0">{children}</CardContent>
    </Card>
  )
}

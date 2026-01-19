'use client'

import { HTMLAttributes, forwardRef } from 'react'
import { motion, HTMLMotionProps } from 'framer-motion'
import { cn } from '@/lib/utils/cn'

interface CardProps extends HTMLMotionProps<'div'> {
  hover?: boolean
  children: React.ReactNode
}

const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ hover = false, className, children, ...props }, ref) => {
    return (
      <motion.div
        ref={ref}
        className={cn(
          'rounded-lg border border-border bg-card text-card-foreground shadow-sm',
          hover && 'transition-shadow duration-200 hover:shadow-md cursor-pointer',
          className
        )}
        whileHover={hover ? { y: -2, transition: { duration: 0.2 } } : undefined}
        {...(props as any)}
      >
        {children}
      </motion.div>
    )
  }
)

Card.displayName = 'Card'

const CardHeader = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('flex flex-col space-y-1.5 p-4 sm:p-5', className)}
      {...props}
    />
  )
)
CardHeader.displayName = 'CardHeader'

const CardTitle = forwardRef<HTMLParagraphElement, HTMLAttributes<HTMLHeadingElement>>(
  ({ className, ...props }, ref) => (
    <h3
      ref={ref}
      className={cn('text-2xl font-semibold leading-none tracking-tight', className)}
      {...props}
    />
  )
)
CardTitle.displayName = 'CardTitle'

const CardDescription = forwardRef<HTMLParagraphElement, HTMLAttributes<HTMLParagraphElement>>(
  ({ className, ...props }, ref) => (
    <p ref={ref} className={cn('text-sm text-muted-foreground', className)} {...props} />
  )
)
CardDescription.displayName = 'CardDescription'

const CardContent = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn('p-4 sm:p-5 pt-0', className)} {...props} />
  )
)
CardContent.displayName = 'CardContent'

const CardFooter = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn('flex items-center p-4 sm:p-5 pt-0', className)} {...props} />
  )
)
CardFooter.displayName = 'CardFooter'

export { Card, CardHeader, CardFooter, CardTitle, CardDescription, CardContent }

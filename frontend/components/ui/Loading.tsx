'use client'

import { HTMLAttributes } from 'react'
import { motion } from 'framer-motion'
import { cn } from '@/lib/utils/cn'

interface LoadingProps extends HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg'
  variant?: 'spinner' | 'dots' | 'pulse'
}

export default function Loading({ size = 'md', variant = 'spinner', className, ...props }: LoadingProps) {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-8 w-8',
    lg: 'h-12 w-12',
  }

  if (variant === 'spinner') {
    return (
      <div className={cn('flex items-center justify-center', className)} {...props}>
        <motion.div
          className={cn('rounded-full border-2 border-primary border-t-transparent', sizeClasses[size])}
          animate={{ rotate: 360 }}
          transition={{ duration: 0.8, repeat: Infinity, ease: 'linear' }}
        />
      </div>
    )
  }

  if (variant === 'dots') {
    return (
      <div className={cn('flex items-center justify-center gap-1', className)} {...props}>
        {[0, 1, 2].map((i) => (
          <motion.div
            key={i}
            className={cn('rounded-full bg-primary', size === 'sm' ? 'h-1.5 w-1.5' : size === 'md' ? 'h-2 w-2' : 'h-3 w-3')}
            animate={{
              y: [0, -8, 0],
            }}
            transition={{
              duration: 0.6,
              repeat: Infinity,
              delay: i * 0.1,
              ease: 'easeInOut',
            }}
          />
        ))}
      </div>
    )
  }

  return (
    <div className={cn('flex items-center justify-center', className)} {...props}>
      <motion.div
        className={cn('rounded-full bg-primary', sizeClasses[size])}
        animate={{
          scale: [1, 1.2, 1],
          opacity: [1, 0.7, 1],
        }}
        transition={{
          duration: 1,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />
    </div>
  )
}

// Skeleton loader component
interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'text' | 'circular' | 'rectangular'
}

export function Skeleton({ variant = 'rectangular', className, ...props }: SkeletonProps) {
  const baseClasses = 'animate-shimmer bg-muted rounded'
  const variantClasses = {
    text: 'h-4',
    circular: 'rounded-full',
    rectangular: 'h-20',
  }

  return (
    <div
      className={cn(baseClasses, variantClasses[variant], className)}
      {...props}
    />
  )
}

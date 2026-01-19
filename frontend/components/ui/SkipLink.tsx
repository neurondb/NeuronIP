'use client'

import Link from 'next/link'
import { cn } from '@/lib/utils/cn'

interface SkipLinkProps {
  href?: string
  className?: string
}

export default function SkipLink({ href = '#main-content', className }: SkipLinkProps) {
  return (
    <Link
      href={href}
      className={cn(
        'absolute -top-full left-4 z-[100] rounded-b-lg bg-primary px-4 py-2 text-primary-foreground',
        'focus:top-0 transition-all duration-200',
        className
      )}
    >
      Skip to main content
    </Link>
  )
}

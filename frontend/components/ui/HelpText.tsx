'use client'

import { ReactNode, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { InformationCircleIcon, XMarkIcon, DocumentTextIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import Tooltip from './Tooltip'

interface HelpTextProps {
  content: ReactNode
  title?: string
  variant?: 'inline' | 'tooltip' | 'popover'
  link?: string
  linkText?: string
  className?: string
}

export default function HelpText({
  content,
  title,
  variant = 'inline',
  link,
  linkText = 'Learn more',
  className,
}: HelpTextProps) {
  const [isOpen, setIsOpen] = useState(false)

  if (variant === 'tooltip') {
    return (
      <Tooltip content={content} variant="help" className={className}>
        <button
          type="button"
          className="inline-flex items-center text-muted-foreground hover:text-foreground transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 rounded"
          aria-label="Help"
        >
          <InformationCircleIcon className="h-4 w-4" />
        </button>
      </Tooltip>
    )
  }

  if (variant === 'popover') {
    return (
      <div className={cn('relative inline-flex', className)}>
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="inline-flex items-center text-muted-foreground hover:text-foreground transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 rounded"
          aria-label="Help"
          aria-expanded={isOpen}
        >
          <InformationCircleIcon className="h-4 w-4" />
        </button>
        <AnimatePresence>
          {isOpen && (
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: -10 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: -10 }}
              transition={{ duration: 0.2 }}
              className="absolute bottom-full left-0 mb-2 w-80 rounded-lg border bg-card p-4 shadow-lg z-50"
            >
              <div className="flex items-start justify-between gap-2 mb-2">
                <div className="flex items-center gap-2">
                  <DocumentTextIcon className="h-4 w-4 text-primary" />
                  {title && <h4 className="text-sm font-semibold">{title}</h4>}
                </div>
                <button
                  type="button"
                  onClick={() => setIsOpen(false)}
                  className="text-muted-foreground hover:text-foreground transition-colors"
                  aria-label="Close"
                >
                  <XMarkIcon className="h-4 w-4" />
                </button>
              </div>
              <div className="text-sm text-muted-foreground [&>p]:mb-2 [&>p:last-child]:mb-0">
                {content}
              </div>
              {link && (
                <a
                  href={link}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="mt-3 inline-flex items-center text-sm text-primary hover:underline"
                >
                  {linkText} →
                </a>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    )
  }

  // Inline variant
  return (
    <div className={cn('flex items-start gap-2 text-sm text-muted-foreground', className)}>
      <InformationCircleIcon className="h-4 w-4 flex-shrink-0 mt-0.5" />
      <div className="flex-1 min-w-0">
        {title && <h4 className="font-medium text-foreground mb-1">{title}</h4>}
        <div className="[&>p]:mb-1 [&>p:last-child]:mb-0">{content}</div>
        {link && (
          <a
            href={link}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-2 inline-flex items-center text-primary hover:underline"
          >
            {linkText} →
          </a>
        )}
      </div>
    </div>
  )
}

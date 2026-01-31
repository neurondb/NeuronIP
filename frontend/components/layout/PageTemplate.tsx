'use client'

import { motion } from 'framer-motion'
import { ReactNode } from 'react'

import type { PageArchetype } from '@/lib/pageArchetypes'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import { cn } from '@/lib/utils/cn'

import {
  PageStateEmpty,
  PageStateError,
  PageStateLoading,
} from './PageStates'

export type PageTemplateStatus = 'idle' | 'loading' | 'empty' | 'error'

export interface PageTemplateProps {
  /** Page title (e.g. "Dashboard", "Users") */
  title: string
  /** Optional short description below title */
  description?: string
  /** Optional primary actions (e.g. "Add User" button) - rendered in header row */
  actions?: ReactNode
  /** Optional filter/search row below header */
  filterRow?: ReactNode
  /** Main content. Ignored when status is loading/empty/error unless children are always shown. */
  children: ReactNode
  /** Archetype for consistent empty/loading/error states */
  archetype?: PageArchetype
  /** When 'loading' | 'empty' | 'error', shows archetype-appropriate state instead of children */
  status?: PageTemplateStatus
  /** Empty state overrides (when status === 'empty') */
  emptyTitle?: string
  emptyDescription?: string
  emptyAction?: { label: string; onClick: () => void }
  /** Error state overrides (when status === 'error') */
  errorTitle?: string
  errorMessage?: string
  onRetry?: () => void
  /** Extra class for the outer wrapper */
  className?: string
  /** If true, skip motion (e.g. for static exports) */
  noMotion?: boolean
}

export default function PageTemplate({
  title,
  description,
  actions,
  filterRow,
  children,
  archetype = 'list-detail',
  status = 'idle',
  emptyTitle,
  emptyDescription,
  emptyAction,
  errorTitle,
  errorMessage,
  onRetry,
  className,
  noMotion,
}: PageTemplateProps) {
  const Wrapper = noMotion ? 'div' : motion.div
  const wrapperProps = noMotion
    ? { className: cn('space-y-4 sm:space-y-6 flex flex-col h-full', className) }
    : {
        variants: staggerContainer,
        initial: 'hidden' as const,
        animate: 'visible' as const,
        className: cn('space-y-4 sm:space-y-6 flex flex-col h-full', className),
      }

  return (
    <Wrapper {...wrapperProps}>
      {/* Header: title, description, actions */}
      {noMotion ? (
        <div className="flex-shrink-0 pb-2">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h1 className="text-2xl sm:text-3xl font-bold text-foreground">{title}</h1>
              {description && (
                <p className="text-sm sm:text-base text-muted-foreground mt-1">{description}</p>
              )}
            </div>
            {actions && <div className="flex items-center gap-2 flex-shrink-0">{actions}</div>}
          </div>
        </div>
      ) : (
        <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h1 className="text-2xl sm:text-3xl font-bold text-foreground">{title}</h1>
              {description && (
                <p className="text-sm sm:text-base text-muted-foreground mt-1">{description}</p>
              )}
            </div>
            {actions && <div className="flex items-center gap-2 flex-shrink-0">{actions}</div>}
          </div>
        </motion.div>
      )}

      {/* Optional filter row */}
      {filterRow && (
        noMotion ? (
          <div className="flex-shrink-0">{filterRow}</div>
        ) : (
          <motion.div variants={slideUp} className="flex-shrink-0">
            {filterRow}
          </motion.div>
        )
      )}

      {/* Content or state */}
      {status === 'loading' && (
        <PageStateLoading archetype={archetype} className="flex-1 min-h-0" />
      )}
      {status === 'empty' && (
        <PageStateEmpty
          archetype={archetype}
          title={emptyTitle}
          description={emptyDescription}
          action={emptyAction}
          className="flex-1 min-h-0"
        />
      )}
      {status === 'error' && (
        <PageStateError
          archetype={archetype}
          title={errorTitle}
          message={errorMessage}
          onRetry={onRetry}
          className="flex-1 min-h-0"
        />
      )}
      {status === 'idle' && (
        noMotion ? (
          <div className="flex-1 min-h-0 flex flex-col">{children}</div>
        ) : (
          <motion.div variants={slideUp} className="flex-1 min-h-0 flex flex-col">
            {children}
          </motion.div>
        )
      )}
    </Wrapper>
  )
}

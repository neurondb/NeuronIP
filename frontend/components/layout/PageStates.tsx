'use client'

import {
  BeakerIcon,
  CubeIcon,
  MagnifyingGlassIcon,
  UserGroupIcon,
} from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import { ReactNode } from 'react'

import EmptyState from '@/components/ui/EmptyState'
import ErrorState from '@/components/ui/ErrorState'
import Loading from '@/components/ui/Loading'
import type { PageArchetype } from '@/lib/pageArchetypes'
import { slideUp } from '@/lib/animations/variants'

const ARCHETYPE_EMPTY: Record<PageArchetype, { title: string; description: string; icon: ReactNode }> = {
  dashboard: {
    title: 'No data yet',
    description: 'Run a search or query to see activity here.',
    icon: <CubeIcon className="h-12 w-12 text-muted-foreground" />,
  },
  search: {
    title: 'Nothing to show',
    description: 'Try a search or explore the catalog to get started.',
    icon: <MagnifyingGlassIcon className="h-12 w-12 text-muted-foreground" />,
  },
  builder: {
    title: 'Start from scratch',
    description: 'Add nodes or blocks to build your workflow or document.',
    icon: <BeakerIcon className="h-12 w-12 text-muted-foreground" />,
  },
  'list-detail': {
    title: 'No items yet',
    description: 'Create your first item to get started.',
    icon: <UserGroupIcon className="h-12 w-12 text-muted-foreground" />,
  },
}

export interface PageStateLoadingProps {
  archetype: PageArchetype
  className?: string
}

export function PageStateLoading({ archetype: _archetype, className }: PageStateLoadingProps) {
  return (
    <motion.div
      variants={slideUp}
      initial="hidden"
      animate="visible"
      className={`flex min-h-[280px] flex-col items-center justify-center gap-4 ${className ?? ''}`}
    >
      <Loading size="lg" variant="spinner" />
      <p className="text-sm text-muted-foreground">Loading…</p>
    </motion.div>
  )
}

export interface PageStateEmptyProps {
  archetype: PageArchetype
  title?: string
  description?: string
  action?: { label: string; onClick: () => void }
  className?: string
}

export function PageStateEmpty({
  archetype,
  title,
  description,
  action,
  className,
}: PageStateEmptyProps) {
  const preset = ARCHETYPE_EMPTY[archetype]
  return (
    <motion.div
      variants={slideUp}
      initial="hidden"
      animate="visible"
      className={className ?? ''}
    >
      <EmptyState
        icon={preset.icon}
        title={title ?? preset.title}
        description={description ?? preset.description}
        action={action}
        className="min-h-[280px]"
      />
    </motion.div>
  )
}

export interface PageStateErrorProps {
  archetype: PageArchetype
  title?: string
  message?: string
  onRetry?: () => void
  className?: string
}

export function PageStateError({
  archetype: _archetype,
  title,
  message,
  onRetry,
  className,
}: PageStateErrorProps) {
  return (
    <motion.div
      variants={slideUp}
      initial="hidden"
      animate="visible"
      className={className ?? ''}
    >
      <ErrorState
        title={title}
        message={message}
        onRetry={onRetry}
        retryLabel="Try again"
        className="min-h-[280px]"
      />
    </motion.div>
  )
}

export interface PageStateSkeletonProps {
  archetype: PageArchetype
  className?: string
}

/** Skeleton matching archetype layout (e.g. list rows vs cards). */
export function PageStateSkeleton({ archetype, className }: PageStateSkeletonProps) {
  const SkeletonBar = () => (
    <div className="h-4 rounded bg-muted animate-pulse" style={{ width: `${60 + Math.random() * 30}%` }} />
  )
  if (archetype === 'dashboard') {
    return (
      <div className={`space-y-4 ${className ?? ''}`}>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-24 rounded-lg bg-muted animate-pulse" />
          ))}
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <div className="h-64 rounded-lg bg-muted animate-pulse lg:col-span-2" />
          <div className="h-64 rounded-lg bg-muted animate-pulse" />
        </div>
      </div>
    )
  }
  if (archetype === 'search' || archetype === 'builder') {
    return (
      <div className={`space-y-3 ${className ?? ''}`}>
        <SkeletonBar />
        <SkeletonBar />
        <SkeletonBar />
        <div className="h-48 rounded-lg bg-muted animate-pulse" />
      </div>
    )
  }
  // list-detail
  return (
    <div className={`space-y-2 ${className ?? ''}`}>
      {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
        <div key={i} className="flex items-center gap-3 py-3">
          <div className="h-10 w-10 rounded-full bg-muted animate-pulse" />
          <div className="flex-1 space-y-1">
            <SkeletonBar />
            <SkeletonBar />
          </div>
        </div>
      ))}
    </div>
  )
}

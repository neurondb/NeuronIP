'use client'

import dynamic from 'next/dynamic'

import Loading from '@/components/ui/Loading'

/** Lazy-loaded BlockEditor (TipTap) to reduce initial bundle. */
export const LazyBlockEditor = dynamic(
  () => import('./BlockEditor').then((mod) => ({ default: mod.BlockEditor })),
  {
    ssr: false,
    loading: () => (
      <div className="flex min-h-[200px] items-center justify-center rounded-lg border border-border bg-muted/20">
        <Loading size="lg" variant="spinner" />
      </div>
    ),
  }
)

'use client'

import dynamic from 'next/dynamic'

import Loading from '@/components/ui/Loading'

/* Lazy load Monaco editor to reduce initial bundle size */
export const LazyCodeEditor = dynamic(() => import('./CodeEditor').then(mod => ({ default: mod.CodeEditor })), {
  loading: () => <Loading size="md" variant="spinner" />,
  ssr: false,
})

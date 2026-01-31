'use client'

import dynamic from 'next/dynamic'

import Loading from '@/components/ui/Loading'

/* Lazy load semantic search components to reduce initial bundle size */
export const LazyChatInterface = dynamic(() => import('./ChatInterface'), {
  loading: () => <Loading size="lg" variant="spinner" />,
  ssr: false,
})

export const LazyDocumentList = dynamic(() => import('./DocumentList'), {
  loading: () => <Loading size="md" variant="spinner" />,
  ssr: false,
})

export const LazySearchResults = dynamic(() => import('./SearchResults'), {
  loading: () => <Loading size="md" variant="spinner" />,
  ssr: false,
})

export const LazyCollectionSetupWizard = dynamic(() => import('./CollectionSetupWizard'), {
  loading: () => <Loading size="lg" variant="spinner" />,
  ssr: false,
})

export const LazySemanticSearchSetupWizard = dynamic(() => import('./SemanticSearchSetupWizard'), {
  loading: () => <Loading size="lg" variant="spinner" />,
  ssr: false,
})

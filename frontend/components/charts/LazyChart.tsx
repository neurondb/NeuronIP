'use client'

import dynamic from 'next/dynamic'

import Loading from '@/components/ui/Loading'

/* Lazy load chart components to reduce initial bundle size */
export const LazyBarChart = dynamic(() => import('./BarChart'), {
  loading: () => <Loading size="md" variant="spinner" />,
  ssr: false,
})

export const LazyLineChart = dynamic(() => import('./LineChart'), {
  loading: () => <Loading size="md" variant="spinner" />,
  ssr: false,
})

export const LazyPieChart = dynamic(() => import('./PieChart'), {
  loading: () => <Loading size="md" variant="spinner" />,
  ssr: false,
})

export const LazyHeatmap = dynamic(() => import('./Heatmap'), {
  loading: () => <Loading size="md" variant="spinner" />,
  ssr: false,
})

export const LazyRadarChart = dynamic(() => import('./RadarChart'), {
  loading: () => <Loading size="md" variant="spinner" />,
  ssr: false,
})

export const LazyInteractiveChart = dynamic(
  () => import('./InteractiveChart').then(mod => ({ default: mod.InteractiveChart })),
  {
    loading: () => <Loading size="md" variant="spinner" />,
    ssr: false,
  }
)

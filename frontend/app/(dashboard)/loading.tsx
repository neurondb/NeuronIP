'use client'

import Loading from '@/components/ui/Loading'

export default function DashboardLoading() {
  return (
    <div className="flex items-center justify-center h-full min-h-[400px]">
      <Loading size="lg" variant="spinner" />
    </div>
  )
}

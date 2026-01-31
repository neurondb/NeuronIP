'use client'

import { usePathname } from 'next/navigation'
import { useEffect } from 'react'

import { useAppStore } from '@/lib/store/useAppStore'

/** Records current path into sidebar recents when it changes. Mount inside dashboard layout. */
export default function RecentPathTracker() {
  const pathname = usePathname()
  const addRecentPath = useAppStore((s) => s.addRecentPath)

  useEffect(() => {
    if (pathname && pathname.startsWith('/') && pathname !== '/login') {
      addRecentPath(pathname)
    }
  }, [pathname, addRecentPath])

  return null
}

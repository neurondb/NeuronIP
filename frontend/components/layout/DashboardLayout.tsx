'use client'

import { ReactNode, useEffect } from 'react'
import Sidebar from './Sidebar'
import Header from './Header'
import ShortcutsModal from './ShortcutsModal'
import SkipLink from '@/components/ui/SkipLink'
import { ToastContainer } from '@/components/ui/Toast'
import ErrorBoundary from '@/components/ui/ErrorBoundary'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useAppStore } from '@/lib/store/useAppStore'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
})

interface DashboardLayoutProps {
  children: ReactNode
}

export default function DashboardLayout({ children }: DashboardLayoutProps) {
  const { theme } = useAppStore()

  useEffect(() => {
    // Apply theme class to document based on store
    const root = document.documentElement
    if (theme === 'dark') {
      root.classList.add('dark')
    } else if (theme === 'light') {
      root.classList.remove('dark')
    } else {
      // System preference
      if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        root.classList.add('dark')
      } else {
        root.classList.remove('dark')
      }
    }
  }, [theme])

  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <SkipLink />
        <div className="flex h-screen overflow-hidden bg-background">
          <Sidebar />
          <div className="flex flex-1 flex-col overflow-hidden">
            <Header />
            <main id="main-content" className="flex-1 overflow-y-auto p-3 sm:p-4 lg:p-5 xl:p-6" tabIndex={-1}>
              <div className="max-w-[1920px] mx-auto h-full flex flex-col">
                {children}
              </div>
            </main>
          </div>
        </div>
        <ToastContainer />
        <ShortcutsModal />
      </QueryClientProvider>
    </ErrorBoundary>
  )
}

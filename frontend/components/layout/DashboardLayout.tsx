'use client'

import { ReactNode, useEffect, Suspense } from 'react'

import ErrorBoundary from '@/components/ui/ErrorBoundary'
import Loading from '@/components/ui/Loading'
import SkipLink from '@/components/ui/SkipLink'
import { ToastContainer } from '@/components/ui/Toast'
import { useAppStore } from '@/lib/store/useAppStore'

import CommandPaletteK from './CommandPaletteK'
import DashboardFooter from './DashboardFooter'
import Header from './Header'
import RecentPathTracker from './RecentPathTracker'
import ShortcutsModal from './ShortcutsModal'
import Sidebar from './Sidebar'
import StatusBar from './StatusBar'

interface DashboardLayoutProps {
  children: ReactNode
}

export default function DashboardLayout({ children }: DashboardLayoutProps) {
  const { theme } = useAppStore()

  useEffect(() => {
    const root = document.documentElement
    if (theme === 'dark') {
      root.classList.add('dark')
    } else if (theme === 'light') {
      root.classList.remove('dark')
    } else {
      if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        root.classList.add('dark')
      } else {
        root.classList.remove('dark')
      }
    }
  }, [theme])

  return (
    <ErrorBoundary>
      <RecentPathTracker />
      <SkipLink />
      <div className="flex h-screen overflow-hidden bg-background">
        <Sidebar />
        <div className="flex flex-1 flex-col overflow-hidden relative">
          <Header />
          <main id="main-content" className="flex-1 overflow-y-auto p-3 sm:p-4 lg:p-5 xl:p-6" style={{ paddingBottom: 'calc(2rem + 32px)' }} tabIndex={-1}>
            <div className="max-w-[1920px] mx-auto h-full flex flex-col">
              <Suspense fallback={
                <div className="flex items-center justify-center h-full min-h-[400px]">
                  <Loading size="lg" variant="spinner" />
                </div>
              }>
                {children}
                <div className="pb-8">
                  <DashboardFooter version={process.env.NEXT_PUBLIC_APP_VERSION || '1.0.0'} showStats={false} />
                </div>
              </Suspense>
            </div>
          </main>
          <StatusBar version={process.env.NEXT_PUBLIC_APP_VERSION || '1.0.0'} />
        </div>
      </div>
      <ToastContainer />
      <CommandPaletteK />
      <ShortcutsModal />
    </ErrorBoundary>
  )
}

'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import {
  UserCircleIcon,
  SunIcon,
  MoonIcon,
  Cog6ToothIcon,
  ArrowRightOnRectangleIcon,
  UserIcon,
} from '@heroicons/react/24/outline'
import { useAppStore } from '@/lib/store/useAppStore'
import GlobalSearch from '@/components/search/GlobalSearch'
import NotificationCenter from '@/components/layout/NotificationCenter'
import { useKeyboardShortcuts } from '@/lib/hooks/useKeyboardShortcuts'
import { Card, CardContent } from '@/components/ui/Card'
import { slideUp, transition } from '@/lib/animations/variants'
import { cn } from '@/lib/utils/cn'

export default function Header() {
  const { toggleSidebar, theme, setTheme } = useAppStore()
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const router = useRouter()

  const toggleTheme = () => {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }

  const handleLogout = async () => {
    // Call logout API to clear server-side session
    try {
      const { logout } = await import('@/lib/auth')
      await logout()
    } catch (error) {
      console.error('Logout error:', error)
    }
    
    // Clear any other stored user data
    if (typeof window !== 'undefined') {
      localStorage.removeItem('user')
      localStorage.removeItem('user_preferences')
      localStorage.removeItem('selected_database')
    }
    
    // Redirect to login page
    router.push('/login')
    router.refresh()
  }

  // Global keyboard shortcuts
  useKeyboardShortcuts({
    onShortcut: (shortcut) => {
      if (shortcut.key === 'B' && shortcut.modifier === 'cmd') {
        toggleSidebar()
      } else if (shortcut.key === 'K' && shortcut.modifier === 'cmd') {
        // Focus search input - handled by GlobalSearch
        const searchInput = document.querySelector('input[placeholder="Search..."]') as HTMLInputElement | null
        searchInput?.focus()
      }
    },
  })

  return (
    <header className="sticky top-0 z-40 h-16 border-b border-border bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/80">
      <div className="flex h-full items-center justify-between px-4 lg:px-6">
        {/* Left side - Search */}
        <div className="flex flex-1 items-center gap-4">
          <GlobalSearch placeholder="Search... (⌘K)" />
        </div>

        {/* Right side - Actions */}
        <div className="flex items-center gap-2">
          {/* Theme toggle */}
          <motion.button
            onClick={toggleTheme}
            className="rounded-lg p-2 hover:bg-accent transition-colors"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            aria-label="Toggle theme"
          >
            {theme === 'dark' ? (
              <SunIcon className="h-5 w-5" />
            ) : (
              <MoonIcon className="h-5 w-5" />
            )}
          </motion.button>

          {/* Notifications */}
          <NotificationCenter />

          {/* User menu */}
          <div className="relative">
            <motion.button
              onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
              className="rounded-lg p-2 hover:bg-accent transition-colors"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              aria-label="User menu"
              aria-expanded={isUserMenuOpen}
            >
              <UserCircleIcon className="h-6 w-6" />
            </motion.button>

            <AnimatePresence>
              {isUserMenuOpen && (
                <>
                  {/* Backdrop to close menu */}
                  <div
                    className="fixed inset-0 z-40"
                    onClick={() => setIsUserMenuOpen(false)}
                  />
                  <motion.div
                    variants={slideUp}
                    initial="hidden"
                    animate="visible"
                    exit="hidden"
                    transition={transition}
                    className="absolute right-0 top-full z-50 mt-2 w-56 rounded-lg border border-border bg-card shadow-lg"
                  >
                    <Card>
                      <CardContent className="p-2">
                        <Link
                          href="/dashboard/users"
                          onClick={() => setIsUserMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
                            'hover:bg-accent hover:text-accent-foreground'
                          )}
                        >
                          <UserIcon className="h-5 w-5" />
                          <span>Profile</span>
                        </Link>
                        <Link
                          href="/dashboard/settings"
                          onClick={() => setIsUserMenuOpen(false)}
                          className={cn(
                            'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
                            'hover:bg-accent hover:text-accent-foreground'
                          )}
                        >
                          <Cog6ToothIcon className="h-5 w-5" />
                          <span>Settings</span>
                        </Link>
                        <div className="my-1 h-px bg-border" />
                        <button
                          onClick={handleLogout}
                          className={cn(
                            'flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
                            'hover:bg-destructive/10 hover:text-destructive'
                          )}
                        >
                          <ArrowRightOnRectangleIcon className="h-5 w-5" />
                          <span>Logout</span>
                        </button>
                      </CardContent>
                    </Card>
                  </motion.div>
                </>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>
    </header>
  )
}

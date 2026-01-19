'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import {
  UserCircleIcon,
  SunIcon,
  MoonIcon,
} from '@heroicons/react/24/outline'
import { useAppStore } from '@/lib/store/useAppStore'
import GlobalSearch from '@/components/search/GlobalSearch'
import NotificationCenter from '@/components/layout/NotificationCenter'
import { useKeyboardShortcuts } from '@/lib/hooks/useKeyboardShortcuts'

export default function Header() {
  const { toggleSidebar, theme, setTheme } = useAppStore()

  const toggleTheme = () => {
    setTheme(theme === 'dark' ? 'light' : 'dark')
  }

  // Global keyboard shortcuts
  useKeyboardShortcuts({
    onShortcut: (shortcut) => {
      if (shortcut.key === 'B' && shortcut.modifier === 'cmd') {
        toggleSidebar()
      } else if (shortcut.key === 'K' && shortcut.modifier === 'cmd') {
        // Focus search input - handled by GlobalSearch
        document.querySelector('input[placeholder="Search..."]')?.focus()
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
          <motion.button
            className="rounded-lg p-2 hover:bg-accent transition-colors"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            <UserCircleIcon className="h-6 w-6" />
          </motion.button>
        </div>
      </div>
    </header>
  )
}

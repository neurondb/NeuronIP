'use client'

import Link from 'next/link'
import { motion } from 'framer-motion'
import { cn } from '@/lib/utils/cn'

interface SimpleFooterProps {
  className?: string
  showLinks?: boolean
}

export default function SimpleFooter({ className, showLinks = true }: SimpleFooterProps) {
  const currentYear = new Date().getFullYear()

  const links = [
    { name: 'Documentation', href: '/docs' },
    { name: 'Support', href: '/dashboard/support' },
    { name: 'Privacy', href: '/privacy' },
    { name: 'Terms', href: '/terms' },
  ]

  return (
    <footer
      className={cn(
        'border-t border-border bg-card/50 backdrop-blur supports-[backdrop-filter]:bg-card/30',
        'py-6 px-4 sm:px-6 lg:px-8',
        className
      )}
    >
      <div className="max-w-7xl mx-auto">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          {/* Copyright */}
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.3 }}
            className="text-sm text-muted-foreground text-center sm:text-left"
          >
            &copy; {currentYear} NeuronIP. All rights reserved.
          </motion.p>

          {/* Links */}
          {showLinks && (
            <motion.nav
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.3, delay: 0.1 }}
              aria-label="Footer navigation"
              className="flex flex-wrap items-center justify-center gap-4 sm:gap-6"
            >
              {links.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className="text-sm text-muted-foreground hover:text-foreground transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 rounded"
                >
                  {link.name}
                </Link>
              ))}
            </motion.nav>
          )}
        </div>
      </div>
    </footer>
  )
}

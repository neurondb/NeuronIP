'use client'

import { cn } from '@/lib/utils/cn'

import SimpleFooter from './SimpleFooter'

interface DashboardFooterProps {
  className?: string
  showStats?: boolean
  version?: string
}

export default function DashboardFooter({
  className,
  showStats = false,
  version = '1.0.0',
}: DashboardFooterProps) {

  return (
    <footer
      className={cn(
        'border-t border-border bg-card/50 backdrop-blur supports-[backdrop-filter]:bg-card/30',
        className
      )}
    >
      <div className="py-6 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          {/* Simple Footer */}
          <SimpleFooter showLinks={true} />
        </div>
      </div>
    </footer>
  )
}

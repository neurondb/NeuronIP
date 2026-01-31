'use client'

import { ChevronRightIcon, HomeIcon } from '@heroicons/react/24/outline'
import Link from 'next/link'
import { usePathname } from 'next/navigation'

import { getRouteMeta } from '@/lib/pageArchetypes'
import { cn } from '@/lib/utils/cn'

export default function Breadcrumbs() {
  const pathname = usePathname()
  if (!pathname || pathname === '/login') return null

  const segments = pathname.split('/').filter(Boolean)
  const crumbs: { path: string; label: string }[] = [{ path: '/', label: 'Home' }]
  let acc = ''
  for (const seg of segments) {
    acc += `/${seg}`
    const m = getRouteMeta(acc)
    crumbs.push({ path: acc, label: m?.title ?? seg })
  }

  return (
    <nav aria-label="Breadcrumb" className="hidden md:flex items-center gap-1 text-sm text-muted-foreground min-w-0">
      {crumbs.map((crumb, i) => {
        const isLast = i === crumbs.length - 1
        return (
          <span key={crumb.path} className="flex items-center gap-1 shrink-0">
            {i > 0 && <ChevronRightIcon className="h-4 w-4 shrink-0 opacity-50" aria-hidden />}
            {isLast ? (
              <span className="font-medium text-foreground truncate flex items-center gap-1" aria-current="page">
                {i === 0 && <HomeIcon className="h-4 w-4 shrink-0" aria-hidden />}
                {crumb.label}
              </span>
            ) : (
              <Link
                href={crumb.path}
                className={cn(
                  'truncate hover:text-accent-foreground transition-colors flex items-center gap-1'
                )}
              >
                {i === 0 && <HomeIcon className="h-4 w-4 shrink-0" aria-hidden />}
                {crumb.label}
              </Link>
            )}
          </span>
        )
      })}
    </nav>
  )
}

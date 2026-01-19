'use client'

import Link from 'next/link'
import Button from '@/components/ui/Button'

export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <div className="text-center space-y-4 max-w-md">
        <h1 className="text-4xl font-bold text-foreground">404</h1>
        <h2 className="text-2xl font-semibold text-foreground">Page Not Found</h2>
        <p className="text-muted-foreground">
          The page you're looking for doesn't exist or has been moved.
        </p>
        <div className="pt-4">
          <Link href="/dashboard">
            <Button>Go to Dashboard</Button>
          </Link>
        </div>
      </div>
    </div>
  )
}

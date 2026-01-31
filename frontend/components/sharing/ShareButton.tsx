'use client'

import { ArrowUpTrayIcon as ShareIcon } from '@heroicons/react/24/outline'
import { useState } from 'react'

import Button from '@/components/ui/Button'
import { microcopy } from '@/lib/copy/microcopy'

import ShareDialog from './ShareDialog'

interface ShareButtonProps {
  resourceType: string
  resourceId: string
  resourceName?: string
  variant?: 'default' | 'outline' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

export default function ShareButton({
  resourceType,
  resourceId,
  resourceName,
  variant = 'outline',
  size = 'md',
  className = '',
}: ShareButtonProps) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <>
      <Button
        onClick={() => setIsOpen(true)}
        variant={variant === 'default' ? 'primary' : variant}
        size={size}
        className={className}
      >
        <ShareIcon className="h-4 w-4 mr-2" />
        {microcopy.sharing.share}
      </Button>
      <ShareDialog
        open={isOpen}
        onOpenChange={setIsOpen}
        resourceType={resourceType}
        resourceId={resourceId}
        resourceName={resourceName}
      />
    </>
  )
}

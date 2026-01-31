'use client'

import { XMarkIcon, LinkIcon, CheckIcon } from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useState } from 'react'

import Button from '@/components/ui/Button'
import { Dialog, DialogContent } from '@/components/ui/Dialog'
import Input from '@/components/ui/Input'
import { showToast } from '@/components/ui/Toast'
import { microcopy } from '@/lib/copy/microcopy'

interface ShareDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  resourceType: string
  resourceId: string
  resourceName?: string
}

type Permission = 'view' | 'edit' | 'full'

export default function ShareDialog({
  open,
  onOpenChange,
  resourceType,
  resourceId,
  resourceName,
}: ShareDialogProps) {
  const [permission, setPermission] = useState<Permission>('view')
  const [link, setLink] = useState('')
  const [copied, setCopied] = useState(false)

  // Generate shareable link
  const generateLink = () => {
    const baseUrl = typeof window !== 'undefined' ? window.location.origin : ''
    const shareLink = `${baseUrl}/shared/${resourceType}/${resourceId}?permission=${permission}`
    setLink(shareLink)
  }

  const handleCopy = () => {
    if (link) {
      navigator.clipboard.writeText(link)
      setCopied(true)
      showToast(microcopy.sharing.linkCopied, 'success')
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const permissionOptions: { value: Permission; label: string; description: string }[] = [
    {
      value: 'view',
      label: microcopy.sharing.viewOnly,
      description: 'Can view but not edit',
    },
    {
      value: 'edit',
      label: microcopy.sharing.canEdit,
      description: 'Can view and edit',
    },
    {
      value: 'full',
      label: microcopy.sharing.fullAccess,
      description: 'Full access including sharing',
    },
  ]

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">{microcopy.sharing.share}</h3>
          <button
            onClick={() => onOpenChange(false)}
            className="text-muted-foreground hover:text-foreground"
          >
            <XMarkIcon className="h-5 w-5" />
          </button>
        </div>

        {resourceName && (
          <p className="text-sm text-muted-foreground mb-4">Sharing: {resourceName}</p>
        )}

        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-2 block">
              {microcopy.sharing.permissions}
            </label>
            <div className="space-y-2">
              {permissionOptions.map((option) => (
                <button
                  key={option.value}
                  onClick={() => setPermission(option.value)}
                  className={`
                    w-full text-left p-3 rounded-lg border transition-colors
                    ${permission === option.value
                      ? 'border-primary bg-primary/5'
                      : 'border-border hover:bg-accent'
                    }
                  `}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium text-sm">{option.label}</div>
                      <div className="text-xs text-muted-foreground mt-1">
                        {option.description}
                      </div>
                    </div>
                    {permission === option.value && (
                      <CheckIcon className="h-5 w-5 text-primary" />
                    )}
                  </div>
                </button>
              ))}
            </div>
          </div>

          <div>
            <Button onClick={generateLink} className="w-full" size="lg">
              <LinkIcon className="h-4 w-4 mr-2" />
              {microcopy.sharing.shareLink}
            </Button>
          </div>

          {link && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              className="space-y-2"
            >
              <Input
                value={link}
                readOnly
                className="font-mono text-xs"
              />
              <Button
                onClick={handleCopy}
                variant="outline"
                className="w-full"
                size="lg"
              >
                {copied ? (
                  <>
                    <CheckIcon className="h-4 w-4 mr-2" />
                    Copied!
                  </>
                ) : (
                  <>
                    <LinkIcon className="h-4 w-4 mr-2" />
                    {microcopy.sharing.copyLink}
                  </>
                )}
              </Button>
            </motion.div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

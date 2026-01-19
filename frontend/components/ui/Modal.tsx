'use client'

import * as Dialog from '@radix-ui/react-dialog'
import { motion, AnimatePresence } from 'framer-motion'
import { XMarkIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'

interface ModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title?: string
  description?: string
  children: React.ReactNode
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
}

const sizeClasses = {
  sm: 'max-w-md',
  md: 'max-w-lg',
  lg: 'max-w-2xl',
  xl: 'max-w-4xl',
  full: 'max-w-7xl',
}

export default function Modal({
  open,
  onOpenChange,
  title,
  description,
  children,
  size = 'md',
}: ModalProps) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <>
            <Dialog.Portal>
              <Dialog.Overlay asChild>
                <motion.div
                  className="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  onClick={() => onOpenChange(false)}
                />
              </Dialog.Overlay>
              <Dialog.Content asChild>
                <motion.div
                  className={cn(
                    'fixed left-1/2 top-1/2 z-50 -translate-x-1/2 -translate-y-1/2 rounded-lg border border-border bg-card p-6 shadow-xl',
                    sizeClasses[size],
                    'w-[90vw] max-h-[90vh] overflow-y-auto'
                  )}
                  initial={{ opacity: 0, scale: 0.95, y: -20 }}
                  animate={{ opacity: 1, scale: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.95, y: -20 }}
                  transition={{ duration: 0.2 }}
                >
                  {(title || description) && (
                    <div className="mb-4">
                      {title && (
                        <Dialog.Title className="text-xl font-semibold">{title}</Dialog.Title>
                      )}
                      {description && (
                        <Dialog.Description className="mt-2 text-sm text-muted-foreground">
                          {description}
                        </Dialog.Description>
                      )}
                    </div>
                  )}
                  {children}
                  <Dialog.Close asChild>
                    <button
                      className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
                      onClick={() => onOpenChange(false)}
                    >
                      <XMarkIcon className="h-5 w-5" />
                      <span className="sr-only">Close</span>
                    </button>
                  </Dialog.Close>
                </motion.div>
              </Dialog.Content>
            </Dialog.Portal>
          </>
        )}
      </AnimatePresence>
    </Dialog.Root>
  )
}

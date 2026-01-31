'use client'

import {
  EllipsisVerticalIcon,
  TrashIcon,
  DocumentDuplicateIcon as DuplicateIcon,
  ArrowUpIcon,
  ArrowDownIcon,
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useState } from 'react'


import { cn } from '@/lib/utils/cn'

import BlockComments from '../collaboration/BlockComments'
import { Button } from '../ui/Button'

interface BlockMenuProps {
  blockId: string
  onDelete?: () => void
  onDuplicate?: () => void
  onMoveUp?: () => void
  onMoveDown?: () => void
  onComment?: (blockId: string, content: string) => void
  comments?: Array<{
    id: string
    userId: string
    userName: string
    userAvatar?: string
    content: string
    createdAt: Date
    resolved?: boolean
  }>
  className?: string
}

export default function BlockMenu({
  blockId,
  onDelete,
  onDuplicate,
  onMoveUp,
  onMoveDown,
  onComment,
  comments = [],
  className,
}: BlockMenuProps) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className={cn('relative group', className)}>
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setIsOpen(!isOpen)}
          className="h-6 w-6 p-0"
        >
          <EllipsisVerticalIcon className="h-4 w-4" />
        </Button>
        <BlockComments blockId={blockId} comments={comments} onAddComment={onComment} />
      </div>

      <AnimatePresence>
        {isOpen && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setIsOpen(false)}
            />
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              className="absolute left-0 top-8 w-48 bg-popover border border-border rounded-lg shadow-lg z-50"
            >
              <div className="p-1">
                {onMoveUp && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      onMoveUp()
                      setIsOpen(false)
                    }}
                    className="w-full justify-start gap-2"
                  >
                    <ArrowUpIcon className="h-4 w-4" />
                    Move up
                  </Button>
                )}
                {onMoveDown && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      onMoveDown()
                      setIsOpen(false)
                    }}
                    className="w-full justify-start gap-2"
                  >
                    <ArrowDownIcon className="h-4 w-4" />
                    Move down
                  </Button>
                )}
                {onDuplicate && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      onDuplicate()
                      setIsOpen(false)
                    }}
                    className="w-full justify-start gap-2"
                  >
                    <DuplicateIcon className="h-4 w-4" />
                    Duplicate
                  </Button>
                )}
                {onDelete && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => {
                      onDelete()
                      setIsOpen(false)
                    }}
                    className="w-full justify-start gap-2 text-destructive hover:text-destructive"
                  >
                    <TrashIcon className="h-4 w-4" />
                    Delete
                  </Button>
                )}
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  )
}

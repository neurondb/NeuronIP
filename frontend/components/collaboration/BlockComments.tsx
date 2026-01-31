'use client'

import { ChatBubbleLeftIcon, XMarkIcon } from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useState } from 'react'


import { cn } from '@/lib/utils/cn'

import { Avatar, AvatarFallback, AvatarImage } from '../ui/Avatar'
import { Button } from '../ui/Button'

import CommentInput from './CommentInput'

interface Comment {
  id: string
  userId: string
  userName: string
  userAvatar?: string
  content: string
  createdAt: Date
  resolved?: boolean
}

interface BlockCommentsProps {
  blockId: string
  comments?: Comment[]
  onAddComment?: (blockId: string, content: string) => void
  onResolveComment?: (commentId: string) => void
  onDeleteComment?: (commentId: string) => void
  className?: string
}

export default function BlockComments({
  blockId,
  comments = [],
  onAddComment,
  onResolveComment,
  onDeleteComment,
  className,
}: BlockCommentsProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [showInput, setShowInput] = useState(false)

  const activeComments = comments.filter((c) => !c.resolved)

  const handleSubmit = (content: string) => {
    onAddComment?.(blockId, content)
    setShowInput(false)
  }

  return (
    <div className={cn('relative', className)}>
      {/* Comment Indicator */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'absolute -left-8 top-0 p-1 rounded hover:bg-accent transition-colors',
          'opacity-0 group-hover:opacity-100',
          activeComments.length > 0 && 'opacity-100'
        )}
      >
        <ChatBubbleLeftIcon
          className={cn(
            'h-4 w-4',
            activeComments.length > 0 ? 'text-primary' : 'text-muted-foreground'
          )}
        />
        {activeComments.length > 0 && (
          <span className="absolute -top-1 -right-1 h-4 w-4 rounded-full bg-primary text-white text-[10px] flex items-center justify-center">
            {activeComments.length}
          </span>
        )}
      </button>

      {/* Comments Panel */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -10 }}
            className="absolute left-0 top-8 w-80 bg-popover border border-border rounded-lg shadow-lg z-50"
          >
            <div className="p-4 border-b border-border flex items-center justify-between">
              <h3 className="font-semibold text-sm">Comments</h3>
              <Button variant="ghost" size="sm" onClick={() => setIsOpen(false)}>
                <XMarkIcon className="h-4 w-4" />
              </Button>
            </div>

            <div className="max-h-96 overflow-y-auto p-4 space-y-4">
              {comments.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-4">
                  No comments yet. Start a conversation!
                </p>
              ) : (
                comments.map((comment) => (
                  <motion.div
                    key={comment.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    className={cn(
                      'flex gap-3',
                      comment.resolved && 'opacity-50'
                    )}
                  >
                    <Avatar className="h-8 w-8 flex-shrink-0">
                      <AvatarImage src={comment.userAvatar} />
                      <AvatarFallback className="text-xs">
                        {comment.userName.charAt(0).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-sm font-medium">{comment.userName}</span>
                        <span className="text-xs text-muted-foreground">
                          {new Date(comment.createdAt).toLocaleDateString()}
                        </span>
                      </div>
                      <p className="text-sm text-foreground">{comment.content}</p>
                      <div className="flex items-center gap-2 mt-2">
                        {!comment.resolved && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => onResolveComment?.(comment.id)}
                            className="text-xs"
                          >
                            Resolve
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => onDeleteComment?.(comment.id)}
                          className="text-xs text-destructive"
                        >
                          Delete
                        </Button>
                      </div>
                    </div>
                  </motion.div>
                ))
              )}
            </div>

            {showInput ? (
              <div className="p-4 border-t border-border">
                <CommentInput onSubmit={handleSubmit} placeholder="Add a comment..." />
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowInput(false)}
                  className="mt-2"
                >
                  Cancel
                </Button>
              </div>
            ) : (
              <div className="p-4 border-t border-border">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowInput(true)}
                  className="w-full"
                >
                  Add comment
                </Button>
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

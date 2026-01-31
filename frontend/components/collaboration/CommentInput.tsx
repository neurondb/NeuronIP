'use client'

import { PaperAirplaneIcon } from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import { useState, useRef, KeyboardEvent } from 'react'

import Button from '@/components/ui/Button'
import { microcopy } from '@/lib/copy/microcopy'

interface CommentInputProps {
  onSubmit: (content: string) => void
  placeholder?: string
  onMention?: (query: string) => Promise<string[]>
  disabled?: boolean
  className?: string
}

export default function CommentInput({
  onSubmit,
  placeholder = microcopy.collaboration.comment,
  onMention,
  disabled = false,
  className = '',
}: CommentInputProps) {
  const [content, setContent] = useState('')
  const [showMentionSuggestions, setShowMentionSuggestions] = useState(false)
  const [mentionQuery, setMentionQuery] = useState('')
  const [mentionSuggestions, setMentionSuggestions] = useState<string[]>([])
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleSubmit = () => {
    if (content.trim() && !disabled) {
      onSubmit(content.trim())
      setContent('')
    }
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault()
      handleSubmit()
    }

    // Handle @ mentions
    if (e.key === '@') {
      setShowMentionSuggestions(true)
    }
  }

  const handleMention = async (query: string) => {
    if (onMention) {
      const suggestions = await onMention(query)
      setMentionSuggestions(suggestions)
    }
  }

  return (
    <div className={`relative ${className}`}>
      <div className="flex gap-2 items-end">
        <div className="flex-1 relative">
          <textarea
            ref={textareaRef}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            rows={3}
            className={`
              w-full rounded-lg border border-border bg-background px-4 py-3
              text-sm placeholder:text-muted-foreground
              focus:outline-none focus:ring-2 focus:ring-ring
              resize-none disabled:opacity-50 disabled:cursor-not-allowed
            `}
          />
          <div className="absolute bottom-2 right-2 text-xs text-muted-foreground">
            Press ⌘+Enter to submit
          </div>
        </div>
        <Button
          onClick={handleSubmit}
          disabled={!content.trim() || disabled}
          size="lg"
        >
          <PaperAirplaneIcon className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}

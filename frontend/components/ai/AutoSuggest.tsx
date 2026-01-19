'use client'

import React, { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { SparklesIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface Suggestion {
  id: string
  text: string
  type: 'query' | 'filter' | 'action'
  score?: number
}

interface AutoSuggestProps {
  query: string
  onSelect: (suggestion: string) => void
  className?: string
}

export default function AutoSuggest({ query, onSelect, className }: AutoSuggestProps) {
  const [suggestions, setSuggestions] = useState<Suggestion[]>([])

  // Mock AI suggestions - in production, call AI API
  React.useEffect(() => {
    if (!query.trim()) {
      setSuggestions([])
      return
    }

    // Simulate AI-generated suggestions
    const mockSuggestions: Suggestion[] = [
      { id: '1', text: `${query} and filter by date`, type: 'query', score: 0.9 },
      { id: '2', text: `Show results for ${query} in last 7 days`, type: 'filter', score: 0.8 },
      { id: '3', text: `Export ${query} data`, type: 'action', score: 0.7 },
    ]

    setSuggestions(mockSuggestions)
  }, [query])

  if (suggestions.length === 0) return null

  return (
    <AnimatePresence>
      {suggestions.length > 0 && (
        <motion.div
          variants={slideUp}
          initial="hidden"
          animate="visible"
          exit="hidden"
          transition={transition}
          className={cn('mt-2 rounded-lg border border-border bg-card shadow-lg p-2', className)}
        >
          <div className="flex items-center gap-2 px-2 py-1 text-xs text-muted-foreground">
            <SparklesIcon className="h-3 w-3" />
            <span>AI Suggestions</span>
          </div>
          <div className="space-y-1">
            {suggestions.map((suggestion) => (
              <button
                key={suggestion.id}
                onClick={() => onSelect(suggestion.text)}
                className="w-full rounded px-3 py-2 text-left text-sm hover:bg-accent transition-colors"
              >
                {suggestion.text}
              </button>
            ))}
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}

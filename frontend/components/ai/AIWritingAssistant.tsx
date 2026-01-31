'use client'

import { SparklesIcon, XMarkIcon } from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useState, useEffect } from 'react'

import { cn } from '@/lib/utils/cn'

import { Button } from '../ui/Button'

interface AIWritingAssistantProps {
  editor: any
  onSuggestion?: (suggestion: string) => void
  className?: string
}

export default function AIWritingAssistant({
  editor,
  onSuggestion,
  className,
}: AIWritingAssistantProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(false)

  const generateSuggestions = async (text: string) => {
    setIsLoading(true)
    try {
      // In production, this would call your AI API
      // For now, we'll simulate with mock suggestions
      await new Promise((resolve) => setTimeout(resolve, 1000))
      
      const mockSuggestions = [
        'Expand on this idea with more details',
        'Add a supporting example',
        'Clarify this point further',
        'Connect this to the main topic',
      ]
      
      setSuggestions(mockSuggestions)
    } catch (error) {
      console.error('Failed to generate suggestions:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleApplySuggestion = (suggestion: string) => {
    if (editor) {
      const { from, to } = editor.state.selection
      editor.chain().focus().insertContentAt(to, ' ' + suggestion).run()
    }
    onSuggestion?.(suggestion)
    setIsOpen(false)
  }

  const handleGenerateBlock = async () => {
    if (!editor) return
    
    setIsLoading(true)
    try {
      // In production, this would call your AI API to generate content
      await new Promise((resolve) => setTimeout(resolve, 1500))
      
      const generatedContent = 'This is AI-generated content. Replace with actual AI API call.'
      editor.chain().focus().insertContent(generatedContent).run()
    } catch (error) {
      console.error('Failed to generate block:', error)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (!editor || !isOpen) return

    const handleUpdate = () => {
      const { from, to } = editor.state.selection
      const text = editor.state.doc.textBetween(from, to) || editor.state.doc.textContent
      
      if (text.length > 10) {
        generateSuggestions(text)
      }
    }

    editor.on('selectionUpdate', handleUpdate)
    return () => {
      editor.off('selectionUpdate', handleUpdate)
    }
  }, [editor, isOpen])

  return (
    <div className={cn('relative', className)}>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setIsOpen(!isOpen)}
        className="gap-2"
      >
        <SparklesIcon className="h-4 w-4" />
        <span className="hidden sm:inline">AI</span>
      </Button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="absolute right-0 top-10 w-80 bg-popover border border-border rounded-lg shadow-lg z-50"
          >
            <div className="p-4 border-b border-border flex items-center justify-between">
              <h3 className="font-semibold text-sm">AI Writing Assistant</h3>
              <Button variant="ghost" size="sm" onClick={() => setIsOpen(false)}>
                <XMarkIcon className="h-4 w-4" />
              </Button>
            </div>

            <div className="p-4 space-y-3">
              <Button
                variant="outline"
                size="sm"
                onClick={handleGenerateBlock}
                disabled={isLoading}
                className="w-full justify-start gap-2"
              >
                <SparklesIcon className="h-4 w-4" />
                {isLoading ? 'Generating...' : 'Generate block'}
              </Button>

              {suggestions.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-muted-foreground mb-2">
                    Suggestions
                  </h4>
                  <div className="space-y-2">
                    {suggestions.map((suggestion, index) => (
                      <button
                        key={index}
                        onClick={() => handleApplySuggestion(suggestion)}
                        className="w-full text-left p-2 rounded hover:bg-accent text-sm transition-colors"
                      >
                        {suggestion}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {isLoading && (
                <div className="text-center py-4">
                  <div className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

'use client'

import { useState, useRef, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { cn } from '@/lib/utils/cn'

interface QueryEditorProps {
  value: string
  onChange: (value: string) => void
  onExecute?: () => void
  placeholder?: string
}

export default function QueryEditor({
  value,
  onChange,
  onExecute,
  placeholder = 'Enter your query or ask a question in natural language...',
}: QueryEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }, [value])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault()
      onExecute?.()
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Query Editor</CardTitle>
          <Button onClick={onExecute} disabled={!value.trim()}>
            Execute
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <textarea
          ref={textareaRef}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className={cn(
            'w-full min-h-[200px] rounded-lg border border-border bg-background px-4 py-3',
            'font-mono text-sm focus:outline-none focus:ring-2 focus:ring-ring',
            'resize-none'
          )}
        />
        <p className="text-xs text-muted-foreground mt-2">
          Press Cmd/Ctrl + Enter to execute
        </p>
      </CardContent>
    </Card>
  )
}

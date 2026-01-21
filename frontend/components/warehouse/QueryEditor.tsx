'use client'

import { useState, useRef, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { cn } from '@/lib/utils/cn'
import Tooltip from '@/components/ui/Tooltip'
import HelpText from '@/components/ui/HelpText'
import { InformationCircleIcon, PlayIcon } from '@heroicons/react/24/outline'

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
          <div className="flex items-center gap-2">
            <CardTitle>Query Editor</CardTitle>
            <Tooltip
              content={
                <div>
                  <p className="font-medium mb-1">Query Editor Help</p>
                  <p className="text-xs mb-2">
                    You can enter either natural language questions or SQL queries.
                  </p>
                  <ul className="text-xs space-y-1">
                    <li>• Natural language: "What are the top 10 products by sales?"</li>
                    <li>• SQL: SELECT * FROM products LIMIT 10</li>
                    <li>• Press Cmd/Ctrl + Enter to execute</li>
                  </ul>
                </div>
              }
              variant="info"
            >
              <InformationCircleIcon className="h-4 w-4 text-muted-foreground cursor-help" />
            </Tooltip>
          </div>
          <Button onClick={onExecute} disabled={!value.trim()}>
            <PlayIcon className="h-4 w-4 mr-2" />
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
        <div className="mt-2 flex items-start justify-between gap-4">
          <p className="text-xs text-muted-foreground">
            Press Cmd/Ctrl + Enter to execute
          </p>
          <HelpText
            variant="tooltip"
            content={
              <div>
                <p className="font-medium mb-1">Query Syntax Help</p>
                <p className="text-xs mb-2">Supported query types:</p>
                <ul className="text-xs space-y-1">
                  <li>• Natural language questions</li>
                  <li>• Standard SQL queries</li>
                  <li>• Parameterized queries</li>
                </ul>
                <p className="text-xs mt-2">
                  <a
                    href="/docs/features/warehouse-qa"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline"
                  >
                    Learn more →
                  </a>
                </p>
              </div>
            }
          />
        </div>
      </CardContent>
    </Card>
  )
}

'use client'

import { useCallback, useEffect, useState } from 'react'

import { cn } from '@/lib/utils/cn'

export interface InlineEditableProps {
  value: string
  onSave: (value: string) => void
  placeholder?: string
  className?: string
  /** Optional class for the display span */
  displayClassName?: string
  /** Optional class for the input */
  inputClassName?: string
  disabled?: boolean
}

/**
 * Notion-like inline edit: display value, double-click or focus to edit.
 * Enter or blur saves; Escape cancels.
 */
export function InlineEditable({
  value,
  onSave,
  placeholder = 'Untitled',
  className,
  displayClassName,
  inputClassName,
  disabled = false,
}: InlineEditableProps) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState(value)

  useEffect(() => {
    setDraft(value)
  }, [value])

  const commit = useCallback(() => {
    const t = draft.trim()
    if (t !== value) onSave(t || value)
    setDraft(value)
    setEditing(false)
  }, [draft, value, onSave])

  const cancel = useCallback(() => {
    setDraft(value)
    setEditing(false)
  }, [value])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        e.preventDefault()
        commit()
      } else if (e.key === 'Escape') {
        e.preventDefault()
        cancel()
      }
    },
    [commit, cancel]
  )

  if (disabled) {
    return <span className={cn('text-foreground', displayClassName, className)}>{value || placeholder}</span>
  }

  if (editing) {
    return (
      <input
        autoFocus
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onBlur={commit}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        className={cn(
          'w-full min-w-0 rounded border border-input bg-background px-1.5 py-0.5 text-sm outline-none',
          'focus:ring-2 focus:ring-ring focus:ring-offset-1 focus:border-transparent',
          inputClassName,
          className
        )}
        aria-label="Edit"
      />
    )
  }

  return (
    <span
      role="button"
      tabIndex={0}
      onClick={() => setEditing(true)}
      onDoubleClick={() => setEditing(true)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          setEditing(true)
        }
      }}
      className={cn(
        'cursor-text rounded px-1 -mx-1 hover:bg-muted/50 transition-colors duration-fast',
        displayClassName,
        className
      )}
      aria-label="Edit (double-click or press Enter)"
    >
      {value || <span className="text-muted-foreground">{placeholder}</span>}
    </span>
  )
}

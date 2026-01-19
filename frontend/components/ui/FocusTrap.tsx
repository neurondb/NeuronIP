'use client'

import { ReactNode, useEffect, useRef } from 'react'

interface FocusTrapProps {
  children: ReactNode
  enabled?: boolean
}

export default function FocusTrap({ children, enabled = true }: FocusTrapProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const previousActiveElementRef = useRef<HTMLElement | null>(null)

  useEffect(() => {
    if (!enabled || !containerRef.current) return

    // Store the previously focused element
    previousActiveElementRef.current = document.activeElement as HTMLElement

    const container = containerRef.current
    const focusableElements = container.querySelectorAll(
      'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )

    const firstElement = focusableElements[0] as HTMLElement
    const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement

    // Focus first element
    if (firstElement) {
      firstElement.focus()
    }

    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return

      if (e.shiftKey) {
        // Shift + Tab
        if (document.activeElement === firstElement) {
          e.preventDefault()
          lastElement?.focus()
        }
      } else {
        // Tab
        if (document.activeElement === lastElement) {
          e.preventDefault()
          firstElement?.focus()
        }
      }
    }

    container.addEventListener('keydown', handleTab)

    return () => {
      container.removeEventListener('keydown', handleTab)
      // Restore focus to previous element
      previousActiveElementRef.current?.focus()
    }
  }, [enabled])

  return (
    <div ref={containerRef} tabIndex={-1}>
      {children}
    </div>
  )
}

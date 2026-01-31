'use client'

import { useEffect, useCallback } from 'react'

export interface KeyboardShortcut {
  key: string
  ctrlKey?: boolean
  metaKey?: boolean
  shiftKey?: boolean
  altKey?: boolean
  action: () => void
  description?: string
}

export function useKeyboardShortcuts(shortcuts: KeyboardShortcut[], enabled = true) {
  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      if (!enabled) return

      for (const shortcut of shortcuts) {
        const keyMatches = event.key.toLowerCase() === shortcut.key.toLowerCase()
        const ctrlMatches = shortcut.ctrlKey ? event.ctrlKey : !event.ctrlKey
        const metaMatches = shortcut.metaKey ? event.metaKey : !event.metaKey
        const shiftMatches = shortcut.shiftKey ? event.shiftKey : !event.shiftKey
        const altMatches = shortcut.altKey ? event.altKey : !event.altKey

        // Handle platform-specific meta key (Cmd on Mac, Ctrl on Windows/Linux)
        const isMac = typeof navigator !== 'undefined' && /Mac|iPhone|iPod|iPad/i.test(navigator.platform)
        const metaKeyMatches = isMac
          ? (shortcut.metaKey ? event.metaKey : !event.metaKey)
          : (shortcut.ctrlKey ? event.ctrlKey : !event.ctrlKey)

        if (
          keyMatches &&
          (shortcut.metaKey ? metaKeyMatches : true) &&
          (shortcut.ctrlKey && !isMac ? ctrlMatches : true) &&
          shiftMatches &&
          altMatches
        ) {
          // Don't trigger if user is typing in an input/textarea
          const target = event.target as HTMLElement
          if (
            target.tagName === 'INPUT' ||
            target.tagName === 'TEXTAREA' ||
            target.isContentEditable
          ) {
            // Allow some shortcuts even in inputs (like Cmd+K for command palette)
            if (shortcut.key !== 'k' && shortcut.key !== '/') {
              continue
            }
          }

          event.preventDefault()
          shortcut.action()
          break
        }
      }
    },
    [shortcuts, enabled]
  )

  useEffect(() => {
    if (!enabled) return

    window.addEventListener('keydown', handleKeyDown)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [handleKeyDown, enabled])
}

// Common keyboard shortcuts for block-based editing
export const notionShortcuts: KeyboardShortcut[] = [
  {
    key: 'k',
    metaKey: true,
    description: 'Open command palette',
    action: () => {
      // This would trigger the command palette
      const event = new CustomEvent('open-command-palette')
      window.dispatchEvent(event)
    },
  },
  {
    key: '/',
    metaKey: true,
    description: 'Open slash commands',
    action: () => {
      // This would trigger slash commands in the editor
      const event = new CustomEvent('open-slash-commands')
      window.dispatchEvent(event)
    },
  },
  {
    key: 'b',
    metaKey: true,
    description: 'Bold',
    action: () => {
      const event = new CustomEvent('editor-format', { detail: { format: 'bold' } })
      window.dispatchEvent(event)
    },
  },
  {
    key: 'i',
    metaKey: true,
    description: 'Italic',
    action: () => {
      const event = new CustomEvent('editor-format', { detail: { format: 'italic' } })
      window.dispatchEvent(event)
    },
  },
  {
    key: 'm',
    metaKey: true,
    shiftKey: true,
    description: 'Toggle comments',
    action: () => {
      const event = new CustomEvent('toggle-comments')
      window.dispatchEvent(event)
    },
  },
  {
    key: 'Escape',
    description: 'Close modals',
    action: () => {
      const event = new CustomEvent('close-modals')
      window.dispatchEvent(event)
    },
  },
]

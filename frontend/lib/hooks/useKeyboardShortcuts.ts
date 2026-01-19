import { useEffect } from 'react'
import { KEYBOARD_SHORTCUTS, KeyboardShortcut, matchesShortcut } from '../constants/shortcuts'

interface UseKeyboardShortcutsOptions {
  onShortcut?: (shortcut: KeyboardShortcut) => void
  shortcuts?: KeyboardShortcut[]
  enabled?: boolean
}

export function useKeyboardShortcuts({
  onShortcut,
  shortcuts = KEYBOARD_SHORTCUTS,
  enabled = true,
}: UseKeyboardShortcutsOptions = {}) {
  useEffect(() => {
    if (!enabled || !onShortcut) return

    const handleKeyDown = (event: KeyboardEvent) => {
      for (const shortcut of shortcuts) {
        if (matchesShortcut(event, shortcut)) {
          event.preventDefault()
          onShortcut(shortcut)
          break
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [enabled, onShortcut, shortcuts])
}

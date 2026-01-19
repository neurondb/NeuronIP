export interface KeyboardShortcut {
  key: string
  description: string
  category: string
  modifier?: 'ctrl' | 'cmd' | 'shift' | 'alt'
}

export const KEYBOARD_SHORTCUTS: KeyboardShortcut[] = [
  {
    key: 'K',
    modifier: 'cmd',
    description: 'Open global search',
    category: 'Navigation',
  },
  {
    key: '?',
    modifier: 'cmd',
    description: 'Show keyboard shortcuts',
    category: 'Navigation',
  },
  {
    key: 'B',
    modifier: 'cmd',
    description: 'Toggle sidebar',
    category: 'Navigation',
  },
  {
    key: 'N',
    modifier: 'cmd',
    description: 'New item',
    category: 'Actions',
  },
  {
    key: 'Escape',
    description: 'Close modal/dialog',
    category: 'Navigation',
  },
  {
    key: 'Enter',
    description: 'Confirm/Submit',
    category: 'Actions',
  },
  {
    key: '/',
    description: 'Focus search',
    category: 'Navigation',
  },
  {
    key: 'G',
    modifier: 'cmd',
    description: 'Go to dashboard',
    category: 'Navigation',
  },
]

export function getShortcutKey(shortcut: KeyboardShortcut): string {
  const modifier = shortcut.modifier === 'cmd' ? '⌘' : shortcut.modifier === 'ctrl' ? 'Ctrl' : ''
  return modifier ? `${modifier} + ${shortcut.key}` : shortcut.key
}

export function matchesShortcut(event: KeyboardEvent, shortcut: KeyboardShortcut): boolean {
  const key = event.key.toUpperCase()
  const shortcutKey = shortcut.key.toUpperCase()

  if (shortcut.modifier === 'cmd') {
    return (event.metaKey || event.ctrlKey) && key === shortcutKey
  }

  if (shortcut.modifier === 'ctrl') {
    return event.ctrlKey && !event.metaKey && key === shortcutKey
  }

  if (shortcut.modifier === 'shift') {
    return event.shiftKey && key === shortcutKey
  }

  if (shortcut.modifier === 'alt') {
    return event.altKey && key === shortcutKey
  }

  return key === shortcutKey && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey
}

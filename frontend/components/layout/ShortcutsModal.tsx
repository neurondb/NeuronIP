'use client'

import { useState, useEffect } from 'react'
import Modal from '@/components/ui/Modal'
import { KEYBOARD_SHORTCUTS, getShortcutKey } from '@/lib/constants/shortcuts'
import { useKeyboardShortcuts } from '@/lib/hooks/useKeyboardShortcuts'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'

export default function ShortcutsModal() {
  const [isOpen, setIsOpen] = useState(false)

  useKeyboardShortcuts({
    onShortcut: (shortcut) => {
      if (shortcut.key === '?' && (shortcut.modifier === 'cmd' || shortcut.modifier === 'ctrl')) {
        setIsOpen(true)
      }
    },
  })

  // Prevent opening if user is typing in an input
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === '?' && (e.metaKey || e.ctrlKey)) {
        const target = e.target as HTMLElement
        if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') {
          return
        }
        e.preventDefault()
        setIsOpen(true)
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  const categories = Array.from(new Set(KEYBOARD_SHORTCUTS.map((s) => s.category)))

  return (
    <Modal
      open={isOpen}
      onOpenChange={setIsOpen}
      title="Keyboard Shortcuts"
      description="Keyboard shortcuts to navigate the application faster"
      size="lg"
    >
      <div className="space-y-6">
        {categories.map((category) => (
          <div key={category}>
            <h3 className="text-sm font-semibold mb-2">{category}</h3>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Shortcut</TableHead>
                  <TableHead>Description</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {KEYBOARD_SHORTCUTS.filter((s) => s.category === category).map((shortcut) => (
                  <TableRow key={`${shortcut.modifier}-${shortcut.key}`}>
                    <TableCell>
                      <kbd className="px-2 py-1 text-xs font-semibold bg-muted rounded border border-border">
                        {getShortcutKey(shortcut)}
                      </kbd>
                    </TableCell>
                    <TableCell>{shortcut.description}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ))}
      </div>
    </Modal>
  )
}

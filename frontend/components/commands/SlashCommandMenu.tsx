'use client'

import { motion, AnimatePresence } from 'framer-motion'
import { useState, useEffect, useRef } from 'react'

import { cn } from '@/lib/utils/cn'

import { searchCommands, slashCommands } from './CommandRegistry'


interface SlashCommandMenuProps {
  editor: any
  query: string
  onSelect: (commandId: string) => void
  position: { top: number; left: number } | null
}

export default function SlashCommandMenu({
  editor,
  query,
  onSelect,
  position,
}: SlashCommandMenuProps) {
  const [selectedIndex, setSelectedIndex] = useState(0)
  const menuRef = useRef<HTMLDivElement>(null)

  const filteredCommands = query ? searchCommands(query) : slashCommands

  useEffect(() => {
    setSelectedIndex(0)
  }, [query])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!position) return

      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSelectedIndex((prev) => (prev + 1) % filteredCommands.length)
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSelectedIndex((prev) => (prev - 1 + filteredCommands.length) % filteredCommands.length)
      } else if (e.key === 'Enter') {
        e.preventDefault()
        if (filteredCommands[selectedIndex]) {
          onSelect(filteredCommands[selectedIndex].id)
        }
      } else if (e.key === 'Escape') {
        e.preventDefault()
        onSelect('')
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [selectedIndex, filteredCommands, position, onSelect])

  if (!position || filteredCommands.length === 0) {
    return null
  }

  const groupedCommands = filteredCommands.reduce(
    (acc, cmd) => {
      if (!acc[cmd.group]) {
        acc[cmd.group] = []
      }
      acc[cmd.group].push(cmd)
      return acc
    },
    {} as Record<string, typeof filteredCommands>
  )

  return (
    <AnimatePresence>
      <motion.div
        ref={menuRef}
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -10 }}
        transition={{ duration: 0.15 }}
        className="fixed z-50 w-64 rounded-lg border border-border bg-popover shadow-lg"
        style={{
          top: `${position.top}px`,
          left: `${position.left}px`,
        }}
      >
        <div className="max-h-80 overflow-y-auto p-1">
          {Object.entries(groupedCommands).map(([group, commands]) => (
            <div key={group} className="mb-1">
              <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase">
                {group}
              </div>
              {commands.map((command, index) => {
                const globalIndex = filteredCommands.indexOf(command)
                const isSelected = globalIndex === selectedIndex
                const Icon = command.icon

                return (
                  <button
                    key={command.id}
                    onClick={() => onSelect(command.id)}
                    className={cn(
                      'w-full flex items-center gap-3 rounded-md px-2 py-2 text-sm transition-colors',
                      isSelected
                        ? 'bg-accent text-accent-foreground'
                        : 'hover:bg-accent/50 text-foreground'
                    )}
                    onMouseEnter={() => setSelectedIndex(globalIndex)}
                  >
                    {Icon && <Icon className="h-4 w-4 flex-shrink-0" />}
                    <div className="flex-1 text-left">
                      <div className="font-medium">{command.label}</div>
                      <div className="text-xs text-muted-foreground">{command.description}</div>
                    </div>
                  </button>
                )
              })}
            </div>
          ))}
        </div>
      </motion.div>
    </AnimatePresence>
  )
}

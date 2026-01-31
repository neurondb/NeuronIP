'use client'

import { motion } from 'framer-motion'

import { typingIndicator } from '@/lib/animations/variants'

interface TypingIndicatorProps {
  names?: string[]
  className?: string
}

export default function TypingIndicator({ names = [], className = '' }: TypingIndicatorProps) {
  if (names.length === 0) {
    return (
      <div className={`flex items-center gap-2 text-sm text-muted-foreground ${className}`}>
        <div className="flex gap-1">
          <motion.div
            className="w-2 h-2 rounded-full bg-muted-foreground"
            animate={{ opacity: [0.4, 1, 0.4] }}
            transition={{ duration: 1.5, repeat: Infinity, delay: 0 }}
          />
          <motion.div
            className="w-2 h-2 rounded-full bg-muted-foreground"
            animate={{ opacity: [0.4, 1, 0.4] }}
            transition={{ duration: 1.5, repeat: Infinity, delay: 0.2 }}
          />
          <motion.div
            className="w-2 h-2 rounded-full bg-muted-foreground"
            animate={{ opacity: [0.4, 1, 0.4] }}
            transition={{ duration: 1.5, repeat: Infinity, delay: 0.4 }}
          />
        </div>
        <span>Thinking...</span>
      </div>
    )
  }

  const nameText = names.length === 1 
    ? `${names[0]} is typing...`
    : names.length === 2
    ? `${names[0]} and ${names[1]} are typing...`
    : `${names[0]} and ${names.length - 1} others are typing...`

  return (
    <div className={`flex items-center gap-2 text-sm text-muted-foreground ${className}`}>
      <div className="flex gap-1">
        <motion.div
          className="w-2 h-2 rounded-full bg-muted-foreground"
          animate={{ opacity: [0.4, 1, 0.4] }}
          transition={{ duration: 1.5, repeat: Infinity, delay: 0 }}
        />
        <motion.div
          className="w-2 h-2 rounded-full bg-muted-foreground"
          animate={{ opacity: [0.4, 1, 0.4] }}
          transition={{ duration: 1.5, repeat: Infinity, delay: 0.2 }}
        />
        <motion.div
          className="w-2 h-2 rounded-full bg-muted-foreground"
          animate={{ opacity: [0.4, 1, 0.4] }}
          transition={{ duration: 1.5, repeat: Infinity, delay: 0.4 }}
        />
      </div>
      <span>{nameText}</span>
    </div>
  )
}

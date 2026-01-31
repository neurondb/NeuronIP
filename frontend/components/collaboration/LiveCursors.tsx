'use client'

import { motion, AnimatePresence } from 'framer-motion'
import { useEffect, useState, useRef } from 'react'

import { cn } from '@/lib/utils/cn'

import { Avatar, AvatarFallback, AvatarImage } from '../ui/Avatar'


interface CursorPosition {
  userId: string
  userName: string
  userAvatar?: string
  x: number
  y: number
  color: string
}

interface LiveCursorsProps {
  roomId: string
  currentUserId: string
  className?: string
}

// Generate a color based on user ID
function getUserColor(userId: string): string {
  const colors = [
    '#ef4444', // red
    '#f59e0b', // amber
    '#10b981', // green
    '#3b82f6', // blue
    '#8b5cf6', // violet
    '#ec4899', // pink
    '#06b6d4', // cyan
    '#f97316', // orange
  ]
  const index = userId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return colors[index % colors.length]
}

export default function LiveCursors({
  roomId,
  currentUserId,
  className,
}: LiveCursorsProps) {
  const [cursors, setCursors] = useState<Map<string, CursorPosition>>(new Map())
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    // This would connect to WebSocket and listen for cursor updates
    // For now, we'll simulate with a mock connection
    const handleCursorUpdate = (data: {
      userId: string
      userName: string
      userAvatar?: string
      x: number
      y: number
    }) => {
      if (data.userId === currentUserId) return

      setCursors((prev) => {
        const next = new Map(prev)
        next.set(data.userId, {
          ...data,
          color: getUserColor(data.userId),
        })
        return next
      })

      // Remove cursor after 2 seconds of inactivity
      setTimeout(() => {
        setCursors((prev) => {
          const next = new Map(prev)
          next.delete(data.userId)
          return next
        })
      }, 2000)
    }

    // Mock: simulate cursor movements
    // In production, this would come from WebSocket
    const mockCursor = setInterval(() => {
      if (Math.random() > 0.7) {
        handleCursorUpdate({
          userId: 'user-' + Math.floor(Math.random() * 3),
          userName: 'User ' + Math.floor(Math.random() * 3),
          x: Math.random() * (containerRef.current?.clientWidth || 800),
          y: Math.random() * (containerRef.current?.clientHeight || 600),
        })
      }
    }, 1000)

    return () => {
      clearInterval(mockCursor)
    }
  }, [roomId, currentUserId])

  // Track local mouse movements
  useEffect(() => {
    if (!containerRef.current) return

    const handleMouseMove = (e: MouseEvent) => {
      // In production, send cursor position to WebSocket
      // For now, we'll just track locally
    }

    containerRef.current.addEventListener('mousemove', handleMouseMove)
    return () => {
      containerRef.current?.removeEventListener('mousemove', handleMouseMove)
    }
  }, [])

  return (
    <div ref={containerRef} className={cn('relative w-full h-full', className)}>
      <AnimatePresence>
        {Array.from(cursors.values()).map((cursor) => (
          <motion.div
            key={cursor.userId}
            initial={{ opacity: 0, scale: 0 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0 }}
            transition={{ duration: 0.2 }}
            className="absolute pointer-events-none z-50"
            style={{
              left: cursor.x,
              top: cursor.y,
              transform: 'translate(-50%, -50%)',
            }}
          >
            <div className="flex items-center gap-2">
              <div
                className="w-4 h-4 rounded-full border-2 border-white shadow-lg"
                style={{ backgroundColor: cursor.color }}
              />
              <div className="px-2 py-1 bg-popover border border-border rounded-md shadow-lg flex items-center gap-2">
                <Avatar className="h-5 w-5">
                  <AvatarImage src={cursor.userAvatar} />
                  <AvatarFallback className="text-xs">
                    {cursor.userName.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="text-xs font-medium text-foreground">{cursor.userName}</span>
              </div>
            </div>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  )
}

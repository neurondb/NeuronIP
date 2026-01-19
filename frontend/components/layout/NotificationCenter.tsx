'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { BellIcon, XMarkIcon } from '@heroicons/react/24/outline'
import { Card, CardContent } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useWebSocket } from '@/lib/hooks/useWebSocket'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface Notification {
  id: string
  type: 'info' | 'success' | 'warning' | 'error'
  title: string
  message: string
  timestamp: Date
  read: boolean
}

export default function NotificationCenter() {
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [isOpen, setIsOpen] = useState(false)

  // WebSocket connection for real-time notifications
  const { isConnected, subscribe } = useWebSocket({
    url: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws',
    enabled: true,
  })

  subscribe('notification', (data: any) => {
    const notification: Notification = {
      id: Date.now().toString(),
      type: data.type || 'info',
      title: data.title || 'Notification',
      message: data.message || '',
      timestamp: new Date(),
      read: false,
    }
    setNotifications((prev) => [notification, ...prev])
  })

  const markAsRead = (id: string) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, read: true } : n))
    )
  }

  const removeNotification = (id: string) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id))
  }

  const unreadCount = notifications.filter((n) => !n.read).length

  return (
    <div className="relative">
      <motion.button
        onClick={() => setIsOpen(!isOpen)}
        className="relative rounded-lg p-2 hover:bg-accent transition-colors"
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.95 }}
      >
        <BellIcon className="h-5 w-5" />
        {unreadCount > 0 && (
          <span className="absolute right-1 top-1 h-5 w-5 flex items-center justify-center rounded-full bg-destructive text-xs text-destructive-foreground">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </motion.button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            variants={slideUp}
            initial="hidden"
            animate="visible"
            exit="hidden"
            transition={transition}
            className="absolute right-0 top-full z-50 mt-2 w-96 rounded-lg border border-border bg-card shadow-lg"
          >
            <Card>
              <CardContent className="p-4 max-h-[500px] overflow-y-auto">
                {notifications.length === 0 ? (
                  <div className="text-center text-muted-foreground py-8">
                    No notifications
                  </div>
                ) : (
                  <div className="space-y-2">
                    {notifications.map((notification) => (
                      <motion.div
                        key={notification.id}
                        variants={slideUp}
                        initial="hidden"
                        animate="visible"
                        className={cn(
                          'rounded-lg p-3 border border-border',
                          !notification.read && 'bg-primary/5'
                        )}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium">{notification.title}</p>
                            <p className="text-xs text-muted-foreground mt-1">
                              {notification.message}
                            </p>
                            <p className="text-xs text-muted-foreground mt-1">
                              {notification.timestamp.toLocaleTimeString()}
                            </p>
                          </div>
                          <button
                            onClick={() => removeNotification(notification.id)}
                            className="rounded p-1 hover:bg-destructive/10"
                          >
                            <XMarkIcon className="h-4 w-4" />
                          </button>
                        </div>
                      </motion.div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

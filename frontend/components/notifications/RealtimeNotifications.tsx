'use client'

import { useEffect } from 'react'
import { toast } from 'sonner'
import { useWebSocket } from '@/lib/websocket/hooks'
import { WebSocketMessage } from '@/lib/websocket/types'

interface Notification {
  id: string
  title: string
  message?: string
  type?: 'info' | 'success' | 'warning' | 'error'
  duration?: number
}

export function RealtimeNotifications() {
  const { subscribe } = useWebSocket()

  useEffect(() => {
    const unsubscribe = subscribe?.('notification', (message: WebSocketMessage) => {
      const notification = message.payload as Notification
      const { title, message: body, type = 'info', duration = 5000 } = notification

      switch (type) {
        case 'success':
          toast.success(title, { description: body, duration })
          break
        case 'error':
          toast.error(title, { description: body, duration })
          break
        case 'warning':
          toast.warning(title, { description: body, duration })
          break
        default:
          toast.info(title, { description: body, duration })
      }
    })

    return () => {
      unsubscribe?.()
    }
  }, [subscribe])

  return null
}

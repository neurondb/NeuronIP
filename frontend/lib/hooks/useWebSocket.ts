import { useEffect, useState, useRef } from 'react'
import { getWebSocketClient } from '../websocket/client'

interface UseWebSocketOptions {
  url?: string
  enabled?: boolean
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { url, enabled = true, onConnect, onDisconnect, onError } = options
  const [isConnected, setIsConnected] = useState(false)
  const handlersRef = useRef<Array<{ type: string; handler: (data: unknown) => void; unsubscribe: () => void }>>([])

  useEffect(() => {
    if (!enabled || !url) {
      return
    }

    try {
      const client = getWebSocketClient(url)

      const unsubscribeConnected = client.on('connected', () => {
        setIsConnected(true)
        onConnect?.()
      })

      const unsubscribeDisconnected = client.on('disconnected', () => {
        setIsConnected(false)
        onDisconnect?.()
      })

      const unsubscribeError = client.on('error', (error) => {
        onError?.(error as Event)
      })

      return () => {
        unsubscribeConnected()
        unsubscribeDisconnected()
        unsubscribeError()
        handlersRef.current.forEach(({ unsubscribe }) => unsubscribe())
        handlersRef.current = []
      }
    } catch (error) {
      console.error('Failed to initialize WebSocket:', error)
    }
  }, [enabled, url, onConnect, onDisconnect, onError])

  const subscribe = (type: string, handler: (data: unknown) => void) => {
    if (!url) return () => {}

    try {
      const client = getWebSocketClient()
      const unsubscribe = client.on(type, handler)
      handlersRef.current.push({ type, handler, unsubscribe })
      return unsubscribe
    } catch (error) {
      console.error('Failed to subscribe to WebSocket:', error)
      return () => {}
    }
  }

  const send = (type: string, payload: unknown) => {
    if (!url) return

    try {
      const client = getWebSocketClient()
      client.send(type, payload)
    } catch (error) {
      console.error('Failed to send WebSocket message:', error)
    }
  }

  return {
    isConnected,
    subscribe,
    send,
  }
}

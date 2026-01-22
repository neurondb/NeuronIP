'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { env } from '@/lib/utils/env'

import { WebSocketClient, createWebSocketClient } from './client'
import { WebSocketMessage, WebSocketStatus, PresenceUser } from './types'

export function useWebSocket() {
  const [status, setStatus] = useState<WebSocketStatus>('disconnected')
  const clientRef = useRef<WebSocketClient | null>(null)

  useEffect(() => {
    const wsUrl = env.NEXT_PUBLIC_WS_URL || env.NEXT_PUBLIC_API_URL?.replace('http', 'ws') + '/ws'
    
    const client = createWebSocketClient({
      url: wsUrl,
      onConnect: () => setStatus('connected'),
      onDisconnect: () => setStatus('disconnected'),
      onError: () => setStatus('error'),
    })

    clientRef.current = client
    client.connect()

    return () => {
      client.disconnect()
    }
  }, [])

  const send = useCallback((message: WebSocketMessage) => {
    clientRef.current?.send(message)
  }, [])

  const subscribe = useCallback((type: string, callback: (message: WebSocketMessage) => void) => {
    return clientRef.current?.subscribe(type, callback)
  }, [])

  return {
    status,
    send,
    subscribe,
    client: clientRef.current,
  }
}

export function useRealtimeQuery<T>(
  queryKey: string[],
  messageType: string,
  initialData?: T
) {
  const queryClient = useQueryClient()
  const { subscribe } = useWebSocket()

  useEffect(() => {
    const unsubscribe = subscribe?.(messageType, (message: WebSocketMessage) => {
      queryClient.setQueryData(queryKey, message.payload)
    })

    return () => {
      unsubscribe?.()
    }
  }, [queryKey, messageType, queryClient, subscribe])

  return useQuery({
    queryKey,
    initialData,
    refetchOnWindowFocus: false,
    refetchOnMount: false,
  })
}

export function usePresence(roomId: string) {
  const [users, setUsers] = useState<PresenceUser[]>([])
  const { send, subscribe } = useWebSocket()

  useEffect(() => {
    // Subscribe to presence updates
    const unsubscribe = subscribe?.('presence', (message: WebSocketMessage) => {
      const payload = message.payload as { roomId?: string; users?: PresenceUser[] }
      if (payload.roomId === roomId) {
        setUsers(payload.users || [])
      }
    })

    // Join presence room
    send({
      type: 'presence.join',
      payload: { roomId },
    })

    return () => {
      // Leave presence room
      send({
        type: 'presence.leave',
        payload: { roomId },
      })
      unsubscribe?.()
    }
  }, [roomId, send, subscribe])

  return users
}

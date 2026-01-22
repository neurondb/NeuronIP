import { WebSocketConfig, WebSocketMessage, WebSocketStatus } from './types'

export class WebSocketClient {
  private ws: WebSocket | null = null
  private config: WebSocketConfig
  private status: WebSocketStatus = 'disconnected'
  private reconnectAttempts = 0
  private reconnectTimer: NodeJS.Timeout | null = null
  private messageQueue: WebSocketMessage[] = []
  private listeners: Map<string, Set<(message: WebSocketMessage) => void>> = new Map()

  constructor(config: WebSocketConfig) {
    this.config = {
      reconnectInterval: 3000,
      maxReconnectAttempts: 10,
      ...config,
    }
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    this.status = 'connecting'
    try {
      this.ws = new WebSocket(this.config.url)

      this.ws.onopen = () => {
        this.status = 'connected'
        this.reconnectAttempts = 0
        this.config.onConnect?.()
        this.flushMessageQueue()
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          this.handleMessage(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      this.ws.onerror = (error) => {
        this.status = 'error'
        this.config.onError?.(error)
      }

      this.ws.onclose = () => {
        this.status = 'disconnected'
        this.config.onDisconnect?.()
        this.attemptReconnect()
      }
    } catch (error) {
      this.status = 'error'
      this.config.onError?.(error as Event)
      this.attemptReconnect()
    }
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.ws?.close()
    this.ws = null
    this.status = 'disconnected'
  }

  send(message: WebSocketMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      this.messageQueue.push(message)
    }
  }

  subscribe(type: string, callback: (message: WebSocketMessage) => void): () => void {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set())
    }
    this.listeners.get(type)!.add(callback)

    return () => {
      this.listeners.get(type)?.delete(callback)
    }
  }

  getStatus(): WebSocketStatus {
    return this.status
  }

  private handleMessage(message: WebSocketMessage): void {
    this.config.onMessage?.(message)

    // Notify type-specific listeners
    const typeListeners = this.listeners.get(message.type)
    if (typeListeners) {
      typeListeners.forEach((callback) => callback(message))
    }

    // Notify wildcard listeners
    const wildcardListeners = this.listeners.get('*')
    if (wildcardListeners) {
      wildcardListeners.forEach((callback) => callback(message))
    }
  }

  private flushMessageQueue(): void {
    while (this.messageQueue.length > 0 && this.ws?.readyState === WebSocket.OPEN) {
      const message = this.messageQueue.shift()
      if (message) {
        this.ws.send(JSON.stringify(message))
      }
    }
  }

  private attemptReconnect(): void {
    if (
      this.reconnectAttempts >= (this.config.maxReconnectAttempts || 10) ||
      this.reconnectTimer
    ) {
      return
    }

    this.reconnectAttempts++
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, this.config.reconnectInterval)
  }
}

let globalClient: WebSocketClient | null = null

export function createWebSocketClient(config: WebSocketConfig): WebSocketClient {
  if (globalClient) {
    globalClient.disconnect()
  }
  globalClient = new WebSocketClient(config)
  return globalClient
}

export function getWebSocketClient(): WebSocketClient | null {
  return globalClient
}

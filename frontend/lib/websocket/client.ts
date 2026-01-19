type MessageHandler = (data: unknown) => void

class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private reconnectDelay = 1000
  private handlers: Map<string, MessageHandler[]> = new Map()
  private isConnected = false

  constructor(url: string) {
    this.url = url
  }

  connect(): void {
    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        this.isConnected = true
        this.reconnectAttempts = 0
        this.emit('connected', {})
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          const { type, payload } = data
          this.emit(type, payload)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        this.emit('error', error)
      }

      this.ws.onclose = () => {
        this.isConnected = false
        this.emit('disconnected', {})
        this.reconnect()
      }
    } catch (error) {
      console.error('Failed to connect WebSocket:', error)
      this.reconnect()
    }
  }

  private reconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached')
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)

    setTimeout(() => {
      console.log(`Reconnecting... (attempt ${this.reconnectAttempts})`)
      this.connect()
    }, delay)
  }

  private emit(type: string, payload: unknown): void {
    const handlers = this.handlers.get(type) || []
    handlers.forEach((handler) => handler(payload))
  }

  on(type: string, handler: MessageHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, [])
    }
    this.handlers.get(type)!.push(handler)

    // Return unsubscribe function
    return () => {
      const handlers = this.handlers.get(type) || []
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  send(type: string, payload: unknown): void {
    if (this.ws && this.isConnected) {
      this.ws.send(JSON.stringify({ type, payload }))
    } else {
      console.warn('WebSocket not connected')
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.handlers.clear()
  }

  getConnected(): boolean {
    return this.isConnected
  }
}

// Singleton instance
let wsClient: WebSocketClient | null = null

export function getWebSocketClient(url?: string): WebSocketClient {
  if (!wsClient && url) {
    wsClient = new WebSocketClient(url)
    wsClient.connect()
  }
  if (!wsClient) {
    throw new Error('WebSocket client not initialized')
  }
  return wsClient
}

export function createWebSocketClient(url: string): WebSocketClient {
  const client = new WebSocketClient(url)
  client.connect()
  return client
}

export default WebSocketClient

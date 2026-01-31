/**
 * Streaming API utilities for real-time responses
 */

export interface StreamingOptions {
  onChunk?: (chunk: string) => void
  onComplete?: (fullText: string) => void
  onError?: (error: Error) => void
}

/**
 * Stream a response from a Server-Sent Events (SSE) endpoint
 */
export async function streamSSE(
  url: string,
  options: StreamingOptions & { body?: any; headers?: Record<string, string> }
): Promise<void> {
  const { onChunk, onComplete, onError, body, headers = {} } = options

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'text/event-stream',
        ...headers,
      },
      body: body ? JSON.stringify(body) : undefined,
    })

    if (!response.ok) {
      throw new Error(`Stream failed: ${response.statusText}`)
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let fullText = ''

    if (!reader) {
      throw new Error('No response body')
    }

    // eslint-disable-next-line no-constant-condition -- stream read loop
    while (true) {
      const { done, value } = await reader.read()

      if (done) {
        break
      }

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') {
            onComplete?.(fullText)
            return
          }

          try {
            const parsed = JSON.parse(data)
            const chunk = parsed.content || parsed.text || parsed.chunk || ''
            fullText += chunk
            onChunk?.(chunk)
          } catch {
            // If not JSON, treat as plain text
            fullText += data
            onChunk?.(data)
          }
        }
      }
    }

    onComplete?.(fullText)
  } catch (error) {
    onError?.(error instanceof Error ? error : new Error('Streaming failed'))
  }
}

/**
 * Stream a response from a WebSocket connection
 */
export function streamWebSocket(
  url: string,
  message: any,
  options: StreamingOptions
): () => void {
  const { onChunk, onComplete, onError } = options
  let fullText = ''

  const ws = new WebSocket(url)

  ws.onopen = () => {
    ws.send(JSON.stringify(message))
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      const chunk = data.content || data.text || data.chunk || ''
      fullText += chunk
      onChunk?.(chunk)
    } catch {
      // If not JSON, treat as plain text
      fullText += event.data
      onChunk?.(event.data)
    }
  }

  ws.onerror = (error) => {
    onError?.(new Error('WebSocket error'))
  }

  ws.onclose = () => {
    onComplete?.(fullText)
  }

  return () => {
    ws.close()
  }
}

'use client'

import { useEffect, useState } from 'react'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'

interface ExecutionEvent {
  execution_id: string
  step_id: string
  event_type: string
  data: Record<string, any>
  timestamp: string
}

interface LiveExecutionTraceProps {
  executionId: string
}

export default function LiveExecutionTrace({ executionId }: LiveExecutionTraceProps) {
  const [events, setEvents] = useState<ExecutionEvent[]>([])
  const [connected, setConnected] = useState(false)

  useEffect(() => {
    const ws = new WebSocket(`ws://localhost:8080/api/v1/workflows/${executionId}/stream`)
    
    ws.onopen = () => setConnected(true)
    ws.onclose = () => setConnected(false)
    ws.onerror = () => setConnected(false)
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      setEvents(prev => [...prev, data])
    }
    
    return () => ws.close()
  }, [executionId])

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Live Execution Trace
          <span className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'}`} />
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2 max-h-96 overflow-y-auto">
          {events.map((event, i) => (
            <div key={i} className="p-2 border rounded text-sm">
              <div className="font-medium">{event.step_id}</div>
              <div className="text-xs text-gray-500">{event.event_type}</div>
              <div className="text-xs mt-1">{new Date(event.timestamp).toLocaleTimeString()}</div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

'use client'

import { useState, useEffect } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'

interface Trace {
  id: string
  agent_id: string
  task: string
  status: string
  steps: TraceStep[]
  start_time: string
  end_time?: string
  duration?: number
}

interface TraceStep {
  id: string
  step_number: number
  step_type: string
  description: string
  tool_name?: string
  duration: number
  timestamp: string
}

export default function AgentTraces() {
  const [traces, setTraces] = useState<Trace[]>([])
  const [selectedTrace, setSelectedTrace] = useState<Trace | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchTraces()
  }, [])

  const fetchTraces = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/observability/agent/traces')
      const data = await response.json()
      setTraces(data)
    } catch (error) {
      console.error('Failed to fetch traces:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchTraceDetails = async (traceId: string) => {
    try {
      const response = await fetch(`/api/v1/observability/agent/traces/${traceId}`)
      const data = await response.json()
      setSelectedTrace(data)
    } catch (error) {
      console.error('Failed to fetch trace details:', error)
    }
  }

  return (
    <div className="grid grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Agent Execution Traces</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {traces.map((trace) => (
              <div
                key={trace.id}
                className="p-3 border rounded cursor-pointer hover:bg-gray-50"
                onClick={() => fetchTraceDetails(trace.id)}
              >
                <div className="font-medium">{trace.task}</div>
                <div className="text-sm text-gray-500">{trace.agent_id}</div>
                <div className="text-xs text-gray-400">{trace.status}</div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
      {selectedTrace && (
        <Card>
          <CardHeader>
            <CardTitle>Trace Details</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <div className="font-medium">Task</div>
                <div className="text-sm">{selectedTrace.task}</div>
              </div>
              <div>
                <div className="font-medium">Steps</div>
                <div className="space-y-2 mt-2">
                  {selectedTrace.steps.map((step) => (
                    <div key={step.id} className="p-2 bg-gray-50 rounded text-sm">
                      <div className="font-medium">Step {step.step_number}: {step.step_type}</div>
                      <div className="text-xs">{step.description}</div>
                      {step.tool_name && <div className="text-xs text-blue-600">Tool: {step.tool_name}</div>}
                      <div className="text-xs text-gray-500">{step.duration}ms</div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

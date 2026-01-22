'use client'

import { useState, useEffect, useRef } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { showToast } from '@/components/ui/Toast'
import { Badge } from '@/components/ui/Badge'

interface MetricLineageNode {
  id: string
  label: string
  type?: 'metric' | 'dataset' | 'table' | 'column'
}

interface MetricLineageEdge {
  from: string
  to: string
  type: string
  depth?: number
}

interface MetricLineageGraphProps {
  metricId: string
  maxDepth?: number
  direction?: 'upstream' | 'downstream' | 'both'
}

export default function MetricLineageGraph({
  metricId,
  maxDepth = 3,
  direction = 'both',
}: MetricLineageGraphProps) {
  const [lineage, setLineage] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const [selectedDepth, setSelectedDepth] = useState(maxDepth)
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    if (metricId) {
      fetchLineage()
    }
  }, [metricId, selectedDepth])

  const fetchLineage = async () => {
    setLoading(true)
    try {
      const response = await fetch(
        `/api/v1/metrics/${metricId}/lineage?max_depth=${selectedDepth}`
      )
      if (!response.ok) throw new Error('Failed to fetch lineage')
      const data = await response.json()
      setLineage(data)
      drawGraph(data)
    } catch (error: any) {
      showToast(error.message || 'Failed to fetch lineage', 'error')
    } finally {
      setLoading(false)
    }
  }

  const drawGraph = (lineageData: any) => {
    if (!canvasRef.current || !lineageData) return

    const canvas = canvasRef.current
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Set canvas size
    canvas.width = canvas.offsetWidth
    canvas.height = 600

    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height)

    const nodes = lineageData.nodes || []
    const edges = lineageData.edges || []

    if (nodes.length === 0) {
      ctx.fillStyle = '#666'
      ctx.font = '16px sans-serif'
      ctx.textAlign = 'center'
      ctx.fillText('No lineage data available', canvas.width / 2, canvas.height / 2)
      return
    }

    // Simple node layout (circular for now)
    const centerX = canvas.width / 2
    const centerY = canvas.height / 2
    const radius = Math.min(canvas.width, canvas.height) / 3

    const nodePositions = new Map<string, { x: number; y: number }>()
    nodes.forEach((node: string, index: number) => {
      const angle = (2 * Math.PI * index) / nodes.length
      const x = centerX + radius * Math.cos(angle)
      const y = centerY + radius * Math.sin(angle)
      nodePositions.set(node, { x, y })
    })

    // Draw edges first
    ctx.strokeStyle = '#888'
    ctx.lineWidth = 2
    edges.forEach((edge: MetricLineageEdge) => {
      const from = nodePositions.get(edge.from)
      const to = nodePositions.get(edge.to)
      if (from && to) {
        ctx.beginPath()
        ctx.moveTo(from.x, from.y)
        ctx.lineTo(to.x, to.y)
        ctx.stroke()

        // Draw arrow
        const angle = Math.atan2(to.y - from.y, to.x - from.x)
        const arrowLength = 10
        const arrowAngle = Math.PI / 6
        ctx.beginPath()
        ctx.moveTo(to.x, to.y)
        ctx.lineTo(
          to.x - arrowLength * Math.cos(angle - arrowAngle),
          to.y - arrowLength * Math.sin(angle - arrowAngle)
        )
        ctx.moveTo(to.x, to.y)
        ctx.lineTo(
          to.x - arrowLength * Math.cos(angle + arrowAngle),
          to.y - arrowLength * Math.sin(angle + arrowAngle)
        )
        ctx.stroke()

        // Draw edge label
        const midX = (from.x + to.x) / 2
        const midY = (from.y + to.y) / 2
        ctx.fillStyle = '#666'
        ctx.font = '12px sans-serif'
        ctx.fillText(edge.type || '', midX, midY - 5)
      }
    })

    // Draw nodes
    nodePositions.forEach((pos, nodeId) => {
      // Node circle
      ctx.beginPath()
      ctx.arc(pos.x, pos.y, 20, 0, 2 * Math.PI)
      ctx.fillStyle = nodeId === metricId ? '#3b82f6' : '#94a3b8'
      ctx.fill()
      ctx.strokeStyle = '#fff'
      ctx.lineWidth = 2
      ctx.stroke()

      // Node label
      ctx.fillStyle = '#000'
      ctx.font = '12px sans-serif'
      ctx.textAlign = 'center'
      const label = nodeId.substring(0, 8) + '...'
      ctx.fillText(label, pos.x, pos.y + 35)
    })
  }

  const getImpactAnalysis = async () => {
    try {
      const response = await fetch(`/api/v1/metrics/${metricId}/impact-analysis`)
      if (!response.ok) throw new Error('Failed to get impact analysis')
      const data = await response.json()
      showToast(
        `${data.affected_count || 0} metrics will be affected`,
        data.affected_count > 0 ? 'warning' : 'success'
      )
    } catch (error: any) {
      showToast(error.message || 'Failed to get impact analysis', 'error')
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Metric Lineage Graph</CardTitle>
            <CardDescription>Visual dependency graph for this metric</CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <select
              value={selectedDepth}
              onChange={(e) => setSelectedDepth(parseInt(e.target.value))}
              className="rounded-lg border border-border bg-background px-3 py-1 text-sm"
            >
              <option value={1}>Depth 1</option>
              <option value={2}>Depth 2</option>
              <option value={3}>Depth 3</option>
              <option value={5}>Depth 5</option>
            </select>
            <Button onClick={getImpactAnalysis} size="sm" variant="secondary">
              Impact Analysis
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="text-center py-8 text-muted-foreground">Loading lineage graph...</div>
        ) : (
          <div className="space-y-4">
            <div className="border rounded-lg overflow-hidden">
              <canvas
                ref={canvasRef}
                className="w-full"
                style={{ minHeight: '600px' }}
              />
            </div>

            {lineage && (
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h4 className="font-medium mb-2">Nodes ({lineage.nodes?.length || 0})</h4>
                  <div className="space-y-1 max-h-40 overflow-y-auto">
                    {(lineage.nodes || []).map((node: string) => (
                      <div key={node} className="text-sm text-muted-foreground">
                        <Badge variant="outline" className="mr-2">
                          {node === metricId ? 'Current' : 'Related'}
                        </Badge>
                        {node.substring(0, 36)}...
                      </div>
                    ))}
                  </div>
                </div>
                <div>
                  <h4 className="font-medium mb-2">Edges ({lineage.edges?.length || 0})</h4>
                  <div className="space-y-1 max-h-40 overflow-y-auto">
                    {(lineage.edges || []).map((edge: MetricLineageEdge, index: number) => (
                      <div key={index} className="text-sm text-muted-foreground">
                        <Badge variant="secondary" className="mr-2">
                          {edge.type}
                        </Badge>
                        Depth: {edge.depth || 0}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

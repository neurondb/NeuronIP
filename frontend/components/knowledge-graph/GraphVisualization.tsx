'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { cn } from '@/lib/utils/cn'

interface Node {
  id: string
  label: string
  type: string
  x?: number
  y?: number
}

interface Edge {
  from: string
  to: string
  label?: string
}

interface GraphVisualizationProps {
  nodes?: Node[]
  edges?: Edge[]
  selectedNodeId?: string
  onNodeClick?: (nodeId: string) => void
}

export default function GraphVisualization({
  nodes = [],
  edges = [],
  selectedNodeId,
  onNodeClick,
}: GraphVisualizationProps) {
  // Mock graph data for visualization
  const mockNodes: Node[] = nodes.length > 0 ? nodes : [
    { id: '1', label: 'Customer', type: 'entity', x: 100, y: 100 },
    { id: '2', label: 'Order', type: 'entity', x: 300, y: 100 },
    { id: '3', label: 'Product', type: 'entity', x: 500, y: 100 },
    { id: '4', label: 'Payment', type: 'entity', x: 300, y: 250 },
  ]

  const mockEdges: Edge[] = edges.length > 0 ? edges : [
    { from: '1', to: '2', label: 'places' },
    { from: '2', to: '3', label: 'contains' },
    { from: '2', to: '4', label: 'requires' },
  ]

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">Knowledge Graph</CardTitle>
        <CardDescription className="text-xs">Visualize entity relationships</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 min-h-0 overflow-auto">
        <div className="relative w-full h-full min-h-[400px] bg-muted/20 rounded-lg border border-border">
          <svg width="100%" height="100%" className="min-h-[400px]">
            {/* Render edges first */}
            {mockEdges.map((edge, index) => {
              const fromNode = mockNodes.find((n) => n.id === edge.from)
              const toNode = mockNodes.find((n) => n.id === edge.to)
              if (!fromNode || !toNode || !fromNode.x || !fromNode.y || !toNode.x || !toNode.y) return null

              return (
                <motion.line
                  key={`edge-${index}`}
                  x1={fromNode.x}
                  y1={fromNode.y}
                  x2={toNode.x}
                  y2={toNode.y}
                  stroke="hsl(var(--muted-foreground))"
                  strokeWidth="2"
                  initial={{ pathLength: 0 }}
                  animate={{ pathLength: 1 }}
                  transition={{ delay: index * 0.1 }}
                >
                  {edge.label && (
                    <title>{edge.label}</title>
                  )}
                </motion.line>
              )
            })}

            {/* Render nodes */}
            {mockNodes.map((node, index) => {
              if (!node.x || !node.y) return null

              const isSelected = selectedNodeId === node.id

              return (
                <motion.g
                  key={node.id}
                  initial={{ opacity: 0, scale: 0 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: index * 0.1 }}
                >
                  <circle
                    cx={node.x}
                    cy={node.y}
                    r={20}
                    fill={isSelected ? 'hsl(var(--primary))' : 'hsl(var(--card))'}
                    stroke={isSelected ? 'hsl(var(--primary))' : 'hsl(var(--border))'}
                    strokeWidth={isSelected ? 3 : 2}
                    className={cn('cursor-pointer hover:opacity-80 transition-opacity')}
                    onClick={() => onNodeClick?.(node.id)}
                  />
                  <text
                    x={node.x}
                    y={node.y + 40}
                    textAnchor="middle"
                    className="text-xs fill-foreground"
                  >
                    {node.label}
                  </text>
                </motion.g>
              )
            })}
          </svg>
        </div>
        <p className="text-xs text-muted-foreground mt-2 text-center">
          {mockNodes.length} entities, {mockEdges.length} relationships
        </p>
      </CardContent>
    </Card>
  )
}

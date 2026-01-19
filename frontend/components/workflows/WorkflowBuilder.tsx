'use client'

import { useState, useCallback } from 'react'
import ReactFlow, {
  Node,
  Edge,
  addEdge,
  Connection,
  useNodesState,
  useEdgesState,
  Controls,
  Background,
  MiniMap,
  NodeTypes,
} from 'reactflow'
import 'reactflow/dist/style.css'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import WorkflowNodePalette from './WorkflowNodePalette'

interface WorkflowNode extends Node {
  data: {
    label: string
    type: 'agent' | 'script' | 'condition' | 'parallel'
    config?: Record<string, any>
  }
}

const nodeTypes: NodeTypes = {
  agent: ({ data }) => (
    <div className="px-4 py-2 bg-blue-100 border-2 border-blue-500 rounded-lg">
      <div className="font-semibold">{data.label}</div>
      <div className="text-xs text-gray-600">{data.type}</div>
    </div>
  ),
  script: ({ data }) => (
    <div className="px-4 py-2 bg-green-100 border-2 border-green-500 rounded-lg">
      <div className="font-semibold">{data.label}</div>
      <div className="text-xs text-gray-600">{data.type}</div>
    </div>
  ),
  condition: ({ data }) => (
    <div className="px-4 py-2 bg-yellow-100 border-2 border-yellow-500 rounded-lg">
      <div className="font-semibold">{data.label}</div>
      <div className="text-xs text-gray-600">{data.type}</div>
    </div>
  ),
  parallel: ({ data }) => (
    <div className="px-4 py-2 bg-purple-100 border-2 border-purple-500 rounded-lg">
      <div className="font-semibold">{data.label}</div>
      <div className="text-xs text-gray-600">{data.type}</div>
    </div>
  ),
}

export default function WorkflowBuilder() {
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [selectedNode, setSelectedNode] = useState<WorkflowNode | null>(null)

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  )

  const addNode = useCallback((type: 'agent' | 'script' | 'condition' | 'parallel') => {
    const newNode: WorkflowNode = {
      id: `node-${Date.now()}`,
      type: type,
      position: { x: Math.random() * 400, y: Math.random() * 400 },
      data: {
        label: `${type.charAt(0).toUpperCase() + type.slice(1)} Node`,
        type: type,
      },
    }
    setNodes((nds) => [...nds, newNode])
  }, [setNodes])

  const saveWorkflow = useCallback(async () => {
    const workflow = {
      steps: nodes.map(node => ({
        id: node.id,
        name: node.data.label,
        type: node.data.type,
        config: node.data.config || {},
        next_steps: edges
          .filter(e => e.source === node.id)
          .map(e => e.target),
      })),
      start_step: nodes[0]?.id || '',
    }
    
    // Save workflow via API
    console.log('Saving workflow:', workflow)
  }, [nodes, edges])

  return (
    <div className="h-full flex flex-col">
      <Card className="flex-1 flex flex-col min-h-0">
        <CardHeader>
          <CardTitle>Workflow Builder</CardTitle>
        </CardHeader>
        <CardContent className="flex-1 flex gap-4 min-h-0">
          <div className="w-64 flex-shrink-0">
            <WorkflowNodePalette onAddNode={addNode} />
          </div>
          <div className="flex-1 min-h-0">
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onConnect={onConnect}
              onNodeClick={(_, node) => setSelectedNode(node as WorkflowNode)}
              nodeTypes={nodeTypes}
              fitView
            >
              <Controls />
              <Background />
              <MiniMap />
            </ReactFlow>
          </div>
          {selectedNode && (
            <div className="w-80 flex-shrink-0 border-l border-border p-4">
              <h3 className="font-semibold mb-2">Node Configuration</h3>
              <div className="space-y-2">
                <div>
                  <label className="text-sm font-medium">Label</label>
                  <input
                    type="text"
                    value={selectedNode.data.label}
                    onChange={(e) => {
                      setNodes((nds) =>
                        nds.map((n) =>
                          n.id === selectedNode.id
                            ? { ...n, data: { ...n.data, label: e.target.value } }
                            : n
                        )
                      )
                    }}
                    className="w-full rounded border border-border px-2 py-1"
                  />
                </div>
                {/* Add more configuration fields based on node type */}
              </div>
            </div>
          )}
        </CardContent>
        <div className="p-4 border-t border-border flex justify-end gap-2">
          <Button onClick={saveWorkflow}>Save Workflow</Button>
        </div>
      </Card>
    </div>
  )
}

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
import Tooltip from '@/components/ui/Tooltip'
import HelpText from '@/components/ui/HelpText'
import { InformationCircleIcon, BookmarkIcon } from '@heroicons/react/24/outline'
import WorkflowNodePalette from './WorkflowNodePalette'

interface WorkflowNode extends Node {
  data: {
    label: string
    type: 'agent' | 'script' | 'condition' | 'parallel' | 'approval' | 'retry'
    config?: Record<string, any>
    condition?: {
      expression: string
      cases: Array<{ value: string; nextStep: string }>
      default: string
    }
    approval?: {
      approverId: string
      timeout?: number
    }
    retry?: {
      maxRetries: number
      backoffMs: number
      failurePath?: string
    }
  }
}

const nodeTypes: NodeTypes = {
  agent: ({ data }) => (
    <Tooltip
      content={
        <div>
          <p className="font-medium mb-1">Agent Node</p>
          <p className="text-xs">
            Executes AI agent tasks with long-term memory. Configure the agent, prompt, and tools to use.
          </p>
        </div>
      }
      variant="info"
    >
      <div className="px-4 py-2 bg-blue-100 border-2 border-blue-500 rounded-lg cursor-pointer">
        <div className="font-semibold">{data.label}</div>
        <div className="text-xs text-gray-600">{data.type}</div>
      </div>
    </Tooltip>
  ),
  script: ({ data }) => (
    <Tooltip
      content={
        <div>
          <p className="font-medium mb-1">Script Node</p>
          <p className="text-xs">
            Runs custom scripts or transformations. Use for data processing, API calls, or custom logic.
          </p>
        </div>
      }
      variant="info"
    >
      <div className="px-4 py-2 bg-green-100 border-2 border-green-500 rounded-lg cursor-pointer">
        <div className="font-semibold">{data.label}</div>
        <div className="text-xs text-gray-600">{data.type}</div>
      </div>
    </Tooltip>
  ),
  condition: ({ data }) => (
    <Tooltip
      content={
        <div>
          <p className="font-medium mb-1">Condition Node</p>
          <p className="text-xs">
            Branches workflow based on conditions. Configure the condition logic and next steps.
          </p>
        </div>
      }
      variant="info"
    >
      <div className="px-4 py-2 bg-yellow-100 border-2 border-yellow-500 rounded-lg cursor-pointer">
        <div className="font-semibold">{data.label}</div>
        <div className="text-xs text-gray-600">{data.type}</div>
      </div>
    </Tooltip>
  ),
  parallel: ({ data }) => (
    <Tooltip
      content={
        <div>
          <p className="font-medium mb-1">Parallel Node</p>
          <p className="text-xs">
            Runs multiple steps in parallel. Use for independent operations that can run simultaneously.
          </p>
        </div>
      }
      variant="info"
    >
      <div className="px-4 py-2 bg-purple-100 border-2 border-purple-500 rounded-lg cursor-pointer">
        <div className="font-semibold">{data.label}</div>
        <div className="text-xs text-gray-600">{data.type}</div>
      </div>
    </Tooltip>
  ),
  approval: ({ data }) => (
    <Tooltip
      content={
        <div>
          <p className="font-medium mb-1">Approval Node</p>
          <p className="text-xs">
            Human approval step. Pauses workflow until approved or rejected by a user.
          </p>
        </div>
      }
      variant="info"
    >
      <div className="px-4 py-2 bg-orange-100 border-2 border-orange-500 rounded-lg cursor-pointer">
        <div className="font-semibold">{data.label}</div>
        <div className="text-xs text-gray-600">{data.type}</div>
      </div>
    </Tooltip>
  ),
  retry: ({ data }) => (
    <Tooltip
      content={
        <div>
          <p className="font-medium mb-1">Retry Node</p>
          <p className="text-xs">
            Retry logic with exponential backoff. Configure max retries and failure path.
          </p>
        </div>
      }
      variant="info"
    >
      <div className="px-4 py-2 bg-red-100 border-2 border-red-500 rounded-lg cursor-pointer">
        <div className="font-semibold">{data.label}</div>
        <div className="text-xs text-gray-600">{data.type}</div>
      </div>
    </Tooltip>
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

  const addNode = useCallback((type: 'agent' | 'script' | 'condition' | 'parallel' | 'approval' | 'retry') => {
    const newNode: WorkflowNode = {
      id: `node-${Date.now()}`,
      type: type,
      position: { x: Math.random() * 400, y: Math.random() * 400 },
      data: {
        label: `${type.charAt(0).toUpperCase() + type.slice(1)} Node`,
        type: type,
        config: {},
      },
    }
    
    // Add default config based on type
    if (type === 'approval') {
      newNode.data.approval = {
        approverId: '',
        timeout: 3600, // 1 hour default
      }
    } else if (type === 'retry') {
      newNode.data.retry = {
        maxRetries: 3,
        backoffMs: 1000,
      }
    } else if (type === 'condition') {
      newNode.data.condition = {
        expression: '',
        cases: [],
        default: '',
      }
    }
    
    setNodes((nds) => [...nds, newNode])
  }, [setNodes])

  const saveWorkflow = useCallback(async () => {
    const workflowDefinition = {
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
    
    const workflow = {
      name: `Workflow ${Date.now()}`,
      workflow_definition: workflowDefinition,
      enabled: true,
    }
    
    try {
      const response = await fetch('/api/v1/workflows', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(workflow),
      })
      
      if (response.ok) {
        const data = await response.json()
        alert(`Workflow saved successfully! ID: ${data.id}`)
      } else {
        alert('Failed to save workflow')
      }
    } catch (error) {
      console.error('Error saving workflow:', error)
      alert('Error saving workflow')
    }
  }, [nodes, edges])

  return (
    <div className="h-full flex flex-col">
      <Card className="flex-1 flex flex-col min-h-0">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CardTitle>Workflow Builder</CardTitle>
              <Tooltip
                content={
                  <div>
                    <p className="font-medium mb-1">Workflow Builder Help</p>
                    <p className="text-xs mb-2">
                      Build multi-step workflows with AI agents. Drag and drop nodes to create your workflow DAG.
                    </p>
                    <ul className="text-xs space-y-1">
                      <li>• Agent nodes: Execute AI agent tasks</li>
                      <li>• Script nodes: Run custom scripts</li>
                      <li>• Condition nodes: Branch based on conditions</li>
                      <li>• Parallel nodes: Run steps in parallel</li>
                    </ul>
                  </div>
                }
                variant="info"
              >
                <InformationCircleIcon className="h-4 w-4 text-muted-foreground cursor-help" />
              </Tooltip>
            </div>
            <HelpText
              variant="tooltip"
              content={
                <div>
                  <p className="font-medium mb-1">Workflow Creation</p>
                  <p className="text-xs mb-2">
                    Workflows allow you to automate complex multi-step processes with AI agents.
                  </p>
                  <p className="text-xs">
                    <a
                      href="/docs/features/agent-workflows"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="underline"
                    >
                      Learn more →
                    </a>
                  </p>
                </div>
              }
            />
          </div>
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
            <div className="w-80 flex-shrink-0 border-l border-border p-4 overflow-y-auto">
              <h3 className="font-semibold mb-4">Node Configuration</h3>
              <div className="space-y-4">
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
                    className="w-full rounded border border-border px-2 py-1 mt-1"
                  />
                </div>

                {/* Approval Configuration */}
                {selectedNode.data.type === 'approval' && selectedNode.data.approval && (
                  <div className="space-y-2">
                    <div>
                      <label className="text-sm font-medium">Approver ID</label>
                      <input
                        type="text"
                        value={selectedNode.data.approval.approverId}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      approval: {
                                        ...n.data.approval!,
                                        approverId: e.target.value,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                        placeholder="user@example.com"
                      />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Timeout (seconds)</label>
                      <input
                        type="number"
                        value={selectedNode.data.approval.timeout}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      approval: {
                                        ...n.data.approval!,
                                        timeout: parseInt(e.target.value) || 3600,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                      />
                    </div>
                  </div>
                )}

                {/* Retry Configuration */}
                {selectedNode.data.type === 'retry' && selectedNode.data.retry && (
                  <div className="space-y-2">
                    <div>
                      <label className="text-sm font-medium">Max Retries</label>
                      <input
                        type="number"
                        value={selectedNode.data.retry.maxRetries}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      retry: {
                                        ...n.data.retry!,
                                        maxRetries: parseInt(e.target.value) || 3,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                        min="0"
                      />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Backoff (ms)</label>
                      <input
                        type="number"
                        value={selectedNode.data.retry.backoffMs}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      retry: {
                                        ...n.data.retry!,
                                        backoffMs: parseInt(e.target.value) || 1000,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                        min="0"
                      />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Failure Path (optional)</label>
                      <input
                        type="text"
                        value={selectedNode.data.retry.failurePath || ''}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      retry: {
                                        ...n.data.retry!,
                                        failurePath: e.target.value || undefined,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                        placeholder="node-id"
                      />
                    </div>
                  </div>
                )}

                {/* Condition Configuration */}
                {selectedNode.data.type === 'condition' && selectedNode.data.condition && (
                  <div className="space-y-2">
                    <div>
                      <label className="text-sm font-medium">Condition Expression</label>
                      <input
                        type="text"
                        value={selectedNode.data.condition.expression}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      condition: {
                                        ...n.data.condition!,
                                        expression: e.target.value,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                        placeholder="status == 'success'"
                      />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Default Next Step</label>
                      <input
                        type="text"
                        value={selectedNode.data.condition.default}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      condition: {
                                        ...n.data.condition!,
                                        default: e.target.value,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                        placeholder="node-id"
                      />
                    </div>
                  </div>
                )}

                {/* Agent Configuration */}
                {selectedNode.data.type === 'agent' && (
                  <div className="space-y-2">
                    <div>
                      <label className="text-sm font-medium">Agent ID</label>
                      <input
                        type="text"
                        value={selectedNode.data.config?.agent_id || ''}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      config: {
                                        ...n.data.config,
                                        agent_id: e.target.value,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                        placeholder="agent-uuid"
                      />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Task</label>
                      <textarea
                        value={selectedNode.data.config?.task || ''}
                        onChange={(e) => {
                          setNodes((nds) =>
                            nds.map((n) =>
                              n.id === selectedNode.id
                                ? {
                                    ...n,
                                    data: {
                                      ...n.data,
                                      config: {
                                        ...n.data.config,
                                        task: e.target.value,
                                      },
                                    },
                                  }
                                : n
                            )
                          )
                        }}
                        className="w-full rounded border border-border px-2 py-1 mt-1"
                        rows={3}
                        placeholder="Task description..."
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </CardContent>
        <div className="p-4 border-t border-border flex justify-end gap-2">
          <Button onClick={saveWorkflow} icon={<BookmarkIcon className="h-4 w-4" />}>
            Save Workflow
          </Button>
        </div>
      </Card>
    </div>
  )
}

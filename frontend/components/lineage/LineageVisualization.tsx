'use client'

import { useMemo, useState } from 'react'
import ReactFlow, { Node, Edge, Background, Controls, MiniMap, ConnectionMode, MarkerType } from 'reactflow'
import 'reactflow/dist/style.css'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useLineageGraph, useLineageImpact, useColumnLineage } from '@/lib/api/queries'
import { ArrowsRightLeftIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline'

interface LineageVisualizationProps {
  resourceType?: string
  resourceId?: string
}

type LineageViewMode = 'table' | 'column'

export default function LineageVisualization({ resourceType, resourceId }: LineageVisualizationProps) {
  const [selectedResourceId, setSelectedResourceId] = useState<string | null>(resourceId || null)
  const [viewMode, setViewMode] = useState<LineageViewMode>('table')
  const [columnLineageParams, setColumnLineageParams] = useState<{
    connectorId: string
    schemaName: string
    tableName: string
    columnName: string
  } | null>(null)

  const { data: graphData, isLoading } = useLineageGraph(!selectedResourceId && viewMode === 'table')
  const { data: impactData } = useLineageImpact(selectedResourceId || '', !!selectedResourceId && viewMode === 'table')
  const { data: columnLineageData, isLoading: isLoadingColumn } = useColumnLineage(
    columnLineageParams?.connectorId || '',
    columnLineageParams?.schemaName || '',
    columnLineageParams?.tableName || '',
    columnLineageParams?.columnName || '',
    viewMode === 'column' && !!columnLineageParams
  )

  const { nodes, edges } = useMemo(() => {
    if (viewMode === 'column') {
      if (!columnLineageData) {
        return { nodes: [], edges: [] }
      }

      const columnNodes: Node[] = (columnLineageData.nodes || []).map((node: any, index: number) => ({
        id: node.id || `col-node-${index}`,
        type: 'default',
        position: {
          x: (index % 5) * 200 + 100,
          y: Math.floor(index / 5) * 150 + 100,
        },
        data: {
          label: `${node.schema_name}.${node.table_name}.${node.column_name}`,
          type: node.node_type,
          ...node,
        },
        style: {
          background: node.node_type === 'source' ? '#3b82f6' : node.node_type === 'derived' ? '#10b981' : '#8b5cf6',
          color: '#fff',
          border: '2px solid #fff',
          borderRadius: '8px',
          padding: '10px',
          width: 180,
          fontSize: 11,
        },
      }))

      const columnEdges: Edge[] = (columnLineageData.edges || []).map((edge: any) => ({
        id: edge.id || `col-edge-${edge.source_node_id}-${edge.target_node_id}`,
        source: edge.source_node_id,
        target: edge.target_node_id,
        label: edge.edge_type || '',
        type: 'smoothstep',
        animated: false,
        markerEnd: {
          type: MarkerType.ArrowClosed,
        },
        style: {
          strokeWidth: 2,
          stroke: '#64748b',
        },
      }))

      return { nodes: columnNodes, edges: columnEdges }
    }

    // Table-level lineage
    if (!graphData) {
      return { nodes: [], edges: [] }
    }

    const lineageNodes: Node[] = (graphData.nodes || []).map((node: any, index: number) => ({
      id: node.id || `node-${index}`,
      type: 'default',
      position: {
        x: (index % 5) * 200 + 100,
        y: Math.floor(index / 5) * 150 + 100,
      },
      data: {
        label: node.node_name || node.id,
        type: node.node_type,
        ...node,
      },
      style: {
        background: node.node_type === 'dataset' ? '#3b82f6' : node.node_type === 'transformation' ? '#10b981' : '#8b5cf6',
        color: '#fff',
        border: '2px solid #fff',
        borderRadius: '8px',
        padding: '10px',
        width: 150,
        fontSize: 12,
      },
    }))

    const lineageEdges: Edge[] = (graphData.edges || []).map((edge: any) => ({
      id: edge.id || `edge-${edge.source_node_id}-${edge.target_node_id}`,
      source: edge.source_node_id,
      target: edge.target_node_id,
      label: edge.edge_type || '',
      type: 'smoothstep',
      animated: false,
      markerEnd: {
        type: MarkerType.ArrowClosed,
      },
      style: {
        strokeWidth: 2,
        stroke: '#64748b',
      },
    }))

    return { nodes: lineageNodes, edges: lineageEdges }
  }, [graphData, columnLineageData, viewMode])

  if (isLoading || isLoadingColumn) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <p className="text-muted-foreground">Loading lineage graph...</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4 h-full flex flex-col">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Data Lineage Graph</CardTitle>
            <div className="flex gap-2 items-center">
              <div className="flex gap-1 border border-border rounded-lg p-1">
                <Button
                  variant={viewMode === 'table' ? 'primary' : 'ghost'}
                  size="sm"
                  onClick={() => {
                    setViewMode('table')
                    setColumnLineageParams(null)
                  }}
                >
                  Table Level
                </Button>
                <Button
                  variant={viewMode === 'column' ? 'primary' : 'ghost'}
                  size="sm"
                  onClick={() => setViewMode('column')}
                >
                  Column Level
                </Button>
              </div>
              {viewMode === 'table' ? (
                <>
                  <input
                    type="text"
                    placeholder="Resource ID for impact analysis"
                    value={selectedResourceId || ''}
                    onChange={(e) => setSelectedResourceId(e.target.value || null)}
                    className="rounded-lg border border-border bg-background px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setSelectedResourceId(null)}
                    disabled={!selectedResourceId}
                  >
                    Show Full Graph
                  </Button>
                </>
              ) : (
                <div className="flex gap-2">
                  <input
                    type="text"
                    placeholder="Connector ID"
                    value={columnLineageParams?.connectorId || ''}
                    onChange={(e) =>
                      setColumnLineageParams((prev) => ({
                        ...prev,
                        connectorId: e.target.value,
                        schemaName: prev?.schemaName || '',
                        tableName: prev?.tableName || '',
                        columnName: prev?.columnName || '',
                      }))
                    }
                    className="rounded-lg border border-border bg-background px-3 py-1.5 text-sm w-32 focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                  <input
                    type="text"
                    placeholder="Schema"
                    value={columnLineageParams?.schemaName || ''}
                    onChange={(e) =>
                      setColumnLineageParams((prev) => ({
                        ...prev,
                        schemaName: e.target.value,
                        connectorId: prev?.connectorId || '',
                        tableName: prev?.tableName || '',
                        columnName: prev?.columnName || '',
                      }))
                    }
                    className="rounded-lg border border-border bg-background px-3 py-1.5 text-sm w-24 focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                  <input
                    type="text"
                    placeholder="Table"
                    value={columnLineageParams?.tableName || ''}
                    onChange={(e) =>
                      setColumnLineageParams((prev) => ({
                        ...prev,
                        tableName: e.target.value,
                        connectorId: prev?.connectorId || '',
                        schemaName: prev?.schemaName || '',
                        columnName: prev?.columnName || '',
                      }))
                    }
                    className="rounded-lg border border-border bg-background px-3 py-1.5 text-sm w-24 focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                  <input
                    type="text"
                    placeholder="Column"
                    value={columnLineageParams?.columnName || ''}
                    onChange={(e) =>
                      setColumnLineageParams((prev) => ({
                        ...prev,
                        columnName: e.target.value,
                        connectorId: prev?.connectorId || '',
                        schemaName: prev?.schemaName || '',
                        tableName: prev?.tableName || '',
                      }))
                    }
                    className="rounded-lg border border-border bg-background px-3 py-1.5 text-sm w-24 focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                </div>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-[600px] border border-border rounded-lg">
            <ReactFlow
              nodes={nodes}
              edges={edges}
              connectionMode={ConnectionMode.Loose}
              fitView
              className="bg-background"
            >
              <Background />
              <Controls />
              <MiniMap />
            </ReactFlow>
          </div>
        </CardContent>
      </Card>

      {selectedResourceId && impactData && (
        <Card>
          <CardHeader>
            <CardTitle>Impact Analysis</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">
                {impactData.count || 0} resources will be impacted by changes to this resource
              </p>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                {(impactData.impacted_nodes || []).slice(0, 9).map((node: any) => (
                  <div
                    key={node.id || node}
                    className="p-2 border border-border rounded-lg text-sm bg-muted/20"
                  >
                    {typeof node === 'string' ? node : node.node_name || node.id}
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
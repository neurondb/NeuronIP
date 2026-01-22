'use client'

import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import apiClient from '@/lib/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { Textarea } from '@/components/ui/Textarea'
import { Badge } from '@/components/ui/Badge'
import { showToast } from '@/components/ui/Toast'
import GraphVisualization from './GraphVisualization'

interface GraphQueryResult {
  nodes: Array<{
    id: string
    entity_name: string
    entity_type?: string
    entity_type_id?: string
    metadata?: Record<string, any>
  }>
  edges: Array<{
    id: string
    source: string
    target: string
    relationship_type: string
    strength?: number
  }>
}

const exampleQueries = [
  {
    name: 'Find all entities',
    query: 'MATCH (n) RETURN n',
  },
  {
    name: 'Find relationships',
    query: 'MATCH (a)-[r]->(b) RETURN a, r, b',
  },
  {
    name: 'Find connected entities',
    query: 'MATCH (a)-[*1..3]-(b) WHERE a.entity_name = "Customer" RETURN a, b',
  },
  {
    name: 'Find by type',
    query: 'MATCH (n) WHERE n.entity_type = "Person" RETURN n',
  },
]

export default function GraphQueryUI() {
  const [query, setQuery] = useState('MATCH (n) RETURN n LIMIT 10')
  const [selectedExample, setSelectedExample] = useState<string | null>(null)

  // Execute graph query
  const { data: queryResult, isLoading, refetch } = useQuery({
    queryKey: ['graph-query', query],
    queryFn: async () => {
      const response = await apiClient.post('/api/v1/knowledge-graph/query', {
        query: query,
      })
      return response.data as GraphQueryResult
    },
    enabled: false, // Don't auto-execute
  })

  const executeQuery = () => {
    if (!query.trim()) {
      showToast('Please enter a query', 'error')
      return
    }
    refetch()
  }

  const useExample = (exampleQuery: string) => {
    setQuery(exampleQuery)
    setSelectedExample(exampleQuery)
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Graph Query Builder</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Example Queries */}
          <div>
            <label className="text-sm font-medium mb-2 block">Example Queries</label>
            <div className="flex flex-wrap gap-2">
              {exampleQueries.map((example, idx) => (
                <Button
                  key={idx}
                  onClick={() => useExample(example.query)}
                  variant={selectedExample === example.query ? 'primary' : 'outline'}
                  size="sm"
                >
                  {example.name}
                </Button>
              ))}
            </div>
          </div>

          {/* Query Input */}
          <div>
            <label className="text-sm font-medium mb-2 block">Query</label>
            <Textarea
              value={query}
              onChange={(e) => {
                setQuery(e.target.value)
                setSelectedExample(null)
              }}
              className="font-mono text-sm"
              rows={6}
              placeholder="MATCH (n) RETURN n LIMIT 10"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Supports Cypher-like syntax: MATCH, WHERE, RETURN, LIMIT
            </p>
          </div>

          {/* Execute Button */}
          <Button onClick={executeQuery} disabled={isLoading} className="w-full">
            {isLoading ? 'Executing...' : 'Execute Query'}
          </Button>
        </CardContent>
      </Card>

      {/* Results */}
      {queryResult && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Graph Visualization */}
          <Card>
            <CardHeader>
              <CardTitle>Graph Visualization</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-96">
                <GraphVisualization
                  nodes={queryResult.nodes.map((n) => ({
                    id: n.id,
                    label: n.entity_name,
                    type: n.entity_type || n.entity_type_id || 'unknown',
                    metadata: n.metadata,
                  }))}
                  edges={queryResult.edges.map((e) => ({
                    from: e.source,
                    to: e.target,
                    label: e.relationship_type,
                  }))}
                />
              </div>
            </CardContent>
          </Card>

          {/* Results Table */}
          <Card>
            <CardHeader>
              <CardTitle>Query Results</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium mb-2">Nodes ({queryResult.nodes.length})</h4>
                  <div className="space-y-2 max-h-48 overflow-y-auto">
                    {queryResult.nodes.map((node) => (
                      <div key={node.id} className="p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{node.entity_name}</span>
                          {node.entity_type && (
                            <Badge variant="outline" className="text-xs">
                              {node.entity_type}
                            </Badge>
                          )}
                        </div>
                        {node.metadata && Object.keys(node.metadata).length > 0 && (
                          <div className="text-xs text-muted-foreground mt-1">
                            {JSON.stringify(node.metadata, null, 2)}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h4 className="font-medium mb-2">Edges ({queryResult.edges.length})</h4>
                  <div className="space-y-2 max-h-48 overflow-y-auto">
                    {queryResult.edges.map((edge) => (
                      <div key={edge.id} className="p-2 border rounded">
                        <div className="flex items-center gap-2">
                          <span className="text-sm">
                            {edge.source} → {edge.target}
                          </span>
                          <Badge variant="outline" className="text-xs">
                            {edge.relationship_type}
                          </Badge>
                          {edge.strength !== undefined && (
                            <span className="text-xs text-muted-foreground">
                              ({edge.strength.toFixed(2)})
                            </span>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {queryResult && queryResult.nodes.length === 0 && queryResult.edges.length === 0 && (
        <Card>
          <CardContent className="py-8 text-center text-muted-foreground">
            No results found
          </CardContent>
        </Card>
      )}
    </div>
  )
}

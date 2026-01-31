'use client'

import { useState } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Input from '@/components/ui/Input'

export default function QueryRewriter() {
  const [query, setQuery] = useState('')
  const [rewrittenQuery, setRewrittenQuery] = useState('')
  const [semantics, setSemantics] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  const handleRewrite = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/warehouse/query/rewrite', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query, query_type: 'sql' }),
      })
      const data = await response.json()
      setRewrittenQuery(data.rewritten_query)
      setSemantics(data.semantics)
    } catch (error) {
      console.error('Failed to rewrite query:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Query Rewriter</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-2">Original Query</label>
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Enter SQL query or natural language query"
            className="font-mono"
          />
        </div>
        <Button onClick={handleRewrite} disabled={loading || !query}>
          Rewrite with Business Semantics
        </Button>
        {rewrittenQuery && (
          <div>
            <label className="block text-sm font-medium mb-2">Rewritten Query</label>
            <pre className="bg-gray-100 p-3 rounded text-sm overflow-x-auto">
              {rewrittenQuery}
            </pre>
          </div>
        )}
        {semantics && (
          <div>
            <h4 className="font-medium mb-2">Matched Business Terms</h4>
            <div className="space-y-2">
              {semantics.matched_metrics?.map((metric: any, i: number) => (
                <div key={i} className="text-sm p-2 bg-blue-50 rounded">
                  <strong>{metric.display_name}</strong>: {metric.sql}
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

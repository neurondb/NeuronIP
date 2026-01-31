'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'

import PageTemplate from '@/components/layout/PageTemplate'
import Button from '@/components/ui/Button'
import { showToast } from '@/components/ui/Toast'
import ChartVisualization from '@/components/warehouse/ChartVisualization'
import QueryEditor from '@/components/warehouse/QueryEditor'
import QueryHistory from '@/components/warehouse/QueryHistory'
import QueryInsights from '@/components/warehouse/QueryInsights'
import ResultsTable from '@/components/warehouse/ResultsTable'
import SchemaExplorer from '@/components/warehouse/SchemaExplorer'
import { slideUp } from '@/lib/animations/variants'
import { useWarehouseQuery } from '@/lib/api/queries'
import { microcopy } from '@/lib/copy/microcopy'

export default function WarehousePage() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<unknown[]>([])
  const [queryResponse, setQueryResponse] = useState<Record<string, unknown> | null>(null)
  const [activeTab, setActiveTab] = useState<'query' | 'history' | 'insights'>('query')
  const { mutate: executeQuery, isPending } = useWarehouseQuery()

  const handleExecute = () => {
    if (!query.trim()) {
      showToast('Ask a question or write a query to get started', 'warning')
      return
    }
    executeQuery(
      { query: query.trim() },
      {
        onSuccess: (data) => {
          setResults(data.results || data.rows || [])
          setQueryResponse(data)
          showToast('Got your results!', 'success')
        },
        onError: (error: unknown) => {
          const err = error as { response?: { data?: { message?: string } } }
          showToast(err?.response?.data?.message || microcopy.warehouse.error, 'error')
        },
      }
    )
  }

  const handleSelectHistoryQuery = (selectedQuery: string) => {
    setQuery(selectedQuery)
    setActiveTab('query')
  }

  const filterRow = (
    <div className="flex gap-2 border-b border-border">
      <Button
        variant={activeTab === 'query' ? 'primary' : 'ghost'}
        onClick={() => setActiveTab('query')}
        size="sm"
      >
        {microcopy.warehouse.queryTab}
      </Button>
      <Button
        variant={activeTab === 'history' ? 'primary' : 'ghost'}
        onClick={() => setActiveTab('history')}
        size="sm"
      >
        {microcopy.warehouse.historyTab}
      </Button>
      <Button
        variant={activeTab === 'insights' ? 'primary' : 'ghost'}
        onClick={() => setActiveTab('insights')}
        size="sm"
      >
        Insights
      </Button>
    </div>
  )

  return (
    <PageTemplate
      title={microcopy.warehouse.title}
      description={microcopy.warehouse.subtitle}
      archetype="search"
      filterRow={filterRow}
    >
      {activeTab === 'query' && (
        <div className="grid gap-3 sm:gap-4 lg:grid-cols-3 flex-1 min-h-0">
          <motion.div variants={slideUp} className="lg:col-span-1 flex flex-col min-h-0">
            <SchemaExplorer />
          </motion.div>
          <motion.div variants={slideUp} className="lg:col-span-2 flex flex-col min-h-0 space-y-3 sm:space-y-4">
            <div className="flex-shrink-0">
              <QueryEditor value={query} onChange={setQuery} onExecute={handleExecute} />
            </div>
            {queryResponse?.chart_config != null && queryResponse?.chart_type != null && (
              <div className="flex-shrink-0">
                <ChartVisualization
                  chartType={queryResponse.chart_type as string}
                  chartConfig={queryResponse.chart_config as Record<string, unknown>}
                  data={results as unknown[]}
                />
              </div>
            )}
            <div className="flex-1 min-h-0">
              <ResultsTable data={results} isLoading={isPending} />
            </div>
          </motion.div>
        </div>
      )}

      {activeTab === 'history' && (
        <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
          <QueryHistory onSelectQuery={handleSelectHistoryQuery} />
        </motion.div>
      )}

      {activeTab === 'insights' && (
        <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
          {queryResponse?.sql != null ? (
            <QueryInsights sql={queryResponse.sql as string} />
          ) : (
            <p className="text-sm text-muted-foreground py-8 text-center">Run a query to see insights.</p>
          )}
        </motion.div>
      )}
    </PageTemplate>
  )
}

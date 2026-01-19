'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import QueryEditor from '@/components/warehouse/QueryEditor'
import ResultsTable from '@/components/warehouse/ResultsTable'
import SchemaExplorer from '@/components/warehouse/SchemaExplorer'
import { useWarehouseQuery } from '@/lib/api/queries'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import { showToast } from '@/components/ui/Toast'

export default function WarehousePage() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<unknown[]>([])
  const { mutate: executeQuery, isPending } = useWarehouseQuery()

  const handleExecute = () => {
    if (!query.trim()) {
      showToast('Please enter a query', 'warning')
      return
    }

    executeQuery(
      {
        query: query.trim(),
      },
      {
        onSuccess: (data) => {
          setResults(data.results || data.rows || [])
          showToast('Query executed successfully', 'success')
        },
        onError: (error: any) => {
          showToast(
            error?.response?.data?.message || 'Query execution failed',
            'error'
          )
        },
      }
    )
  }

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Warehouse</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Query your data warehouse with natural language or SQL
        </p>
      </motion.div>

      {/* Main Content Grid */}
      <div className="grid gap-3 sm:gap-4 lg:grid-cols-3 flex-1 min-h-0">
        {/* Left: Schema Explorer */}
        <motion.div variants={slideUp} className="lg:col-span-1 flex flex-col min-h-0">
          <SchemaExplorer />
        </motion.div>

        {/* Right: Query Editor and Results */}
        <motion.div variants={slideUp} className="lg:col-span-2 flex flex-col min-h-0 space-y-3 sm:space-y-4">
          <div className="flex-shrink-0">
            <QueryEditor
              value={query}
              onChange={setQuery}
              onExecute={handleExecute}
            />
          </div>
          <div className="flex-1 min-h-0">
            <ResultsTable data={results} isLoading={isPending} />
          </div>
        </motion.div>
      </div>
    </motion.div>
  )
}

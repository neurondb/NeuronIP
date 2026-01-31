'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'

import { LazyBlockEditor } from '@/components/blocks'
import ChatInterface from '@/components/semantic/ChatInterface'
import DocumentList from '@/components/semantic/DocumentList'
import SearchResults from '@/components/semantic/SearchResults'
import type { SearchResult } from '@/components/semantic/SearchResults'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import PageTemplate from '@/components/layout/PageTemplate'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { slideUp } from '@/lib/animations/variants'
import { useSemanticSearch } from '@/lib/api/queries'
import { microcopy } from '@/lib/copy/microcopy'

export default function SemanticPage() {
  const [searchQuery, _setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<unknown[]>([])
  const [notes, setNotes] = useState('')
  const [activeTab, setActiveTab] = useState<'search' | 'notes'>('search')
  const { mutate: search, isPending } = useSemanticSearch()

  const _handleSearch = () => {
    if (!searchQuery.trim()) return
    search(
      { query: searchQuery, top_k: 10 },
      { onSuccess: (data) => setSearchResults(data.results || data.matches || []) }
    )
  }

  return (
    <PageTemplate
      title={microcopy.semantic.title}
      description={microcopy.semantic.subtitle}
      archetype="search"
    >
      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)} className="flex-1 flex flex-col min-h-0">
        <TabsList className="flex-shrink-0">
          <TabsTrigger value="search">Search & Chat</TabsTrigger>
          <TabsTrigger value="notes">Notes</TabsTrigger>
        </TabsList>

        <TabsContent value="search" className="flex-1 min-h-0 mt-4">
          <div className="grid gap-4 md:gap-5 lg:gap-6 md:grid-cols-2 lg:grid-cols-5 h-full">
            <motion.div variants={slideUp} className="flex flex-col min-h-0 lg:col-span-3 lg:border-r lg:border-border lg:pr-6">
              <ChatInterface />
            </motion.div>
            <motion.div variants={slideUp} className="flex flex-col min-h-0 lg:col-span-2">
              {searchResults.length > 0 ? (
                <SearchResults results={searchResults as SearchResult[]} isLoading={isPending} />
              ) : (
                <DocumentList />
              )}
            </motion.div>
          </div>
        </TabsContent>

        <TabsContent value="notes" className="flex-1 min-h-0 mt-4">
          <Card className="h-full flex flex-col">
            <CardHeader>
              <CardTitle>Query Notes & Documentation</CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col min-h-0">
              <div className="flex-1 min-h-0">
                <LazyBlockEditor
                  content={notes}
                  onChange={setNotes}
                  placeholder="Type '/' for commands, or start writing notes about your queries..."
                  editable={true}
                  showToolbar={true}
                  className="h-full"
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </PageTemplate>
  )
}

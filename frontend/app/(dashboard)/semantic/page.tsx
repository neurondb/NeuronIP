'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import ChatInterface from '@/components/semantic/ChatInterface'
import DocumentList from '@/components/semantic/DocumentList'
import SearchResults from '@/components/semantic/SearchResults'
import { useSemanticSearch } from '@/lib/api/queries'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function SemanticPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<unknown[]>([])
  const { mutate: search, isPending } = useSemanticSearch()

  const handleSearch = () => {
    if (!searchQuery.trim()) return

    search(
      {
        query: searchQuery,
        top_k: 10,
      },
      {
        onSuccess: (data) => {
          setSearchResults(data.results || data.matches || [])
        },
      }
    )
  }

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-4 lg:space-y-6 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Semantic Search</h1>
        <p className="text-sm text-muted-foreground mt-1">
          AI-powered semantic search and RAG interface
        </p>
      </motion.div>

      {/* Main Content Grid */}
      <div className="grid gap-4 md:gap-5 lg:gap-6 md:grid-cols-2 lg:grid-cols-5 flex-1 min-h-0">
        {/* Chat Interface - 60% width on large screens, equal on medium */}
        <motion.div 
          variants={slideUp} 
          className="flex flex-col min-h-0 lg:col-span-3 lg:border-r lg:border-border lg:pr-6"
        >
          <ChatInterface />
        </motion.div>

        {/* Search Results / Document List - 40% width on large screens, equal on medium */}
        <motion.div variants={slideUp} className="flex flex-col min-h-0 lg:col-span-2">
          {searchResults.length > 0 ? (
            <SearchResults results={searchResults as any} isLoading={isPending} />
          ) : (
            <DocumentList />
          )}
        </motion.div>
      </div>
    </motion.div>
  )
}

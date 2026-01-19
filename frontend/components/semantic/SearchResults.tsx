'use client'

import { motion } from 'framer-motion'
import { Card, CardContent } from '@/components/ui/Card'
import { slideUp, transition } from '@/lib/animations/variants'

interface SearchResult {
  id: string
  title: string
  content: string
  score: number
  metadata?: Record<string, unknown>
}

interface SearchResultsProps {
  results: SearchResult[]
  isLoading?: boolean
}

export default function SearchResults({ results, isLoading }: SearchResultsProps) {
  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="text-center text-muted-foreground">Searching...</div>
        </CardContent>
      </Card>
    )
  }

  if (results.length === 0) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="text-center text-muted-foreground">No results found</div>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-4 h-full overflow-y-auto">
      {results.map((result, index) => (
        <motion.div
          key={result.id}
          variants={slideUp}
          initial="hidden"
          animate="visible"
          transition={{ ...transition, delay: index * 0.1 }}
        >
          <Card hover>
            <CardContent className="p-4">
              <div className="flex items-start justify-between mb-2">
                <h4 className="font-semibold text-foreground">{result.title}</h4>
                <span className="text-xs text-muted-foreground">
                  {(result.score * 100).toFixed(1)}% match
                </span>
              </div>
              <p className="text-sm text-muted-foreground line-clamp-3">{result.content}</p>
            </CardContent>
          </Card>
        </motion.div>
      ))}
    </div>
  )
}

'use client'

import { motion } from 'framer-motion'
import { ClockIcon, TrashIcon } from '@heroicons/react/24/outline'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useSearchHistory } from '@/lib/hooks/useSearchHistory'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

export default function SearchHistory() {
  const { history, removeFromHistory, clearHistory } = useSearchHistory()

  if (history.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Search History</CardTitle>
          <CardDescription>Your recent searches will appear here</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Search History</CardTitle>
            <CardDescription>Your recent searches</CardDescription>
          </div>
          <Button variant="ghost" size="sm" onClick={clearHistory}>
            Clear All
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {history.map((item, index) => (
            <motion.div
              key={item.id}
              variants={slideUp}
              initial="hidden"
              animate="visible"
              transition={{ ...transition, delay: index * 0.05 }}
              className="flex items-center gap-3 rounded-lg p-3 hover:bg-accent transition-colors group"
            >
              <ClockIcon className="h-4 w-4 text-muted-foreground" />
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{item.query}</p>
                {item.module && (
                  <p className="text-xs text-muted-foreground capitalize">{item.module}</p>
                )}
              </div>
              <button
                onClick={() => removeFromHistory(item.id)}
                className="opacity-0 group-hover:opacity-100 transition-opacity rounded p-1 hover:bg-destructive/10"
              >
                <TrashIcon className="h-4 w-4 text-muted-foreground" />
              </button>
            </motion.div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

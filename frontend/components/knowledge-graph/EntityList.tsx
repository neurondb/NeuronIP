'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { MagnifyingGlassIcon } from '@heroicons/react/24/outline'
import { useSearchEntities } from '@/lib/api/queries'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface Entity {
  id: string
  name: string
  type: string
  description?: string
}

export default function EntityList({ onSelectEntity }: { onSelectEntity?: (entity: Entity) => void }) {
  const [searchQuery, setSearchQuery] = useState('')
  const { mutate: searchEntities, data: searchResults, isPending } = useSearchEntities()

  const handleSearch = () => {
    if (!searchQuery.trim()) return
    searchEntities({ query: searchQuery })
  }

  const entities: Entity[] = searchResults?.entities || []

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">Entities</CardTitle>
        <CardDescription className="text-xs">Search knowledge graph entities</CardDescription>
        <div className="flex gap-2 mt-3">
          <div className="relative flex-1">
            <MagnifyingGlassIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              placeholder="Search entities..."
              className={cn(
                'w-full rounded-lg border border-border bg-background py-2 pl-9 pr-3',
                'text-sm placeholder:text-muted-foreground',
                'focus:outline-none focus:ring-2 focus:ring-ring'
              )}
            />
          </div>
          <Button size="sm" onClick={handleSearch} disabled={isPending || !searchQuery.trim()}>
            Search
          </Button>
        </div>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto min-h-0">
        {isPending ? (
          <div className="flex items-center justify-center py-8">
            <Loading size="md" />
          </div>
        ) : entities.length === 0 ? (
          <div className="text-center text-muted-foreground py-8">
            <p className="text-sm">No entities found</p>
            <p className="text-xs mt-1">Search for entities in the knowledge graph</p>
          </div>
        ) : (
          <motion.table className="w-full border-collapse">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left p-2 text-xs font-medium text-muted-foreground">Name</th>
                <th className="text-left p-2 text-xs font-medium text-muted-foreground">Type</th>
                <th className="text-left p-2 text-xs font-medium text-muted-foreground">Description</th>
              </tr>
            </thead>
            <tbody>
              {entities.map((entity, index) => (
                <motion.tr
                  key={entity.id}
                  variants={slideUp}
                  initial="hidden"
                  animate="visible"
                  transition={{ ...transition, delay: index * 0.03 }}
                  onClick={() => onSelectEntity?.(entity)}
                  className="hover:bg-muted/50 cursor-pointer transition-colors"
                >
                  <td className="p-2 font-medium">{entity.name}</td>
                  <td className="p-2 text-sm text-muted-foreground">{entity.type}</td>
                  <td className="p-2 text-sm text-muted-foreground">{entity.description || '-'}</td>
                </motion.tr>
              ))}
            </tbody>
          </motion.table>
        )}
      </CardContent>
    </Card>
  )
}

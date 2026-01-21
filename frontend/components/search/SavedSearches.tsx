'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import CreateSavedSearchDialog from './CreateSavedSearchDialog'
import { BookmarkIcon, PlayIcon, PencilIcon, TrashIcon } from '@heroicons/react/24/outline'
import { slideUp } from '@/lib/animations/variants'
import { listSavedSearches } from '@/lib/api/saved-searches'

interface SavedSearch {
  id: string
  name: string
  description?: string
  query: string
  tags?: string[]
  is_public: boolean
  created_at: string
}

export default function SavedSearches() {
  const [searches, setSearches] = useState<SavedSearch[]>([])
  const [loading, setLoading] = useState(true)
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)

  useEffect(() => {
    const loadSearches = async () => {
      try {
        const data = await listSavedSearches()
        setSearches(data)
      } catch (err) {
        console.error('Failed to fetch saved searches:', err)
      } finally {
        setLoading(false)
      }
    }
    loadSearches()
  }, [])

  const handleExecute = (searchId: string) => {
    // Execute saved search
    console.log('Executing search:', searchId)
  }

  const handleEdit = (searchId: string) => {
    // Edit saved search
    console.log('Editing search:', searchId)
  }

  const handleDelete = (searchId: string) => {
    // Delete saved search
    console.log('Deleting search:', searchId)
  }

  const handleCreate = () => {
    setIsCreateDialogOpen(true)
  }

  const handleCreated = async () => {
    // Refresh searches
    try {
      const data = await listSavedSearches()
      setSearches(data)
    } catch (err) {
      console.error('Failed to fetch saved searches:', err)
    }
  }

  if (loading) {
    return <div className="text-muted-foreground">Loading saved searches...</div>
  }

  if (searches.length === 0) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <BookmarkIcon className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground mb-4">No saved searches</p>
          <Button onClick={handleCreate}>Create Saved Search</Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {searches.map((search, index) => (
        <motion.div key={search.id} variants={slideUp} transition={{ delay: index * 0.05 }}>
          <Card className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
                  <CardTitle className="text-lg">{search.name}</CardTitle>
                  {search.description && (
                    <CardDescription className="mt-1">{search.description}</CardDescription>
                  )}
                </div>
                {search.is_public && (
                  <span className="text-xs bg-primary/10 text-primary px-2 py-1 rounded">Public</span>
                )}
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <p className="text-sm text-muted-foreground line-clamp-2">{search.query}</p>
                {search.tags && search.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {search.tags.map(tag => (
                      <span key={tag} className="text-xs bg-muted px-2 py-1 rounded">
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
                <div className="flex gap-2 pt-2">
                  <Button variant="outline" size="sm" onClick={() => handleExecute(search.id)}>
                    <PlayIcon className="h-4 w-4 mr-1" />
                    Run
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => handleEdit(search.id)}>
                    <PencilIcon className="h-4 w-4 mr-1" />
                    Edit
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => handleDelete(search.id)}>
                    <TrashIcon className="h-4 w-4 mr-1" />
                    Delete
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      ))}

      {/* Create Saved Search Dialog */}
      <CreateSavedSearchDialog
        open={isCreateDialogOpen}
        onOpenChange={setIsCreateDialogOpen}
        onCreated={handleCreated}
      />
    </div>
  )
}

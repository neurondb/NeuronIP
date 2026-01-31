'use client'

import { useState, useEffect } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import Input from '@/components/ui/Input'

interface Annotation {
  id: string
  target_path: string
  annotation_text: string
  author_id: string
  created_at: string
}

interface InlineAnnotationsProps {
  resourceType: string
  resourceId: string
}

export default function InlineAnnotations({ resourceType, resourceId }: InlineAnnotationsProps) {
  const [annotations, setAnnotations] = useState<Annotation[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [newAnnotation, setNewAnnotation] = useState('')
  const [targetPath, setTargetPath] = useState('')

  useEffect(() => {
    fetchAnnotations()
  }, [resourceType, resourceId])

  const fetchAnnotations = async () => {
    try {
      const response = await fetch(`/api/v1/collaboration/annotations?resource_type=${resourceType}&resource_id=${resourceId}`)
      const data = await response.json()
      setAnnotations(data)
    } catch (error) {
      console.error('Failed to fetch annotations:', error)
    }
  }

  const addAnnotation = async () => {
    try {
      await fetch('/api/v1/collaboration/annotations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          resource_type: resourceType,
          resource_id: resourceId,
          target_type: 'cell',
          target_path: targetPath,
          annotation_text: newAnnotation,
        }),
      })
      setNewAnnotation('')
      setTargetPath('')
      setShowAdd(false)
      fetchAnnotations()
    } catch (error) {
      console.error('Failed to add annotation:', error)
    }
  }

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-medium">Annotations</h3>
          <Button onClick={() => setShowAdd(!showAdd)} size="sm">
            Add Annotation
          </Button>
        </div>
        {showAdd && (
          <div className="mb-4 space-y-2">
            <Input
              placeholder="Target path (e.g., row.0.column.revenue)"
              value={targetPath}
              onChange={(e) => setTargetPath(e.target.value)}
            />
            <Input
              placeholder="Annotation text"
              value={newAnnotation}
              onChange={(e) => setNewAnnotation(e.target.value)}
            />
            <Button onClick={addAnnotation} size="sm">Save</Button>
          </div>
        )}
        <div className="space-y-2">
          {annotations.map((annotation) => (
            <div key={annotation.id} className="p-2 bg-yellow-50 border rounded text-sm">
              <div className="font-medium">{annotation.target_path}</div>
              <div>{annotation.annotation_text}</div>
              <div className="text-xs text-gray-500">{annotation.author_id} • {new Date(annotation.created_at).toLocaleString()}</div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

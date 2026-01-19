'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { useCreateDocument } from '@/lib/api/queries'
import { slideUp, transition } from '@/lib/animations/variants'
import { cn } from '@/lib/utils/cn'

interface Document {
  id: string
  title: string
  type: string
  size: string
  uploadedAt: Date
}

export default function DocumentList() {
  const [documents, setDocuments] = useState<Document[]>([])
  const [isDragging, setIsDragging] = useState(false)
  const { mutate: createDocument, isPending } = useCreateDocument()

  const handleFileUpload = async (files: FileList | null) => {
    if (!files || files.length === 0) return

    Array.from(files).forEach((file) => {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('collection_id', '')

      createDocument(formData, {
        onSuccess: (data) => {
          // Add new document to list
          const newDoc: Document = {
            id: data.id || Date.now().toString(),
            title: file.name,
            type: file.type || 'unknown',
            size: `${(file.size / 1024).toFixed(2)} KB`,
            uploadedAt: new Date(),
          }
          setDocuments((prev) => [...prev, newDoc])
        },
      })
    })
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const handleDragLeave = () => {
    setIsDragging(false)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    handleFileUpload(e.dataTransfer.files)
  }

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle>Documents</CardTitle>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col min-h-0">
        {/* Upload Area */}
        <motion.div
          variants={slideUp}
          initial="hidden"
          animate="visible"
          transition={transition}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          className={cn(
            'border-2 border-dashed rounded-lg p-6 lg:p-8 text-center transition-colors flex-shrink-0',
            isDragging
              ? 'border-primary bg-primary/5'
              : 'border-border hover:border-primary/50'
          )}
        >
          <input
            type="file"
            multiple
            onChange={(e) => handleFileUpload(e.target.files)}
            className="hidden"
            id="file-upload"
            accept=".txt,.md,.pdf,.doc,.docx"
          />
          <label
            htmlFor="file-upload"
            className="cursor-pointer flex flex-col items-center gap-2"
          >
            <p className="text-sm text-muted-foreground">
              Drag and drop files here, or click to select
            </p>
            <Button variant="outline" size="sm">
              Choose Files
            </Button>
          </label>
        </motion.div>

        {/* Document Table */}
        {documents.length > 0 && (
          <div className="mt-6 flex-1 overflow-y-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Title</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Size</TableHead>
                  <TableHead>Uploaded</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {documents.map((doc) => (
                  <TableRow key={doc.id}>
                    <TableCell className="font-medium">{doc.title}</TableCell>
                    <TableCell>{doc.type}</TableCell>
                    <TableCell>{doc.size}</TableCell>
                    <TableCell>{doc.uploadedAt.toLocaleDateString()}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

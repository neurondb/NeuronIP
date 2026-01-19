'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { useExtractEntities } from '@/lib/api/queries'
import { cn } from '@/lib/utils/cn'

export default function EntityExtractor({ onExtract }: { onExtract?: (entities: unknown[]) => void }) {
  const [text, setText] = useState('')
  const { mutate: extractEntities, isPending } = useExtractEntities()

  const handleExtract = () => {
    if (!text.trim() || isPending) return

    extractEntities(
      { text: text.trim() },
      {
        onSuccess: (data) => {
          onExtract?.(data.entities || [])
        },
      }
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base sm:text-lg">Extract Entities</CardTitle>
        <CardDescription className="text-xs">Extract entities from text</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="text-sm font-medium mb-2 block">Text Input</label>
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Enter text to extract entities from..."
            rows={6}
            className={cn(
              'w-full rounded-lg border border-border bg-background px-4 py-3',
              'text-sm focus:outline-none focus:ring-2 focus:ring-ring resize-none'
            )}
          />
        </div>
        <Button onClick={handleExtract} disabled={!text.trim() || isPending} className="w-full">
          {isPending ? <Loading size="sm" /> : 'Extract Entities'}
        </Button>
      </CardContent>
    </Card>
  )
}

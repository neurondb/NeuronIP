'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { PaperAirplaneIcon } from '@heroicons/react/24/outline'
import { useInferModel, useModel } from '@/lib/api/queries'
import { cn } from '@/lib/utils/cn'

interface InferenceInterfaceProps {
  modelId: string
  modelName?: string
}

export default function InferenceInterface({ modelId, modelName }: InferenceInterfaceProps) {
  const [input, setInput] = useState('')
  const [output, setOutput] = useState('')
  const { data: model } = useModel(modelId, !!modelId)
  const { mutate: infer, isPending } = useInferModel(modelId)

  const handleInfer = () => {
    if (!input.trim() || isPending) return

    infer(
      { input: input.trim() },
      {
        onSuccess: (data) => {
          setOutput(data.output || data.result || JSON.stringify(data, null, 2))
        },
        onError: () => {
          setOutput('Error: Inference failed')
        },
      }
    )
  }

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">
          {modelName || model?.name || 'Model Inference'}
        </CardTitle>
        <CardDescription className="text-xs">Run inference on the selected model</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col min-h-0 space-y-4">
        {/* Input */}
        <div className="flex-shrink-0">
          <label className="text-sm font-medium mb-2 block">Input</label>
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Enter input for model inference..."
            rows={6}
            className={cn(
              'w-full rounded-lg border border-border bg-background px-4 py-3',
              'text-sm font-mono focus:outline-none focus:ring-2 focus:ring-ring',
              'resize-none'
            )}
          />
        </div>

        {/* Output */}
        <div className="flex-1 min-h-0 flex flex-col">
          <label className="text-sm font-medium mb-2 block">Output</label>
          <div className="flex-1 min-h-0 rounded-lg border border-border bg-muted p-4 overflow-y-auto">
            {isPending ? (
              <div className="flex items-center justify-center h-full">
                <Loading size="md" />
              </div>
            ) : output ? (
              <pre className="text-sm font-mono whitespace-pre-wrap">{output}</pre>
            ) : (
              <p className="text-sm text-muted-foreground">Output will appear here...</p>
            )}
          </div>
        </div>

        {/* Action */}
        <div className="flex-shrink-0">
          <Button onClick={handleInfer} disabled={!input.trim() || isPending} className="w-full">
            <PaperAirplaneIcon className="h-4 w-4 mr-2" />
            Run Inference
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

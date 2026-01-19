'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Loading from '@/components/ui/Loading'
import { useModels } from '@/lib/api/queries'
import ModelCard from './ModelCard'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

interface Model {
  id: string
  name: string
  model_type: string
  status: 'active' | 'inactive' | 'error'
  description?: string
}

export default function ModelList({ onSelectModel }: { onSelectModel?: (modelId: string) => void }) {
  const { data: models, isLoading } = useModels()

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6">
          <Loading size="md" />
        </CardContent>
      </Card>
    )
  }

  const modelList: Model[] = models || [
    {
      id: '1',
      name: 'GPT-4',
      model_type: 'language-model',
      status: 'active',
      description: 'OpenAI GPT-4 for natural language processing',
    },
    {
      id: '2',
      name: 'Claude-3',
      model_type: 'language-model',
      status: 'active',
      description: 'Anthropic Claude-3 for conversational AI',
    },
    {
      id: '3',
      name: 'Embeddings Model',
      model_type: 'embedding',
      status: 'active',
      description: 'Text embeddings for semantic search',
    },
  ]

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">Registered Models</CardTitle>
        <CardDescription className="text-xs">Manage your AI models</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto min-h-0">
        <motion.div
          variants={staggerContainer}
          initial="hidden"
          animate="visible"
          className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3"
        >
          {modelList.map((model, index) => (
            <ModelCard
              key={model.id}
              {...model}
              onInfer={() => onSelectModel?.(model.id)}
            />
          ))}
        </motion.div>
      </CardContent>
    </Card>
  )
}

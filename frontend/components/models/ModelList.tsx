'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Loading from '@/components/ui/Loading'
import Button from '@/components/ui/Button'
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

interface ModelListProps {
  onSelectModel?: (modelId: string) => void
  onCreateNew?: () => void
}

export default function ModelList({ onSelectModel, onCreateNew }: ModelListProps) {
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

  const modelList: Model[] = models || []

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">Registered Models</CardTitle>
        <CardDescription className="text-xs">Manage your AI models</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto min-h-0">
        {modelList.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground mb-4">No models registered. Register one to get started.</p>
            {onCreateNew && (
              <Button onClick={onCreateNew} size="md">
                Register Your First Model
              </Button>
            )}
          </div>
        ) : (
          <motion.div
            variants={staggerContainer}
            initial="hidden"
            animate="visible"
            className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3"
          >
            {modelList.map((model, index) => (
              <ModelCard
                key={model.id}
                id={model.id}
                name={model.name}
                type={model.model_type}
                status={model.status}
                description={model.description}
                onInfer={() => onSelectModel?.(model.id)}
              />
            ))}
          </motion.div>
        )}
      </CardContent>
    </Card>
  )
}

'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useCreateWorkflow } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'

interface Template {
  id: string
  name: string
  description: string
  category: string
  definition: Record<string, unknown>
}

const TEMPLATES: Template[] = [
  {
    id: 'data-pipeline',
    name: 'Data Pipeline',
    description: 'ETL workflow for data transformation',
    category: 'Data',
    definition: {
      name: 'Data Pipeline',
      steps: [
        { id: 'extract', type: 'script', name: 'Extract Data' },
        { id: 'transform', type: 'script', name: 'Transform Data' },
        { id: 'load', type: 'script', name: 'Load Data' },
      ],
      start_step: 'extract',
    },
  },
  {
    id: 'approval-workflow',
    name: 'Approval Workflow',
    description: 'Multi-step approval process with conditions',
    category: 'Business',
    definition: {
      name: 'Approval Workflow',
      steps: [
        { id: 'submit', type: 'agent', name: 'Submit Request' },
        { id: 'review', type: 'condition', name: 'Review Decision' },
        { id: 'approve', type: 'agent', name: 'Approve' },
        { id: 'reject', type: 'agent', name: 'Reject' },
      ],
      start_step: 'submit',
    },
  },
  {
    id: 'notification-chain',
    name: 'Notification Chain',
    description: 'Sequential notifications with escalation',
    category: 'Communication',
    definition: {
      name: 'Notification Chain',
      steps: [
        { id: 'notify1', type: 'agent', name: 'First Notification' },
        { id: 'wait', type: 'script', name: 'Wait Period' },
        { id: 'notify2', type: 'agent', name: 'Escalation Notification' },
      ],
      start_step: 'notify1',
    },
  },
  {
    id: 'parallel-processing',
    name: 'Parallel Processing',
    description: 'Execute multiple tasks in parallel',
    category: 'Performance',
    definition: {
      name: 'Parallel Processing',
      steps: [
        { id: 'split', type: 'parallel', name: 'Split Tasks' },
        { id: 'process1', type: 'script', name: 'Process Task 1' },
        { id: 'process2', type: 'script', name: 'Process Task 2' },
        { id: 'merge', type: 'script', name: 'Merge Results' },
      ],
      start_step: 'split',
    },
  },
]

export default function WorkflowTemplates() {
  const [selectedCategory, setSelectedCategory] = useState<string>('All')
  const { mutate: createWorkflow, isPending } = useCreateWorkflow()

  const categories = ['All', ...Array.from(new Set(TEMPLATES.map((t) => t.category)))]

  const filteredTemplates =
    selectedCategory === 'All'
      ? TEMPLATES
      : TEMPLATES.filter((t) => t.category === selectedCategory)

  const handleUseTemplate = (template: Template) => {
    createWorkflow(template.definition, {
      onSuccess: () => {
        showToast(`Workflow "${template.name}" created successfully`, 'success')
      },
      onError: () => {
        showToast('Failed to create workflow from template', 'error')
      },
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Workflow Templates</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Category Filter */}
        <div className="flex gap-2 mb-4 flex-wrap">
          {categories.map((category) => (
            <button
              key={category}
              onClick={() => setSelectedCategory(category)}
              className={`px-3 py-1 rounded-full text-sm transition-colors ${
                selectedCategory === category
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {category}
            </button>
          ))}
        </div>

        {/* Templates Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filteredTemplates.map((template) => (
            <motion.div
              key={template.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="p-4 border border-border rounded-lg hover:shadow-md transition-shadow"
            >
              <div className="flex items-start justify-between mb-2">
                <div>
                  <h3 className="font-semibold">{template.name}</h3>
                  <span className="text-xs text-muted-foreground">{template.category}</span>
                </div>
              </div>
              <p className="text-sm text-muted-foreground mb-4">{template.description}</p>
              <Button
                onClick={() => handleUseTemplate(template)}
                disabled={isPending}
                size="sm"
                className="w-full"
              >
                {isPending ? 'Creating...' : 'Use Template'}
              </Button>
            </motion.div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

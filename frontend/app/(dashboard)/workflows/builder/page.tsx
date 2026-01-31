'use client'

import dynamic from 'next/dynamic'

import PageTemplate from '@/components/layout/PageTemplate'
import Loading from '@/components/ui/Loading'

const WorkflowBuilder = dynamic(
  () => import('@/components/workflows/WorkflowBuilder').then((mod) => ({ default: mod.default })),
  {
    ssr: false,
    loading: () => (
      <div className="flex h-[400px] items-center justify-center rounded-lg border border-border bg-muted/30">
        <Loading size="lg" variant="spinner" />
      </div>
    ),
  }
)

export default function WorkflowBuilderPage() {
  return (
    <PageTemplate
      title="Workflow Builder"
      description="Create and edit workflows with a visual drag-and-drop interface"
      archetype="builder"
    >
      <WorkflowBuilder />
    </PageTemplate>
  )
}

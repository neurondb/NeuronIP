'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'

interface WorkflowNodePaletteProps {
  onAddNode: (type: 'agent' | 'script' | 'condition' | 'parallel') => void
}

export default function WorkflowNodePalette({ onAddNode }: WorkflowNodePaletteProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Node Palette</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        <Button
          onClick={() => onAddNode('agent')}
          className="w-full justify-start"
          variant="outline"
        >
          Agent Node
        </Button>
        <Button
          onClick={() => onAddNode('script')}
          className="w-full justify-start"
          variant="outline"
        >
          Script Node
        </Button>
        <Button
          onClick={() => onAddNode('condition')}
          className="w-full justify-start"
          variant="outline"
        >
          Condition Node
        </Button>
        <Button
          onClick={() => onAddNode('parallel')}
          className="w-full justify-start"
          variant="outline"
        >
          Parallel Node
        </Button>
      </CardContent>
    </Card>
  )
}

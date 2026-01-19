'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import { CheckCircleIcon, XCircleIcon, ClockIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface Execution {
  id: string
  workflowId: string
  status: 'running' | 'completed' | 'failed'
  startedAt: Date
  completedAt?: Date
  duration?: number
}

interface ExecutionLogProps {
  executions: Execution[]
}

export default function ExecutionLog({ executions }: ExecutionLogProps) {
  const getStatusIcon = (status: Execution['status']) => {
    switch (status) {
      case 'completed':
        return <CheckCircleIcon className="h-5 w-5 text-green-600" />
      case 'failed':
        return <XCircleIcon className="h-5 w-5 text-red-600" />
      case 'running':
        return <ClockIcon className="h-5 w-5 text-blue-600" />
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Execution Log</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Status</TableHead>
              <TableHead>Started</TableHead>
              <TableHead>Duration</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {executions.map((execution, index) => (
              <motion.tr
                key={execution.id}
                variants={slideUp}
                initial="hidden"
                animate="visible"
                transition={{ ...transition, delay: index * 0.05 }}
              >
                <TableCell>
                  <div className="flex items-center gap-2">
                    {getStatusIcon(execution.status)}
                    <span className="capitalize">{execution.status}</span>
                  </div>
                </TableCell>
                <TableCell>{execution.startedAt.toLocaleString()}</TableCell>
                <TableCell>
                  {execution.duration
                    ? `${execution.duration}ms`
                    : 'In progress...'}
                </TableCell>
              </motion.tr>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}

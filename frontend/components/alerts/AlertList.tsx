'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { CheckCircleIcon, XCircleIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline'
import { useAlerts, useResolveAlert } from '@/lib/api/queries'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface Alert {
  id: string
  type: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  message: string
  status: 'active' | 'resolved'
  createdAt: Date
}

export default function AlertList() {
  const { data: alerts, isLoading } = useAlerts()
  const { mutate: resolveAlert } = useResolveAlert()

  const getSeverityIcon = (severity: Alert['severity']) => {
    switch (severity) {
      case 'critical':
      case 'high':
        return <XCircleIcon className="h-5 w-5 text-red-600" />
      case 'medium':
        return <ExclamationTriangleIcon className="h-5 w-5 text-yellow-600" />
      case 'low':
        return <CheckCircleIcon className="h-5 w-5 text-blue-600" />
    }
  }

  const getSeverityColor = (severity: Alert['severity']) => {
    switch (severity) {
      case 'critical':
      case 'high':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
      case 'low':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
    }
  }

  const alertList: Alert[] = alerts?.alerts || [
    {
      id: '1',
      type: 'anomaly',
      severity: 'high',
      message: 'Unusual spike in query volume detected',
      status: 'active',
      createdAt: new Date(Date.now() - 1000 * 60 * 15),
    },
    {
      id: '2',
      type: 'compliance',
      severity: 'critical',
      message: 'Compliance violation detected in document processing',
      status: 'active',
      createdAt: new Date(Date.now() - 1000 * 60 * 30),
    },
    {
      id: '3',
      type: 'performance',
      severity: 'medium',
      message: 'Workflow execution time exceeded threshold',
      status: 'resolved',
      createdAt: new Date(Date.now() - 1000 * 60 * 60 * 2),
    },
  ]

  const handleResolve = (id: string) => {
    resolveAlert(id)
  }

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">Active Alerts</CardTitle>
        <CardDescription className="text-xs">
          {alertList.filter((a) => a.status === 'active').length} active alerts
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto min-h-0">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loading size="md" />
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Severity</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Message</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {alertList.map((alert, index) => (
                <motion.tr
                  key={alert.id}
                  variants={slideUp}
                  initial="hidden"
                  animate="visible"
                  transition={{ ...transition, delay: index * 0.03 }}
                  className="hover:bg-muted/50 transition-colors"
                >
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {getSeverityIcon(alert.severity)}
                      <span className={cn('px-2 py-1 rounded-full text-xs font-medium', getSeverityColor(alert.severity))}>
                        {alert.severity}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm capitalize">{alert.type}</TableCell>
                  <TableCell className="font-medium">{alert.message}</TableCell>
                  <TableCell>
                    <span className={cn(
                      'px-2 py-1 rounded-full text-xs font-medium',
                      alert.status === 'active' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                    )}>
                      {alert.status}
                    </span>
                  </TableCell>
                  <TableCell>
                    {alert.status === 'active' && (
                      <Button size="sm" variant="outline" onClick={() => handleResolve(alert.id)}>
                        Resolve
                      </Button>
                    )}
                  </TableCell>
                </motion.tr>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}

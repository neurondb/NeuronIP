'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { useAlertRules } from '@/lib/api/queries'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface AlertRule {
  id: string
  name: string
  condition: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  enabled: boolean
}

export default function AlertRules({ onCreateRule }: { onCreateRule?: () => void }) {
  const { data: rules, isLoading } = useAlertRules()

  const ruleList: AlertRule[] = rules?.rules || [
    {
      id: '1',
      name: 'Query Volume Spike',
      condition: 'query_count > 1000 per hour',
      severity: 'high',
      enabled: true,
    },
    {
      id: '2',
      name: 'Compliance Violation',
      condition: 'compliance_score < 0.8',
      severity: 'critical',
      enabled: true,
    },
    {
      id: '3',
      name: 'Performance Degradation',
      condition: 'avg_response_time > 5s',
      severity: 'medium',
      enabled: true,
    },
  ]

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-base sm:text-lg">Alert Rules</CardTitle>
            <CardDescription className="text-xs">Configure alert conditions</CardDescription>
          </div>
          {onCreateRule && (
            <Button size="sm" onClick={onCreateRule}>
              Create Rule
            </Button>
          )}
        </div>
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
                <TableHead>Name</TableHead>
                <TableHead>Condition</TableHead>
                <TableHead>Severity</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {ruleList.map((rule, index) => (
                <motion.tr
                  key={rule.id}
                  variants={slideUp}
                  initial="hidden"
                  animate="visible"
                  transition={{ ...transition, delay: index * 0.03 }}
                  className="hover:bg-muted/50 cursor-pointer transition-colors"
                >
                  <TableCell className="font-medium">{rule.name}</TableCell>
                  <TableCell className="text-sm font-mono text-muted-foreground">{rule.condition}</TableCell>
                  <TableCell className="text-sm capitalize">{rule.severity}</TableCell>
                  <TableCell>
                    <span className={cn(
                      'px-2 py-1 rounded-full text-xs font-medium',
                      rule.enabled ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                    )}>
                      {rule.enabled ? 'Enabled' : 'Disabled'}
                    </span>
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

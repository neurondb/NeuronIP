'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import AnomalyChart from '@/components/compliance/AnomalyChart'
import ComplianceTable from '@/components/compliance/ComplianceTable'
import { useComplianceCheck, useComplianceAnomalies } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

interface ComplianceCheck {
  id: string
  entityType: string
  entityId: string
  status: 'compliant' | 'non-compliant'
  checkedAt: Date
  matches: number
}

export default function CompliancePage() {
  const [entityType, setEntityType] = useState('')
  const [entityId, setEntityId] = useState('')
  const [entityContent, setEntityContent] = useState('')
  const [checks, setChecks] = useState<ComplianceCheck[]>([])
  
  const { mutate: checkCompliance, isPending } = useComplianceCheck()
  const { data: anomalies, isLoading: anomaliesLoading } = useComplianceAnomalies()

  const handleCheck = () => {
    if (!entityType || !entityId || !entityContent) {
      showToast('Please fill in all fields', 'warning')
      return
    }

    checkCompliance(
      {
        entity_type: entityType,
        entity_id: entityId,
        entity_content: entityContent,
      },
      {
        onSuccess: (data) => {
          const check: ComplianceCheck = {
            id: Date.now().toString(),
            entityType,
            entityId,
            status: (data.matches?.length || 0) === 0 ? 'compliant' : 'non-compliant',
            checkedAt: new Date(),
            matches: data.matches?.length || data.count || 0,
          }
          setChecks((prev) => [check, ...prev])
          showToast('Compliance check completed', 'success')
        },
        onError: () => {
          showToast('Compliance check failed', 'error')
        },
      }
    )
  }

  // Mock chart data
  const chartData = [
    { date: 'Mon', anomalies: 2, violations: 1 },
    { date: 'Tue', anomalies: 5, violations: 2 },
    { date: 'Wed', anomalies: 3, violations: 0 },
    { date: 'Thu', anomalies: 7, violations: 3 },
    { date: 'Fri', anomalies: 4, violations: 1 },
    { date: 'Sat', anomalies: 1, violations: 0 },
    { date: 'Sun', anomalies: 3, violations: 1 },
  ]

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Compliance</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Monitor compliance and detect anomalies
        </p>
      </motion.div>

      {/* Compliance Check Form */}
      <motion.div variants={slideUp}>
        <Card>
          <CardHeader>
            <CardTitle>Compliance Check</CardTitle>
            <CardDescription>Check entity compliance against rules</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <label className="text-sm font-medium mb-2 block">Entity Type</label>
              <input
                type="text"
                value={entityType}
                onChange={(e) => setEntityType(e.target.value)}
                placeholder="e.g., document, user, data"
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Entity ID</label>
              <input
                type="text"
                value={entityId}
                onChange={(e) => setEntityId(e.target.value)}
                placeholder="Entity identifier"
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Content</label>
              <textarea
                value={entityContent}
                onChange={(e) => setEntityContent(e.target.value)}
                placeholder="Entity content to check"
                rows={4}
                className="w-full rounded-lg border border-border bg-background px-4 py-2 focus:outline-none focus:ring-2 focus:ring-ring resize-none"
              />
            </div>
            <Button onClick={handleCheck} disabled={isPending}>
              Check Compliance
            </Button>
          </CardContent>
        </Card>
      </motion.div>

      {/* Anomaly Chart */}
      <motion.div variants={slideUp} className="flex-shrink-0">
        <AnomalyChart data={chartData} />
      </motion.div>

      {/* Compliance Checks Table */}
      {checks.length > 0 && (
        <motion.div variants={slideUp} className="flex-1 min-h-0">
          <ComplianceTable checks={checks} />
        </motion.div>
      )}
    </motion.div>
  )
}

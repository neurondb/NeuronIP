'use client'

import { motion } from 'framer-motion'
import { useState } from 'react'

import AnomalyChart from '@/components/compliance/AnomalyChart'
import ComplianceDashboard from '@/components/compliance/ComplianceDashboard'
import ComplianceTable from '@/components/compliance/ComplianceTable'
import PolicyManager from '@/components/compliance/PolicyManager'
import RuleBuilder from '@/components/compliance/RuleBuilder'
import Button from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { showToast } from '@/components/ui/Toast'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import { useComplianceCheck, useComplianceAnomalies, getComplianceReportExportUrl } from '@/lib/api/queries'

interface ComplianceCheck {
  id: string
  entityType: string
  entityId: string
  status: 'compliant' | 'non-compliant'
  checkedAt: Date
  matches: number
}

export default function CompliancePage() {
  const [activeTab, setActiveTab] = useState<'check' | 'policies' | 'rules' | 'dashboard'>('check')
  const [entityType, setEntityType] = useState('')
  const [entityId, setEntityId] = useState('')
  const [entityContent, setEntityContent] = useState('')
  const [checks, setChecks] = useState<ComplianceCheck[]>([])
  
  const { mutate: checkCompliance, isPending } = useComplianceCheck()
  // Anomalies data can be used for future enhancements
  useComplianceAnomalies()

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
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Governance & Compliance</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Monitor compliance, manage policies, and detect anomalies
          </p>
        </div>
        <div className="flex gap-2">
          <a
            href={getComplianceReportExportUrl({ format: 'csv' })}
            download
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium hover:bg-muted"
          >
            Export report CSV
          </a>
          <a
            href={getComplianceReportExportUrl({ format: 'json' })}
            download
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium hover:bg-muted"
          >
            Export report JSON
          </a>
        </div>
      </motion.div>

      {/* Tabs */}
      <motion.div variants={slideUp} className="flex-shrink-0">
        <div className="flex gap-2 border-b border-border">
          <Button
            variant={activeTab === 'check' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('check')}
            size="sm"
          >
            Check
          </Button>
          <Button
            variant={activeTab === 'policies' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('policies')}
            size="sm"
          >
            Policies
          </Button>
          <Button
            variant={activeTab === 'rules' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('rules')}
            size="sm"
          >
            Rules
          </Button>
          <Button
            variant={activeTab === 'dashboard' ? 'primary' : 'ghost'}
            onClick={() => setActiveTab('dashboard')}
            size="sm"
          >
            Dashboard
          </Button>
        </div>
      </motion.div>

      {/* Tab Content */}
      {activeTab === 'check' && (
        <>
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
        </>
      )}

      {activeTab === 'policies' && (
        <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
          <PolicyManager />
        </motion.div>
      )}

      {activeTab === 'rules' && (
        <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto">
          <RuleBuilder />
        </motion.div>
      )}

      {activeTab === 'dashboard' && (
        <motion.div variants={slideUp} className="flex-1 min-h-0 overflow-y-auto space-y-4">
          <ComplianceDashboard />
          <AnomalyChart data={chartData} />
        </motion.div>
      )}
    </motion.div>
  )
}

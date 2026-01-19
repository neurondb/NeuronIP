'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import AlertList from '@/components/alerts/AlertList'
import AlertRules from '@/components/alerts/AlertRules'
import Modal from '@/components/ui/Modal'
import Button from '@/components/ui/Button'
import { useCreateAlertRule } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export default function AlertsPage() {
  const [createRuleOpen, setCreateRuleOpen] = useState(false)
  const [ruleName, setRuleName] = useState('')
  const [condition, setCondition] = useState('')
  const [severity, setSeverity] = useState<'low' | 'medium' | 'high' | 'critical'>('medium')
  const { mutate: createRule, isPending } = useCreateAlertRule()

  const handleCreateRule = (e: React.FormEvent) => {
    e.preventDefault()
    if (!ruleName || !condition) {
      showToast('Name and condition are required', 'warning')
      return
    }

    createRule(
      {
        name: ruleName,
        condition,
        severity,
        enabled: true,
      },
      {
        onSuccess: () => {
          showToast('Alert rule created successfully', 'success')
          setCreateRuleOpen(false)
          setRuleName('')
          setCondition('')
          setSeverity('medium')
        },
        onError: () => {
          showToast('Failed to create alert rule', 'error')
        },
      }
    )
  }

  return (
    <>
      <motion.div
        variants={staggerContainer}
        initial="hidden"
        animate="visible"
        className="space-y-3 sm:space-y-4 flex flex-col h-full"
      >
        {/* Page Header */}
        <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
          <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Alerts</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Monitor system alerts and manage alert rules
          </p>
        </motion.div>

        {/* Main Content Grid */}
        <div className="grid gap-3 sm:gap-4 lg:grid-cols-2 flex-1 min-h-0">
          {/* Alert List */}
          <motion.div variants={slideUp} className="flex flex-col min-h-0">
            <AlertList />
          </motion.div>

          {/* Alert Rules */}
          <motion.div variants={slideUp} className="flex flex-col min-h-0">
            <AlertRules onCreateRule={() => setCreateRuleOpen(true)} />
          </motion.div>
        </div>
      </motion.div>

      {/* Create Rule Modal */}
      <Modal
        open={createRuleOpen}
        onOpenChange={setCreateRuleOpen}
        title="Create Alert Rule"
        description="Define conditions for alert triggering"
        size="md"
      >
        <form onSubmit={handleCreateRule} className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-2 block">Rule Name</label>
            <input
              type="text"
              value={ruleName}
              onChange={(e) => setRuleName(e.target.value)}
              placeholder="e.g., Query Volume Spike"
              className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-2 block">Condition</label>
            <input
              type="text"
              value={condition}
              onChange={(e) => setCondition(e.target.value)}
              placeholder="e.g., query_count > 1000 per hour"
              className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-ring"
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium mb-2 block">Severity</label>
            <select
              value={severity}
              onChange={(e) => setSeverity(e.target.value as typeof severity)}
              className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={() => setCreateRuleOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              Create Rule
            </Button>
          </div>
        </form>
      </Modal>
    </>
  )
}

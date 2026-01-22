'use client'

import { motion } from 'framer-motion'
import GeneralSettings from '@/components/settings/GeneralSettings'
import IntegrationSettings from '@/components/settings/IntegrationSettings'
import SecuritySettings from '@/components/settings/SecuritySettings'
import PreferencesPanel from '@/components/settings/PreferencesPanel'
import { staggerContainer, slideUp } from '@/lib/animations/variants'

export const dynamic = 'force-dynamic'

export default function SettingsPage() {
  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-3 sm:space-y-4 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <h1 className="text-2xl sm:text-3xl font-bold text-foreground">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage application settings and preferences
        </p>
      </motion.div>

      {/* Settings Grid */}
      <div className="grid gap-3 sm:gap-4 lg:grid-cols-2 flex-1 min-h-0 overflow-y-auto">
        <motion.div variants={slideUp}>
          <GeneralSettings />
        </motion.div>

        <motion.div variants={slideUp}>
          <IntegrationSettings />
        </motion.div>

        <motion.div variants={slideUp}>
          <SecuritySettings />
        </motion.div>

        <motion.div variants={slideUp}>
          <PreferencesPanel />
        </motion.div>
      </div>
    </motion.div>
  )
}

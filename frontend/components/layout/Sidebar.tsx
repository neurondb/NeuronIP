'use client'

import { usePathname } from 'next/navigation'
import Link from 'next/link'
import { motion } from 'framer-motion'
import {
  HomeIcon,
  MagnifyingGlassIcon,
  CubeIcon,
  Cog6ToothIcon,
  CommandLineIcon,
  LifebuoyIcon,
  Bars3Icon,
  XMarkIcon,
  CpuChipIcon,
  CircleStackIcon,
  BellAlertIcon,
  KeyIcon,
  UserGroupIcon,
  SparklesIcon,
  ServerIcon,
  ChartBarIcon,
  UserCircleIcon,
  EyeIcon,
  ArrowsRightLeftIcon,
  DocumentTextIcon,
  CreditCardIcon,
  PuzzlePieceIcon,
  ClockIcon,
  FolderIcon,
} from '@heroicons/react/24/outline'
import { useAppStore } from '@/lib/store/useAppStore'
import { cn } from '@/lib/utils/cn'

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: HomeIcon },
  { name: 'Semantic Search', href: '/dashboard/semantic', icon: MagnifyingGlassIcon },
  { name: 'Warehouse', href: '/dashboard/warehouse', icon: CubeIcon },
  { name: 'Data Sources', href: '/dashboard/data-sources', icon: ServerIcon },
  { name: 'Metrics', href: '/dashboard/metrics', icon: ChartBarIcon },
  { name: 'Agent Hub', href: '/dashboard/agents', icon: UserCircleIcon },
  { name: 'Observability', href: '/dashboard/observability', icon: EyeIcon },
  { name: 'Data Lineage', href: '/dashboard/lineage', icon: ArrowsRightLeftIcon },
  { name: 'Audit', href: '/dashboard/audit', icon: DocumentTextIcon },
  { name: 'Billing', href: '/dashboard/billing', icon: CreditCardIcon },
  { name: 'Integrations', href: '/dashboard/integrations', icon: PuzzlePieceIcon },
  { name: 'Versioning', href: '/dashboard/versioning', icon: ClockIcon },
  { name: 'Data Catalog', href: '/dashboard/catalog', icon: FolderIcon },
  { name: 'Workflows', href: '/dashboard/workflows', icon: CommandLineIcon },
  { name: 'Compliance', href: '/dashboard/compliance', icon: Cog6ToothIcon },
  { name: 'Support', href: '/dashboard/support', icon: LifebuoyIcon },
  { name: 'Models', href: '/dashboard/models', icon: CpuChipIcon },
  { name: 'Knowledge Graph', href: '/dashboard/knowledge-graph', icon: CircleStackIcon },
  { name: 'Alerts', href: '/dashboard/alerts', icon: BellAlertIcon },
  { name: 'API Keys', href: '/dashboard/api-keys', icon: KeyIcon },
  { name: 'Users', href: '/dashboard/users', icon: UserGroupIcon },
  { name: 'Settings', href: '/dashboard/settings', icon: Cog6ToothIcon },
  { name: 'Features', href: '/dashboard/features', icon: SparklesIcon },
]

export default function Sidebar() {
  const pathname = usePathname()
  const { sidebarCollapsed, toggleSidebar } = useAppStore()

  return (
    <>
      {/* Mobile backdrop */}
      {!sidebarCollapsed && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={toggleSidebar}
        />
      )}

      {/* Sidebar */}
      <motion.aside
        className={cn(
          'fixed top-0 left-0 z-50 h-screen bg-card border-r border-border transition-all duration-300 lg:static lg:z-auto',
          sidebarCollapsed ? '-translate-x-full lg:translate-x-0 lg:w-16' : 'w-64'
        )}
        initial={false}
        animate={{ width: sidebarCollapsed ? 64 : 256 }}
        transition={{ duration: 0.3 }}
      >
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex h-16 items-center justify-between border-b border-border px-4">
            {!sidebarCollapsed && (
              <motion.h2
                className="text-xl font-bold text-foreground"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
              >
                NeuronIP
              </motion.h2>
            )}
            <button
              onClick={toggleSidebar}
              className="rounded-lg p-2 hover:bg-accent transition-colors"
            >
              {sidebarCollapsed ? (
                <Bars3Icon className="h-5 w-5" />
              ) : (
                <XMarkIcon className="h-5 w-5 lg:hidden" />
              )}
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-1 px-3 py-4">
            {navigation.map((item) => {
              const isActive = pathname === item.href
              const Icon = item.icon

              return (
                <Link key={item.name} href={item.href}>
                  <motion.div
                    className={cn(
                      'group flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-primary text-primary-foreground'
                        : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                    )}
                    whileHover={{ x: 4 }}
                    transition={{ duration: 0.2 }}
                  >
                    <Icon className="h-5 w-5 flex-shrink-0" />
                    {!sidebarCollapsed && (
                      <motion.span
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        transition={{ delay: 0.1 }}
                      >
                        {item.name}
                      </motion.span>
                    )}
                  </motion.div>
                </Link>
              )
            })}
          </nav>
        </div>
      </motion.aside>
    </>
  )
}

'use client'

import { usePathname } from 'next/navigation'
import Link from 'next/link'
import { motion, AnimatePresence } from 'framer-motion'
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
  InformationCircleIcon,
  ChevronDownIcon,
  ChevronRightIcon,
} from '@heroicons/react/24/outline'
import { useAppStore } from '@/lib/store/useAppStore'
import { cn } from '@/lib/utils/cn'

interface NavItem {
  name: string
  href: string
  icon: React.ComponentType<{ className?: string }>
}

interface NavGroup {
  id: string
  name: string
  icon: React.ComponentType<{ className?: string }>
  items: NavItem[]
}

const navigationGroups: NavGroup[] = [
  {
    id: 'overview',
    name: 'Overview',
    icon: HomeIcon,
    items: [
      { name: 'Dashboard', href: '/dashboard', icon: HomeIcon },
      { name: 'Why NeuronIP', href: '/dashboard/why-neuronip', icon: InformationCircleIcon },
    ],
  },
  {
    id: 'data-analytics',
    name: 'Data & Analytics',
    icon: CubeIcon,
    items: [
      { name: 'Semantic Search', href: '/dashboard/semantic', icon: MagnifyingGlassIcon },
      { name: 'Warehouse', href: '/dashboard/warehouse', icon: CubeIcon },
      { name: 'Data Sources', href: '/dashboard/data-sources', icon: ServerIcon },
      { name: 'Metrics', href: '/dashboard/metrics', icon: ChartBarIcon },
      { name: 'Data Catalog', href: '/dashboard/catalog', icon: FolderIcon },
      { name: 'Knowledge Graph', href: '/dashboard/knowledge-graph', icon: CircleStackIcon },
    ],
  },
  {
    id: 'ai-automation',
    name: 'AI & Automation',
    icon: CpuChipIcon,
    items: [
      { name: 'Agent Hub', href: '/dashboard/agents', icon: UserCircleIcon },
      { name: 'Models', href: '/dashboard/models', icon: CpuChipIcon },
      { name: 'Workflows', href: '/dashboard/workflows', icon: CommandLineIcon },
    ],
  },
  {
    id: 'observability',
    name: 'Observability & Monitoring',
    icon: EyeIcon,
    items: [
      { name: 'Observability', href: '/dashboard/observability', icon: EyeIcon },
      { name: 'Alerts', href: '/dashboard/alerts', icon: BellAlertIcon },
      { name: 'Data Lineage', href: '/dashboard/lineage', icon: ArrowsRightLeftIcon },
    ],
  },
  {
    id: 'governance',
    name: 'Governance & Compliance',
    icon: DocumentTextIcon,
    items: [
      { name: 'Compliance', href: '/dashboard/compliance', icon: Cog6ToothIcon },
      { name: 'Audit', href: '/dashboard/audit', icon: DocumentTextIcon },
      { name: 'Versioning', href: '/dashboard/versioning', icon: ClockIcon },
    ],
  },
  {
    id: 'administration',
    name: 'Administration',
    icon: Cog6ToothIcon,
    items: [
      { name: 'Users', href: '/dashboard/users', icon: UserGroupIcon },
      { name: 'API Keys', href: '/dashboard/api-keys', icon: KeyIcon },
      { name: 'Integrations', href: '/dashboard/integrations', icon: PuzzlePieceIcon },
      { name: 'Settings', href: '/dashboard/settings', icon: Cog6ToothIcon },
    ],
  },
  {
    id: 'business',
    name: 'Business',
    icon: CreditCardIcon,
    items: [
      { name: 'Billing', href: '/dashboard/billing', icon: CreditCardIcon },
      { name: 'Support', href: '/dashboard/support', icon: LifebuoyIcon },
      { name: 'Features', href: '/dashboard/features', icon: SparklesIcon },
    ],
  },
]

function NavGroupSection({ group }: { group: NavGroup }) {
  const pathname = usePathname()
  const { sidebarCollapsed, expandedNavGroups, toggleNavGroup } = useAppStore()
  const isExpanded = expandedNavGroups[group.id] ?? false
  const GroupIcon = group.icon

  // Check if any item in the group is active
  const hasActiveItem = group.items.some((item) => pathname === item.href)

  // For overview group, always show items (no collapse)
  if (group.id === 'overview') {
    return (
      <div className="space-y-1">
        {group.items.map((item) => {
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
      </div>
    )
  }

  return (
    <div className="space-y-1">
      <button
        onClick={() => toggleNavGroup(group.id)}
        className={cn(
          'w-full flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
          hasActiveItem
            ? 'bg-primary/10 text-primary'
            : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
        )}
      >
        <GroupIcon className="h-5 w-5 flex-shrink-0" />
        {!sidebarCollapsed && (
          <>
            <span className="flex-1 text-left">{group.name}</span>
            {isExpanded ? (
              <ChevronDownIcon className="h-4 w-4" />
            ) : (
              <ChevronRightIcon className="h-4 w-4" />
            )}
          </>
        )}
      </button>
      <AnimatePresence>
        {isExpanded && !sidebarCollapsed && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="overflow-hidden"
          >
            <div className="ml-4 space-y-1 border-l border-border pl-4">
              {group.items.map((item) => {
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
                      <span>{item.name}</span>
                    </motion.div>
                  </Link>
                )
              })}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
      {/* Show items as icons only when collapsed */}
      {sidebarCollapsed && (
        <div className="space-y-1">
          {group.items.map((item) => {
            const isActive = pathname === item.href
            const Icon = item.icon
            return (
              <Link key={item.name} href={item.href} title={item.name}>
                <motion.div
                  className={cn(
                    'flex items-center justify-center rounded-lg p-2 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                  )}
                  whileHover={{ scale: 1.1 }}
                  transition={{ duration: 0.2 }}
                >
                  <Icon className="h-5 w-5" />
                </motion.div>
              </Link>
            )
          })}
        </div>
      )}
    </div>
  )
}

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
          'fixed top-0 left-0 z-50 h-screen bg-card border-r border-border transition-all duration-300 lg:static lg:z-auto overflow-y-auto',
          sidebarCollapsed ? '-translate-x-full lg:translate-x-0 lg:w-16' : 'w-64'
        )}
        initial={false}
        animate={{ width: sidebarCollapsed ? 64 : 256 }}
        transition={{ duration: 0.3 }}
      >
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex h-16 items-center justify-between border-b border-border px-4 shrink-0">
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
              aria-label={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            >
              {sidebarCollapsed ? (
                <Bars3Icon className="h-5 w-5" />
              ) : (
                <XMarkIcon className="h-5 w-5 lg:hidden" />
              )}
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-2 px-3 py-4 overflow-y-auto">
            {navigationGroups.map((group) => (
              <NavGroupSection key={group.id} group={group} />
            ))}
          </nav>
        </div>
      </motion.aside>
    </>
  )
}

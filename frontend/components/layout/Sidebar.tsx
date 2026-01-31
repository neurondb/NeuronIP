'use client'

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
  StarIcon,
} from '@heroicons/react/24/outline'
import { StarIcon as StarIconSolid } from '@heroicons/react/24/solid'
import { motion, AnimatePresence } from 'framer-motion'
import Link from 'next/link'
import { usePathname } from 'next/navigation'

import { getRouteMeta } from '@/lib/pageArchetypes'
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

/** Fewer top-level groups, clearer naming (Notion-like flow). */
const navigationGroups: NavGroup[] = [
  {
    id: 'overview',
    name: 'Overview',
    icon: HomeIcon,
    items: [
      { name: 'Dashboard', href: '/', icon: HomeIcon },
      { name: 'Why NeuronIP', href: '/why-neuronip', icon: InformationCircleIcon },
    ],
  },
  {
    id: 'data',
    name: 'Data',
    icon: CubeIcon,
    items: [
      { name: 'Semantic Search', href: '/semantic', icon: MagnifyingGlassIcon },
      { name: 'Warehouse', href: '/warehouse', icon: CubeIcon },
      { name: 'Data Sources', href: '/data-sources', icon: ServerIcon },
      { name: 'Metrics', href: '/metrics', icon: ChartBarIcon },
      { name: 'Catalog', href: '/catalog', icon: FolderIcon },
      { name: 'Knowledge Graph', href: '/knowledge-graph', icon: CircleStackIcon },
    ],
  },
  {
    id: 'ai',
    name: 'AI & Automation',
    icon: CpuChipIcon,
    items: [
      { name: 'Agents', href: '/agents', icon: UserCircleIcon },
      { name: 'Models', href: '/models', icon: CpuChipIcon },
      { name: 'Workflows', href: '/workflows', icon: CommandLineIcon },
    ],
  },
  {
    id: 'ops',
    name: 'Ops & Governance',
    icon: EyeIcon,
    items: [
      { name: 'Observability', href: '/observability', icon: EyeIcon },
      { name: 'Alerts', href: '/alerts', icon: BellAlertIcon },
      { name: 'Lineage', href: '/lineage', icon: ArrowsRightLeftIcon },
      { name: 'Compliance', href: '/compliance', icon: Cog6ToothIcon },
      { name: 'Audit', href: '/audit', icon: DocumentTextIcon },
      { name: 'Versioning', href: '/versioning', icon: ClockIcon },
    ],
  },
  {
    id: 'admin',
    name: 'Admin',
    icon: Cog6ToothIcon,
    items: [
      { name: 'Users', href: '/users', icon: UserGroupIcon },
      { name: 'API Keys', href: '/api-keys', icon: KeyIcon },
      { name: 'Integrations', href: '/integrations', icon: PuzzlePieceIcon },
      { name: 'Settings', href: '/settings', icon: Cog6ToothIcon },
      { name: 'Billing', href: '/billing', icon: CreditCardIcon },
      { name: 'Support', href: '/support', icon: LifebuoyIcon },
      { name: 'Features', href: '/features', icon: SparklesIcon },
    ],
  },
]

function PinnedRecentLink({
  href,
  label,
  isActive,
  isPinned,
  onTogglePin,
}: {
  href: string
  label: string
  isActive: boolean
  isPinned: boolean
  onTogglePin: () => void
}) {
  return (
    <div className="group/link flex items-center gap-1 rounded-lg px-2 py-1.5">
      <Link href={href} prefetch className="flex-1 min-w-0" onClick={(e) => e.stopPropagation()}>
        <span
          className={cn(
            'block truncate text-sm font-medium transition-colors',
            isActive
              ? 'text-primary'
              : 'text-muted-foreground hover:text-accent-foreground'
          )}
        >
          {label}
        </span>
      </Link>
      <button
        type="button"
        onClick={(e) => {
          e.preventDefault()
          onTogglePin()
        }}
        className="shrink-0 rounded p-1 opacity-0 group-hover/link:opacity-100 focus-visible:opacity-100 transition-opacity hover:bg-accent"
        aria-label={isPinned ? 'Unpin' : 'Pin'}
      >
        {isPinned ? (
          <StarIconSolid className="h-4 w-4 text-primary" />
        ) : (
          <StarIcon className="h-4 w-4 text-muted-foreground" />
        )}
      </button>
    </div>
  )
}

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
            <Link 
              key={item.name} 
              href={item.href}
              prefetch={true}
            >
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
                  <Link 
                    key={item.name} 
                    href={item.href}
                    prefetch={true}
                  >
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
              <Link 
                key={item.name} 
                href={item.href} 
                title={item.name}
                prefetch={true}
              >
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
  const {
    sidebarCollapsed,
    toggleSidebar,
    recentPaths,
    pinnedPaths,
    togglePinnedPath,
  } = useAppStore()

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
            <Link href="/" className="flex items-center gap-2 flex-shrink-0">
              <div className="w-10 h-10 rounded-lg flex items-center justify-center overflow-hidden bg-white dark:bg-slate-800">
                <img 
                  src="/logo.png" 
                  alt="NeuronIP" 
                  className="w-full h-full object-contain p-1"
                />
              </div>
              {!sidebarCollapsed && (
                <motion.h2
                  className="text-xl font-bold text-foreground"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                >
                  NeuronIP
                </motion.h2>
              )}
            </Link>
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

          {/* Pinned & Recents (Notion-like) */}
          {(pinnedPaths.length > 0 || recentPaths.length > 0) && !sidebarCollapsed && (
            <div className="shrink-0 space-y-2 border-b border-border px-3 py-3">
              {pinnedPaths.length > 0 && (
                <div className="space-y-1">
                  <p className="px-3 py-1 text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    Pinned
                  </p>
                  {pinnedPaths.map((path) => {
                    const meta = getRouteMeta(path) ?? { title: path, path }
                    const isActive = pathname === path
                    return (
                      <PinnedRecentLink
                        key={path}
                        href={path}
                        label={meta.title}
                        isActive={isActive}
                        isPinned
                        onTogglePin={() => togglePinnedPath(path)}
                      />
                    )
                  })}
                </div>
              )}
              {recentPaths.length > 0 && (
                <div className="space-y-1">
                  <p className="px-3 py-1 text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    Recents
                  </p>
                  {recentPaths
                    .filter((path) => path !== pathname && !pinnedPaths.includes(path))
                    .slice(0, 5)
                    .map((path) => {
                      const meta = getRouteMeta(path) ?? { title: path, path }
                      const isActive = pathname === path
                      return (
                        <PinnedRecentLink
                          key={path}
                          href={path}
                          label={meta.title}
                          isActive={isActive}
                          isPinned={pinnedPaths.includes(path)}
                          onTogglePin={() => togglePinnedPath(path)}
                        />
                      )
                    })}
                </div>
              )}
            </div>
          )}

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

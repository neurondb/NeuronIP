/**
 * Page archetypes for consistent layout, empty/loading/error states, and UX patterns.
 * Used by PageTemplate and page-level components.
 */
export type PageArchetype =
  | 'dashboard'   // Overview: metrics, charts, activity
  | 'search'      // Search/explore: semantic, warehouse, catalog, knowledge-graph
  | 'builder'     // Visual builder: workflows, notion blocks
  | 'list-detail' // Admin/list: users, api-keys, integrations, settings, audit, etc.

export interface RouteMeta {
  path: string
  archetype: PageArchetype
  title: string
  description?: string
}

/** Map dashboard route path → archetype and default copy */
export const ROUTE_ARCHETYPE_MAP: Record<string, RouteMeta> = {
  '/': { path: '/', archetype: 'dashboard', title: 'Dashboard', description: 'Overview of searches, queries, and compliance' },
  '/analytics': { path: '/analytics', archetype: 'dashboard', title: 'Analytics', description: 'Usage and performance analytics' },
  '/observability': { path: '/observability', archetype: 'dashboard', title: 'Observability', description: 'Monitoring and traces' },
  '/compliance': { path: '/compliance', archetype: 'dashboard', title: 'Compliance', description: 'Policies and compliance checks' },
  '/quality': { path: '/quality', archetype: 'dashboard', title: 'Data Quality', description: 'Quality checks and scores' },
  '/billing': { path: '/billing', archetype: 'dashboard', title: 'Billing', description: 'Usage and billing' },

  '/semantic': { path: '/semantic', archetype: 'search', title: 'Semantic Search', description: 'Search and chat over your data' },
  '/warehouse': { path: '/warehouse', archetype: 'search', title: 'Data Warehouse', description: 'Query and explore your warehouse' },
  '/catalog': { path: '/catalog', archetype: 'search', title: 'Data Catalog', description: 'Datasets and metadata' },
  '/knowledge-graph': { path: '/knowledge-graph', archetype: 'search', title: 'Knowledge Graph', description: 'Entities and relationships' },
  '/data-sources': { path: '/data-sources', archetype: 'search', title: 'Data Sources', description: 'Connectors and ingestion' },
  '/metrics': { path: '/metrics', archetype: 'search', title: 'Metrics', description: 'Business metrics' },

  '/workflows': { path: '/workflows', archetype: 'list-detail', title: 'Workflows', description: 'Workflow runs and definitions' },
  '/workflows/builder': { path: '/workflows/builder', archetype: 'builder', title: 'Workflow Builder', description: 'Create and edit workflows visually' },
  '/notion-ui': { path: '/notion-ui', archetype: 'builder', title: 'Notion UI', description: 'Blocks and databases' },

  '/users': { path: '/users', archetype: 'list-detail', title: 'Users', description: 'User accounts and permissions' },
  '/api-keys': { path: '/api-keys', archetype: 'list-detail', title: 'API Keys', description: 'API key management' },
  '/integrations': { path: '/integrations', archetype: 'list-detail', title: 'Integrations', description: 'Connected apps and webhooks' },
  '/settings': { path: '/settings', archetype: 'list-detail', title: 'Settings', description: 'Organization and preferences' },
  '/audit': { path: '/audit', archetype: 'list-detail', title: 'Audit Logs', description: 'Activity and audit trail' },
  '/alerts': { path: '/alerts', archetype: 'list-detail', title: 'Alerts', description: 'Alert rules and history' },
  '/lineage': { path: '/lineage', archetype: 'list-detail', title: 'Data Lineage', description: 'Lineage visualization' },
  '/versioning': { path: '/versioning', archetype: 'list-detail', title: 'Versioning', description: 'Data versions' },
  '/support': { path: '/support', archetype: 'list-detail', title: 'Support', description: 'Tickets and help' },
  '/features': { path: '/features', archetype: 'list-detail', title: 'Features', description: 'Feature overview' },
  '/why-neuronip': { path: '/why-neuronip', archetype: 'list-detail', title: 'Why NeuronIP', description: 'Product overview' },
  '/agents': { path: '/agents', archetype: 'list-detail', title: 'Agent Hub', description: 'AI agents' },
  '/models': { path: '/models', archetype: 'list-detail', title: 'Models', description: 'ML models and governance' },
}

/** Get route meta for path (exact or nearest parent). */
export function getRouteMeta(pathname: string): RouteMeta | undefined {
  if (ROUTE_ARCHETYPE_MAP[pathname]) return ROUTE_ARCHETYPE_MAP[pathname]
  // Try without trailing slash
  const normalized = pathname.replace(/\/$/, '') || '/'
  if (ROUTE_ARCHETYPE_MAP[normalized]) return ROUTE_ARCHETYPE_MAP[normalized]
  // Match longest prefix (e.g. /workflows/builder → workflows/builder)
  const segments = normalized.split('/').filter(Boolean)
  for (let i = segments.length; i >= 1; i--) {
    const candidate = '/' + segments.slice(0, i).join('/')
    if (ROUTE_ARCHETYPE_MAP[candidate]) return ROUTE_ARCHETYPE_MAP[candidate]
  }
  return undefined
}

/** Get archetype for path. */
export function getArchetype(pathname: string): PageArchetype {
  return getRouteMeta(pathname)?.archetype ?? 'list-detail'
}

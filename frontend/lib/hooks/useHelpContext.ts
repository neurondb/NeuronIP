'use client'

import { useState, useCallback, useEffect } from 'react'
import { usePathname } from 'next/navigation'

export interface HelpContext {
  page: string
  section?: string
  showHelp: () => void
  hideHelp: () => void
  isHelpVisible: boolean
  getHelpContent: (key: string) => string | undefined
  getHelpLink: (key: string) => string | undefined
}

// Help content mapping by page/section
const helpContent: Record<string, Record<string, { content: string; link?: string }>> = {
  dashboard: {
    overview: {
      content: 'View your platform activity, quick actions, and key metrics here.',
      link: '/docs/getting-started',
    },
  },
  semantic: {
    search: {
      content: 'Use natural language to search your knowledge base semantically. Results are ranked by meaning, not just keywords.',
      link: '/docs/features/semantic-search',
    },
    collections: {
      content: 'Manage collections of documents for semantic search. Each collection can have different embedding models.',
      link: '/docs/features/semantic-search',
    },
  },
  warehouse: {
    query: {
      content: 'Ask natural language questions or write SQL queries. The system will generate SQL, execute it safely, and provide visualizations.',
      link: '/docs/features/warehouse-qa',
    },
    schemas: {
      content: 'Manage warehouse schemas. The system automatically discovers schemas and can help generate SQL queries.',
      link: '/docs/features/warehouse-qa',
    },
  },
  workflows: {
    builder: {
      content: 'Build multi-step workflows with agents. Drag and drop nodes to create your workflow DAG.',
      link: '/docs/features/agent-workflows',
    },
    execution: {
      content: 'Monitor workflow executions. View logs, retry failed steps, and debug issues.',
      link: '/docs/features/agent-workflows',
    },
  },
  agents: {
    create: {
      content: 'Create AI agents with custom prompts, tools, and memory. Agents can execute workflows and interact with data.',
      link: '/docs/integrations/neuronagent',
    },
    manage: {
      content: 'Manage your agents, view their capabilities, and monitor their performance.',
      link: '/docs/integrations/neuronagent',
    },
  },
  'data-sources': {
    setup: {
      content: 'Connect external data sources like databases, APIs, or file storage. Configure sync schedules and transformations.',
      link: '/docs/development/setup',
    },
    connectors: {
      content: 'Choose from available connectors for various data sources. Each connector has specific configuration options.',
      link: '/docs/development/setup',
    },
  },
  compliance: {
    policies: {
      content: 'Create compliance policies to match against data. Policies use semantic matching and anomaly detection.',
      link: '/docs/features/compliance',
    },
    checks: {
      content: 'Run compliance checks on your data sources and view violation reports.',
      link: '/docs/features/compliance',
    },
  },
  support: {
    tickets: {
      content: 'Manage support tickets with AI assistance. The system retrieves similar cases and suggests solutions.',
      link: '/docs/features/support-memory',
    },
  },
}

// Documentation links by feature
const docLinks: Record<string, string> = {
  'semantic-search': '/docs/features/semantic-search',
  'warehouse-qa': '/docs/features/warehouse-qa',
  'agent-workflows': '/docs/features/agent-workflows',
  compliance: '/docs/features/compliance',
  'support-memory': '/docs/features/support-memory',
  api: '/docs/api/overview',
  architecture: '/docs/architecture/overview',
  'getting-started': '/docs/getting-started',
}

export function useHelpContext(section?: string): HelpContext {
  const pathname = usePathname()
  const [isHelpVisible, setIsHelpVisible] = useState(false)

  // Extract page name from pathname
  const page = pathname.split('/').pop() || 'dashboard'
  const normalizedPage = page === 'dashboard' && pathname === '/dashboard' ? 'dashboard' : page

  const showHelp = useCallback(() => {
    setIsHelpVisible(true)
  }, [])

  const hideHelp = useCallback(() => {
    setIsHelpVisible(false)
  }, [])

  const getHelpContent = useCallback(
    (key: string): string | undefined => {
      const pageHelp = helpContent[normalizedPage]
      if (!pageHelp) return undefined

      const helpKey = section || key
      const help = pageHelp[helpKey] || pageHelp[key]
      return help?.content
    },
    [normalizedPage, section]
  )

  const getHelpLink = useCallback(
    (key: string): string | undefined => {
      const pageHelp = helpContent[normalizedPage]
      if (!pageHelp) return docLinks[key] || docLinks[normalizedPage]

      const helpKey = section || key
      const help = pageHelp[helpKey] || pageHelp[key]
      return help?.link || docLinks[key] || docLinks[normalizedPage]
    },
    [normalizedPage, section]
  )

  // Auto-hide help when pathname changes
  useEffect(() => {
    setIsHelpVisible(false)
  }, [pathname])

  return {
    page: normalizedPage,
    section,
    showHelp,
    hideHelp,
    isHelpVisible,
    getHelpContent,
    getHelpLink,
  }
}

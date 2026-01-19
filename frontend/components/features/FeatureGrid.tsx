'use client'

import { motion } from 'framer-motion'
import FeatureCard from './FeatureCard'
import {
  HomeIcon,
  MagnifyingGlassIcon,
  CubeIcon,
  CommandLineIcon,
  Cog6ToothIcon,
  LifebuoyIcon,
  CpuChipIcon,
  CircleStackIcon,
  BellAlertIcon,
  KeyIcon,
  UserGroupIcon,
  ChartBarIcon,
  ServerIcon,
  UserCircleIcon,
  EyeIcon,
  ArrowsRightLeftIcon,
  DocumentTextIcon,
  CreditCardIcon,
  PuzzlePieceIcon,
  ClockIcon,
  FolderIcon,
} from '@heroicons/react/24/outline'
import { staggerContainer } from '@/lib/animations/variants'

const features = [
  {
    title: 'Dashboard',
    description: 'Overview of your NeuronIP platform activity',
    href: '/dashboard',
    icon: <HomeIcon className="h-6 w-6" />,
    color: 'text-blue-600',
  },
  {
    title: 'Semantic Search',
    description: 'AI-powered semantic search and RAG interface',
    href: '/dashboard/semantic',
    icon: <MagnifyingGlassIcon className="h-6 w-6" />,
    color: 'text-blue-600',
  },
  {
    title: 'Warehouse',
    description: 'Query your data warehouse with natural language or SQL',
    href: '/dashboard/warehouse',
    icon: <CubeIcon className="h-6 w-6" />,
    color: 'text-green-600',
  },
  {
    title: 'Workflows',
    description: 'Execute and monitor your automated workflows',
    href: '/dashboard/workflows',
    icon: <CommandLineIcon className="h-6 w-6" />,
    color: 'text-purple-600',
  },
  {
    title: 'Compliance',
    description: 'Monitor compliance and detect anomalies',
    href: '/dashboard/compliance',
    icon: <Cog6ToothIcon className="h-6 w-6" />,
    color: 'text-orange-600',
  },
  {
    title: 'Support',
    description: 'Manage helpdesk tickets and customer support',
    href: '/dashboard/support',
    icon: <LifebuoyIcon className="h-6 w-6" />,
    color: 'text-yellow-600',
  },
  {
    title: 'Models',
    description: 'Manage AI models and run inferences',
    href: '/dashboard/models',
    icon: <CpuChipIcon className="h-6 w-6" />,
    color: 'text-indigo-600',
  },
  {
    title: 'Knowledge Graph',
    description: 'Explore entity relationships and graph traversal',
    href: '/dashboard/knowledge-graph',
    icon: <CircleStackIcon className="h-6 w-6" />,
    color: 'text-teal-600',
  },
  {
    title: 'Alerts',
    description: 'Monitor system alerts and manage alert rules',
    href: '/dashboard/alerts',
    icon: <BellAlertIcon className="h-6 w-6" />,
    color: 'text-red-600',
  },
  {
    title: 'API Keys',
    description: 'Manage API keys and rate limits',
    href: '/dashboard/api-keys',
    icon: <KeyIcon className="h-6 w-6" />,
    color: 'text-gray-600',
  },
  {
    title: 'Users',
    description: 'Manage user accounts and permissions',
    href: '/dashboard/users',
    icon: <UserGroupIcon className="h-6 w-6" />,
    color: 'text-pink-600',
  },
  {
    title: 'Analytics',
    description: 'Comprehensive analytics dashboard',
    href: '/dashboard/analytics',
    icon: <ChartBarIcon className="h-6 w-6" />,
    color: 'text-cyan-600',
  },
  {
    title: 'Data Sources',
    description: 'Connectors for PostgreSQL, S3, APIs, and SaaS tools',
    href: '/dashboard/data-sources',
    icon: <ServerIcon className="h-6 w-6" />,
    color: 'text-blue-600',
  },
  {
    title: 'Metrics',
    description: 'Define KPIs, dimensions, and business terms for semantic layer',
    href: '/dashboard/metrics',
    icon: <ChartBarIcon className="h-6 w-6" />,
    color: 'text-green-600',
  },
  {
    title: 'Agent Hub',
    description: 'Manage and deploy AI agents with performance tracking',
    href: '/dashboard/agents',
    icon: <UserCircleIcon className="h-6 w-6" />,
    color: 'text-purple-600',
  },
  {
    title: 'Observability',
    description: 'Query performance, latency, cost, and system logs',
    href: '/dashboard/observability',
    icon: <EyeIcon className="h-6 w-6" />,
    color: 'text-orange-600',
  },
  {
    title: 'Data Lineage',
    description: 'Track data transformations and visual impact analysis',
    href: '/dashboard/lineage',
    icon: <ArrowsRightLeftIcon className="h-6 w-6" />,
    color: 'text-teal-600',
  },
  {
    title: 'Audit',
    description: 'Full history of user actions, agent actions, and compliance trail',
    href: '/dashboard/audit',
    icon: <DocumentTextIcon className="h-6 w-6" />,
    color: 'text-red-600',
  },
  {
    title: 'Billing',
    description: 'Track seats, API calls, queries, embeddings, and monetization',
    href: '/dashboard/billing',
    icon: <CreditCardIcon className="h-6 w-6" />,
    color: 'text-yellow-600',
  },
  {
    title: 'Integrations',
    description: 'Slack, Teams, CRM, ERP, ticketing, email, and webhooks',
    href: '/dashboard/integrations',
    icon: <PuzzlePieceIcon className="h-6 w-6" />,
    color: 'text-indigo-600',
  },
  {
    title: 'Versioning',
    description: 'Version control for models, embeddings, workflows, and metrics',
    href: '/dashboard/versioning',
    icon: <ClockIcon className="h-6 w-6" />,
    color: 'text-pink-600',
  },
  {
    title: 'Data Catalog',
    description: 'Browse datasets, fields, owners, and semantic discovery',
    href: '/dashboard/catalog',
    icon: <FolderIcon className="h-6 w-6" />,
    color: 'text-cyan-600',
  },
]

export default function FeatureGrid() {
  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 sm:gap-4"
    >
      {features.map((feature) => (
        <FeatureCard key={feature.title} {...feature} />
      ))}
    </motion.div>
  )
}

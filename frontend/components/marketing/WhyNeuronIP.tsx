'use client'

import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { 
  BoltIcon, 
  LockClosedIcon, 
  ServerIcon, 
  DocumentTextIcon, 
  CheckCircleIcon, 
  CurrencyDollarIcon 
} from '@heroicons/react/24/outline'
import { slideUp } from '@/lib/animations/variants'

interface Claim {
  id: string
  title: string
  icon: React.ReactNode
  metric: string
  description: string
  details: string[]
  color: string
}

const claims: Claim[] = [
  {
    id: 'latency',
    title: 'Sub-100ms Latency',
    icon: <BoltIcon className="h-8 w-8" />,
    metric: '< 100ms semantic search, < 500ms warehouse queries',
    description: 'Lightning-fast responses that feel instant. Semantic search completes in under 100ms, warehouse queries in under 500ms.',
    details: [
      'Sub-100ms semantic search response time',
      'Sub-500ms warehouse query execution',
      'Optimized vector similarity search',
      'Efficient caching and indexing',
      'p95 latency consistently under targets'
    ],
    color: 'text-yellow-500'
  },
  {
    id: 'privacy',
    title: 'Complete Privacy',
    icon: <LockClosedIcon className="h-8 w-8" />,
    metric: 'Data never leaves your infrastructure',
    description: 'Your data stays on your servers. No cloud dependencies, no data exfiltration, complete control.',
    details: [
      '100% on-premise deployment option',
      'No data sent to external services',
      'End-to-end encryption',
      'Self-hosted vector database',
      'GDPR and SOC 2 compliant architecture'
    ],
    color: 'text-green-500'
  },
  {
    id: 'on-prem',
    title: 'Full On-Premise',
    icon: <ServerIcon className="h-8 w-8" />,
    metric: 'Zero cloud dependencies',
    description: 'Deploy entirely on your infrastructure. No external API calls, no cloud lock-in, complete sovereignty.',
    details: [
      'Runs entirely on your servers',
      'No external API dependencies',
      'Works in air-gapped environments',
      'Full control over infrastructure',
      'No vendor lock-in'
    ],
    color: 'text-blue-500'
  },
  {
    id: 'auditability',
    title: 'Complete Auditability',
    icon: <DocumentTextIcon className="h-8 w-8" />,
    metric: 'Every action logged and traceable',
    description: 'Complete audit trail for compliance, security, and debugging. Every query, every action, every decision is logged.',
    details: [
      'Complete audit trail for all operations',
      'Request ID propagation across systems',
      'Query history and explanations',
      'Agent action logging',
      'Compliance-ready audit logs'
    ],
    color: 'text-purple-500'
  },
  {
    id: 'deterministic',
    title: 'Deterministic SQL',
    icon: <CheckCircleIcon className="h-8 w-8" />,
    metric: 'All queries are explainable and reproducible',
    description: 'Every SQL query is deterministic, explainable, and auditable. No black-box AI, just transparent, reproducible results.',
    details: [
      'Deterministic SQL generation',
      'Full query explanations',
      'Reproducible results',
      'Query validation and sanitization',
      'No unpredictable AI behavior'
    ],
    color: 'text-indigo-500'
  },
  {
    id: 'cost',
    title: '10x Cost Savings',
    icon: <CurrencyDollarIcon className="h-8 w-8" />,
    metric: '10x cheaper than cloud alternatives',
    description: 'Self-hosted means no per-query costs, no API call fees, no data transfer charges. Predictable, transparent pricing.',
    details: [
      'No per-query charges',
      'No API call fees',
      'No data transfer costs',
      'Predictable infrastructure costs',
      'Open-source core, enterprise features'
    ],
    color: 'text-emerald-500'
  }
]

export default function WhyNeuronIP() {
  return (
    <div className="space-y-6 sm:space-y-8">
      {/* Header */}
      <div className="text-center space-y-4">
        <h1 className="text-3xl sm:text-4xl font-bold text-foreground">
          Why NeuronIP?
        </h1>
        <p className="text-lg text-muted-foreground max-w-3xl mx-auto">
          Six hard claims that make NeuronIP the enterprise-grade choice for AI-native intelligence
        </p>
      </div>

      {/* Claims Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
        {claims.map((claim, index) => (
          <motion.div
            key={claim.id}
            variants={slideUp}
            transition={{ delay: index * 0.1 }}
          >
            <Card className="h-full hover:shadow-lg transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between mb-2">
                  <div className={`${claim.color} flex-shrink-0`}>
                    {claim.icon}
                  </div>
                </div>
                <CardTitle className="text-xl">{claim.title}</CardTitle>
                <CardDescription className="text-base font-semibold text-foreground mt-2">
                  {claim.metric}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-4">
                  {claim.description}
                </p>
                <ul className="space-y-2">
                  {claim.details.map((detail, idx) => (
                    <li key={idx} className="flex items-start text-sm">
                      <CheckCircleIcon className="h-4 w-4 text-green-500 mr-2 mt-0.5 flex-shrink-0" />
                      <span className="text-muted-foreground">{detail}</span>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>

      {/* Cost Calculator Section */}
      <motion.div variants={slideUp} className="mt-8">
        <Card>
          <CardHeader>
            <CardTitle>Cost Calculator</CardTitle>
            <CardDescription>
              Compare NeuronIP costs vs. cloud alternatives
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="p-4 bg-muted rounded-lg">
                  <div className="text-sm text-muted-foreground mb-1">Cloud Alternative</div>
                  <div className="text-2xl font-bold">$10,000/mo</div>
                  <div className="text-xs text-muted-foreground mt-1">
                    For 1M queries, 100GB data
                  </div>
                </div>
                <div className="p-4 bg-primary/10 rounded-lg">
                  <div className="text-sm text-muted-foreground mb-1">NeuronIP</div>
                  <div className="text-2xl font-bold text-primary">$1,000/mo</div>
                  <div className="text-xs text-muted-foreground mt-1">
                    Same workload, self-hosted
                  </div>
                </div>
                <div className="p-4 bg-green-500/10 rounded-lg">
                  <div className="text-sm text-muted-foreground mb-1">Savings</div>
                  <div className="text-2xl font-bold text-green-500">90%</div>
                  <div className="text-xs text-muted-foreground mt-1">
                    $9,000/month saved
                  </div>
                </div>
              </div>
              <p className="text-sm text-muted-foreground">
                *Costs based on typical enterprise workloads. Actual savings may vary based on infrastructure and usage patterns.
              </p>
            </div>
          </CardContent>
        </Card>
      </motion.div>

      {/* CTA Section */}
      <motion.div variants={slideUp} className="text-center space-y-4 pt-4">
        <h2 className="text-2xl font-bold text-foreground">
          Ready to experience NeuronIP?
        </h2>
        <p className="text-muted-foreground">
          Get started with our demo or contact us for a custom evaluation
        </p>
        <div className="flex gap-4 justify-center">
          <a
            href="/dashboard"
            className="px-6 py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:bg-primary/90 transition-colors"
          >
            Try Demo
          </a>
          <a
            href="/dashboard/features"
            className="px-6 py-3 bg-muted text-foreground rounded-lg font-medium hover:bg-muted/80 transition-colors"
          >
            View Features
          </a>
        </div>
      </motion.div>
    </div>
  )
}

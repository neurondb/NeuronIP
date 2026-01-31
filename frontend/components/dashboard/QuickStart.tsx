'use client'

import {
  MagnifyingGlassIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  ArrowRightIcon,
} from '@heroicons/react/24/outline'
import { motion } from 'framer-motion'
import Link from 'next/link'

import Button from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { staggerContainer, slideUp, transition } from '@/lib/animations/variants'
import { microcopy } from '@/lib/copy/microcopy'

const quickStartItems = [
  {
    title: 'Search your data',
    description: 'Find information by meaning, not just keywords',
    icon: MagnifyingGlassIcon,
    href: '/semantic',
    color: 'text-blue-600 dark:text-blue-400',
    bgColor: 'bg-blue-50 dark:bg-blue-950/30',
  },
  {
    title: 'Ask a question',
    description: 'Get answers about your data in plain English',
    icon: ChatBubbleLeftRightIcon,
    href: '/warehouse',
    color: 'text-green-600 dark:text-green-400',
    bgColor: 'bg-green-50 dark:bg-green-950/30',
  },
  {
    title: 'Explore features',
    description: 'See what NeuronIP can do for you',
    icon: SparklesIcon,
    href: '/features',
    color: 'text-purple-600 dark:text-purple-400',
    bgColor: 'bg-purple-50 dark:bg-purple-950/30',
  },
]

export default function QuickStart() {
  return (
    <Card className="border-2 border-primary/20 bg-gradient-to-br from-primary/5 to-transparent">
      <CardContent className="p-6 sm:p-8">
        <div className="mb-6">
          <h3 className="text-xl sm:text-2xl font-bold mb-2">
            {microcopy.dashboard.quickStart.title}
          </h3>
          <p className="text-sm sm:text-base text-muted-foreground">
            {microcopy.dashboard.quickStart.subtitle}
          </p>
        </div>
        <motion.div
          variants={staggerContainer}
          initial="hidden"
          animate="visible"
          className="grid grid-cols-1 sm:grid-cols-3 gap-4"
        >
          {quickStartItems.map((item, index) => {
            const Icon = item.icon
            return (
              <motion.div
                key={item.title}
                variants={slideUp}
                transition={{ ...transition, delay: index * 0.1 }}
              >
                <Link href={item.href}>
                  <div
                    className={`
                      group relative p-6 rounded-xl border border-border
                      hover:border-primary/50 transition-all cursor-pointer
                      ${item.bgColor}
                      hover:shadow-lg hover:scale-[1.02]
                    `}
                  >
                    <div className="flex items-start gap-4">
                      <div className={`p-3 rounded-lg bg-background/80 ${item.bgColor}`}>
                        <Icon className={`h-6 w-6 ${item.color}`} />
                      </div>
                      <div className="flex-1 min-w-0">
                        <h4 className="font-semibold mb-1 group-hover:text-primary transition-colors">
                          {item.title}
                        </h4>
                        <p className="text-sm text-muted-foreground leading-relaxed">
                          {item.description}
                        </p>
                      </div>
                    </div>
                    <ArrowRightIcon className="absolute top-6 right-6 h-5 w-5 text-muted-foreground group-hover:text-primary group-hover:translate-x-1 transition-all" />
                  </div>
                </Link>
              </motion.div>
            )
          })}
        </motion.div>
      </CardContent>
    </Card>
  )
}

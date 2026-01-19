'use client'

import Link from 'next/link'
import { motion } from 'framer-motion'
import {
  MagnifyingGlassIcon,
  CubeIcon,
  CommandLineIcon,
  Cog6ToothIcon,
} from '@heroicons/react/24/outline'
import { Card } from '@/components/ui/Card'
import { cn } from '@/lib/utils/cn'
import { staggerContainer, slideUp, transition } from '@/lib/animations/variants'

const actions = [
  {
    name: 'Semantic Search',
    href: '/dashboard/semantic',
    icon: MagnifyingGlassIcon,
    color: 'text-blue-600 dark:text-blue-400',
  },
  {
    name: 'Warehouse Query',
    href: '/dashboard/warehouse',
    icon: CubeIcon,
    color: 'text-green-600 dark:text-green-400',
  },
  {
    name: 'Run Workflow',
    href: '/dashboard/workflows',
    icon: CommandLineIcon,
    color: 'text-purple-600 dark:text-purple-400',
  },
  {
    name: 'Compliance Check',
    href: '/dashboard/compliance',
    icon: Cog6ToothIcon,
    color: 'text-orange-600 dark:text-orange-400',
  },
]

export default function QuickActions() {
  return (
    <Card>
      <div className="p-4 sm:p-5">
        <h3 className="text-base sm:text-lg font-semibold mb-3">Quick Actions</h3>
        <motion.div
          variants={staggerContainer}
          initial="hidden"
          animate="visible"
          className="grid grid-cols-2 sm:grid-cols-4 gap-2 sm:gap-3"
        >
          {actions.map((action, index) => {
            const Icon = action.icon
            return (
              <Link key={action.name} href={action.href}>
                <motion.div
                  variants={slideUp}
                  transition={{ ...transition, delay: index * 0.1 }}
                  className={cn(
                    'flex flex-col items-center justify-center gap-1.5 sm:gap-2 p-3 sm:p-4 rounded-lg border border-border',
                    'hover:bg-accent hover:border-primary transition-all cursor-pointer',
                    'min-h-[80px] sm:min-h-[90px]'
                  )}
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <Icon className={cn('h-6 w-6 sm:h-7 sm:w-7', action.color)} />
                  <span className="text-xs sm:text-sm font-medium text-center leading-tight">{action.name}</span>
                </motion.div>
              </Link>
            )
          })}
        </motion.div>
      </div>
    </Card>
  )
}

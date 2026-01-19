'use client'

import Link from 'next/link'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface FeatureCardProps {
  title: string
  description: string
  href: string
  icon: ReactNode
  color?: string
}

export default function FeatureCard({ title, description, href, icon, color }: FeatureCardProps) {
  return (
    <Link href={href}>
      <motion.div variants={slideUp} initial="hidden" animate="visible" transition={transition}>
        <Card hover className={cn('h-full', color)}>
          <CardHeader>
            <div className="flex items-center gap-3 mb-2">
              <div className="h-10 w-10 flex items-center justify-center text-primary">{icon}</div>
              <CardTitle className="text-base font-semibold">{title}</CardTitle>
            </div>
            <CardDescription className="text-xs mt-1">{description}</CardDescription>
          </CardHeader>
        </Card>
      </motion.div>
    </Link>
  )
}

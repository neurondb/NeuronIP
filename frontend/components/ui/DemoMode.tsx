'use client'

import { SparklesIcon, XMarkIcon } from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useState, ReactNode } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { microcopy } from '@/lib/copy/microcopy'

interface DemoModeProps {
  children: ReactNode
  title?: string
  description?: string
  onTryIt?: () => void
  demoContent?: ReactNode
}

export default function DemoMode({
  children,
  title = 'Try it now',
  description = 'See how this works with a live example',
  onTryIt,
  demoContent,
}: DemoModeProps) {
  const [isDemoActive, setIsDemoActive] = useState(false)

  const handleTryIt = () => {
    if (onTryIt) {
      onTryIt()
    }
    setIsDemoActive(true)
  }

  if (isDemoActive && demoContent) {
    return (
      <AnimatePresence>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -20 }}
          className="relative"
        >
          <Card className="border-2 border-primary/30 bg-gradient-to-br from-primary/5 to-transparent">
            <CardContent className="p-6">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-2">
                  <SparklesIcon className="h-5 w-5 text-primary" />
                  <span className="text-sm font-medium text-primary">Demo Mode</span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setIsDemoActive(false)}
                >
                  <XMarkIcon className="h-4 w-4" />
                </Button>
              </div>
              {demoContent}
            </CardContent>
          </Card>
        </motion.div>
      </AnimatePresence>
    )
  }

  return (
    <div className="relative">
      {children}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm rounded-lg"
      >
        <Card className="max-w-md mx-4">
          <CardContent className="p-6 text-center">
            <SparklesIcon className="h-12 w-12 text-primary mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">{title}</h3>
            <p className="text-sm text-muted-foreground mb-6">{description}</p>
            <Button onClick={handleTryIt} size="lg" className="w-full">
              {microcopy.actions.seeItInAction}
            </Button>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}

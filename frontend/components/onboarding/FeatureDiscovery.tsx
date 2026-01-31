'use client'

import { XMarkIcon, SparklesIcon } from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useState, useEffect } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'

interface FeatureHint {
  id: string
  title: string
  description: string
  action?: {
    label: string
    onClick: () => void
  }
  dismissible?: boolean
}

interface FeatureDiscoveryProps {
  featureId: string
  title: string
  description: string
  action?: {
    label: string
    onClick: () => void
  }
  showOnce?: boolean
}

const DISCOVERY_STORAGE_KEY = 'neuronip_feature_discovery'

export default function FeatureDiscovery({
  featureId,
  title,
  description,
  action,
  showOnce = true,
}: FeatureDiscoveryProps) {
  const [isVisible, setIsVisible] = useState(false)

  useEffect(() => {
    if (typeof window === 'undefined') return

    const discovered = localStorage.getItem(DISCOVERY_STORAGE_KEY)
    const discoveredFeatures = discovered ? JSON.parse(discovered) : {}

    if (showOnce && discoveredFeatures[featureId]) {
      return
    }

    // Show after a short delay for better UX
    const timer = setTimeout(() => {
      setIsVisible(true)
    }, 1000)

    return () => clearTimeout(timer)
  }, [featureId, showOnce])

  const handleDismiss = () => {
    setIsVisible(false)
    if (showOnce && typeof window !== 'undefined') {
      const discovered = localStorage.getItem(DISCOVERY_STORAGE_KEY)
      const discoveredFeatures = discovered ? JSON.parse(discovered) : {}
      discoveredFeatures[featureId] = true
      localStorage.setItem(DISCOVERY_STORAGE_KEY, JSON.stringify(discoveredFeatures))
    }
  }

  if (!isVisible) return null

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -20 }}
        className="fixed bottom-6 right-6 z-50 max-w-sm"
      >
        <Card className="border-2 border-primary/30 bg-gradient-to-br from-primary/5 to-transparent shadow-lg">
          <CardContent className="p-4">
            <div className="flex items-start gap-3">
              <div className="p-2 rounded-lg bg-primary/10">
                <SparklesIcon className="h-5 w-5 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <h4 className="font-semibold text-sm mb-1">{title}</h4>
                <p className="text-xs text-muted-foreground mb-3">{description}</p>
                {action && (
                  <Button
                    onClick={() => {
                      action.onClick()
                      handleDismiss()
                    }}
                    size="sm"
                    className="w-full"
                  >
                    {action.label}
                  </Button>
                )}
              </div>
              <button
                onClick={handleDismiss}
                className="text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
              >
                <XMarkIcon className="h-4 w-4" />
              </button>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </AnimatePresence>
  )
}

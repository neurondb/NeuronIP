'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { showToast } from '@/components/ui/Toast'
import { Badge } from '@/components/ui/Badge'

interface TimeGrain {
  id: string
  name: string
  display_name: string
  duration: string
  is_default: boolean
}

interface TimeGrainSelectorProps {
  metricId: string
  selectedTimeGrain?: string
  onTimeGrainChange?: (timeGrainId: string) => void
  allowMultiple?: boolean
}

export default function TimeGrainSelector({
  metricId,
  selectedTimeGrain,
  onTimeGrainChange,
  allowMultiple = false,
}: TimeGrainSelectorProps) {
  const [timeGrains, setTimeGrains] = useState<TimeGrain[]>([])
  const [metricTimeGrains, setMetricTimeGrains] = useState<string[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchTimeGrains()
    if (metricId) {
      fetchMetricTimeGrains()
    }
  }, [metricId])

  const fetchTimeGrains = async () => {
    try {
      const response = await fetch('/api/v1/metrics/time-grains')
      if (!response.ok) throw new Error('Failed to fetch time grains')
      const data = await response.json()
      setTimeGrains(data)
    } catch (error) {
      console.error('Error fetching time grains:', error)
    }
  }

  const fetchMetricTimeGrains = async () => {
    try {
      const response = await fetch(`/api/v1/metrics/${metricId}/time-grains`)
      if (!response.ok) throw new Error('Failed to fetch metric time grains')
      const data = await response.json()
      setMetricTimeGrains(data.map((tg: TimeGrain) => tg.id))
    } catch (error) {
      console.error('Error fetching metric time grains:', error)
    }
  }

  const handleTimeGrainToggle = async (timeGrainId: string, isDefault: boolean = false) => {
    if (!metricId) {
      showToast('Please select a metric first', 'warning')
      return
    }

    setLoading(true)
    try {
      const isSelected = metricTimeGrains.includes(timeGrainId)
      
      if (allowMultiple) {
        if (isSelected) {
          // Remove time grain
          const response = await fetch(
            `/api/v1/metrics/${metricId}/time-grains/${timeGrainId}`,
            { method: 'DELETE' }
          )
          if (!response.ok) throw new Error('Failed to remove time grain')
          setMetricTimeGrains(metricTimeGrains.filter(id => id !== timeGrainId))
        } else {
          // Add time grain
          const response = await fetch(`/api/v1/metrics/${metricId}/time-grains`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ time_grain_id: timeGrainId, is_default: isDefault }),
          })
          if (!response.ok) throw new Error('Failed to add time grain')
          setMetricTimeGrains([...metricTimeGrains, timeGrainId])
        }
      } else {
        // Single selection mode
        const response = await fetch(`/api/v1/metrics/${metricId}/time-grains`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ time_grain_id: timeGrainId, is_default: true }),
        })
        if (!response.ok) throw new Error('Failed to set time grain')
        setMetricTimeGrains([timeGrainId])
        onTimeGrainChange?.(timeGrainId)
      }
      
      showToast('Time grain updated', 'success')
      fetchMetricTimeGrains()
    } catch (error: any) {
      showToast(error.message || 'Failed to update time grain', 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Time Grains</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {timeGrains.map((timeGrain) => {
            const isSelected = metricTimeGrains.includes(timeGrain.id)
            const isDefault = timeGrain.is_default

            return (
              <motion.div
                key={timeGrain.id}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
              >
                <Button
                  variant={isSelected ? 'primary' : 'outline'}
                  onClick={() => handleTimeGrainToggle(timeGrain.id, isDefault)}
                  disabled={loading}
                  className="w-full flex flex-col items-center gap-1 h-auto py-3"
                >
                  <span className="font-medium">{timeGrain.display_name}</span>
                  <span className="text-xs opacity-70">{timeGrain.duration}</span>
                  {isSelected && (
                    <Badge variant="success" className="mt-1">Selected</Badge>
                  )}
                  {isDefault && !isSelected && (
                    <Badge variant="secondary" className="mt-1">Default</Badge>
                  )}
                </Button>
              </motion.div>
            )
          })}
        </div>
        
        {metricTimeGrains.length > 0 && (
          <div className="mt-4 p-3 bg-muted rounded-lg">
            <p className="text-sm font-medium mb-1">Selected Time Grains:</p>
            <div className="flex flex-wrap gap-2">
              {metricTimeGrains.map((tgId) => {
                const tg = timeGrains.find(t => t.id === tgId)
                return tg ? (
                  <Badge key={tgId} variant="primary">{tg.display_name}</Badge>
                ) : null
              })}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

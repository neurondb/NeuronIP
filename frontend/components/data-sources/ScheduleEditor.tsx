'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import Select from '@/components/ui/Select'

interface ScheduleEditorProps {
  schedule?: string
  onSave: (schedule: string) => void
  onCancel: () => void
}

export default function ScheduleEditor({ schedule, onSave, onCancel }: ScheduleEditorProps) {
  const [scheduleType, setScheduleType] = useState<'manual' | 'interval' | 'cron'>('interval')
  const [intervalValue, setIntervalValue] = useState('60')
  const [intervalUnit, setIntervalUnit] = useState<'minutes' | 'hours' | 'days'>('minutes')
  const [cronExpression, setCronExpression] = useState('0 * * * *')

  const handleSave = () => {
    let finalSchedule = ''
    switch (scheduleType) {
      case 'manual':
        finalSchedule = 'manual'
        break
      case 'interval':
        finalSchedule = `every_${intervalValue}_${intervalUnit}`
        break
      case 'cron':
        finalSchedule = cronExpression
        break
    }
    onSave(finalSchedule)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Sync Schedule</CardTitle>
        <CardDescription>Configure when this connector should sync data</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Schedule Type</label>
            <Select
              value={scheduleType}
              onChange={(e) => setScheduleType(e.target.value as any)}
              options={[
                { value: 'manual', label: 'Manual Only' },
                { value: 'interval', label: 'Interval' },
                { value: 'cron', label: 'Cron Expression' },
              ]}
            />
          </div>

          {scheduleType === 'interval' && (
            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="block text-sm font-medium mb-1">Interval</label>
                <Input
                  type="number"
                  value={intervalValue}
                  onChange={(e) => setIntervalValue(e.target.value)}
                  min="1"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Unit</label>
                <Select
                  value={intervalUnit}
                  onChange={(e) => setIntervalUnit(e.target.value as any)}
                  options={[
                    { value: 'minutes', label: 'Minutes' },
                    { value: 'hours', label: 'Hours' },
                    { value: 'days', label: 'Days' },
                  ]}
                />
              </div>
            </div>
          )}

          {scheduleType === 'cron' && (
            <div>
              <label className="block text-sm font-medium mb-1">Cron Expression</label>
              <Input
                value={cronExpression}
                onChange={(e) => setCronExpression(e.target.value)}
                placeholder="0 * * * *"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Format: minute hour day month weekday (e.g., "0 * * * *" = every hour)
              </p>
            </div>
          )}

          <div className="flex gap-2 pt-4">
            <Button onClick={handleSave}>Save Schedule</Button>
            <Button variant="outline" onClick={onCancel}>Cancel</Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

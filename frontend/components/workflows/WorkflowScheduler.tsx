'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import {
  useWorkflowSchedules,
  useScheduleWorkflow,
} from '@/lib/api/queries'
import { useQueryClient } from '@tanstack/react-query'
import { QUERY_KEYS } from '@/lib/utils/constants'
import { showToast } from '@/components/ui/Toast'

interface WorkflowSchedulerProps {
  workflowId: string
}

export default function WorkflowScheduler({ workflowId }: WorkflowSchedulerProps) {
  const [scheduleType, setScheduleType] = useState<'interval' | 'cron' | 'once'>('interval')
  const [intervalValue, setIntervalValue] = useState('')
  const [intervalUnit, setIntervalUnit] = useState<'minutes' | 'hours' | 'days'>('hours')
  const [cronExpression, setCronExpression] = useState('')
  const [scheduledTime, setScheduledTime] = useState('')
  const { data: schedules, isLoading } = useWorkflowSchedules(workflowId)
  const { mutate: scheduleWorkflow, isPending: isScheduling } = useScheduleWorkflow(workflowId)
  const [cancellingScheduleId, setCancellingScheduleId] = useState<string | null>(null)
  const queryClient = useQueryClient()

  const handleSchedule = () => {
    let scheduleConfig: Record<string, unknown> = {}

    if (scheduleType === 'interval') {
      if (!intervalValue) {
        showToast('Please enter an interval value', 'warning')
        return
      }
      scheduleConfig = {
        type: 'interval',
        interval: parseInt(intervalValue),
        unit: intervalUnit,
      }
    } else if (scheduleType === 'cron') {
      if (!cronExpression) {
        showToast('Please enter a cron expression', 'warning')
        return
      }
      scheduleConfig = {
        type: 'cron',
        cron_expression: cronExpression,
      }
    } else if (scheduleType === 'once') {
      if (!scheduledTime) {
        showToast('Please select a scheduled time', 'warning')
        return
      }
      scheduleConfig = {
        type: 'once',
        scheduled_time: scheduledTime,
      }
    }

    scheduleWorkflow(scheduleConfig, {
      onSuccess: () => {
        showToast('Workflow scheduled successfully', 'success')
        setIntervalValue('')
        setCronExpression('')
        setScheduledTime('')
      },
      onError: () => {
        showToast('Failed to schedule workflow', 'error')
      },
    })
  }

  const handleCancel = async (scheduleId: string) => {
    setCancellingScheduleId(scheduleId)
    try {
      const response = await fetch(`/api/v1/workflows/${workflowId}/schedules/${scheduleId}/cancel`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      })
      
      if (response.ok) {
        showToast('Schedule cancelled successfully', 'success')
        queryClient.invalidateQueries({ queryKey: QUERY_KEYS.workflowSchedules(workflowId) })
      } else {
        showToast('Failed to cancel schedule', 'error')
      }
    } catch (error) {
      showToast('Failed to cancel schedule', 'error')
    } finally {
      setCancellingScheduleId(null)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Workflow Scheduler</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Create Schedule */}
        <div className="mb-6 p-4 border border-border rounded-lg">
          <h3 className="font-semibold mb-3">Create Schedule</h3>
          <div className="space-y-3">
            {/* Schedule Type */}
            <div>
              <label className="text-sm font-medium mb-1 block">Schedule Type</label>
              <div className="flex gap-2">
                {(['interval', 'cron', 'once'] as const).map((type) => (
                  <button
                    key={type}
                    onClick={() => setScheduleType(type)}
                    className={`px-3 py-1 rounded text-sm ${
                      scheduleType === type
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </button>
                ))}
              </div>
            </div>

            {/* Interval Schedule */}
            {scheduleType === 'interval' && (
              <div className="flex gap-2">
                <div className="flex-1">
                  <label className="text-sm font-medium mb-1 block">Interval Value</label>
                  <input
                    type="number"
                    value={intervalValue}
                    onChange={(e) => setIntervalValue(e.target.value)}
                    placeholder="1"
                    className="w-full rounded border border-border px-3 py-2"
                  />
                </div>
                <div className="flex-1">
                  <label className="text-sm font-medium mb-1 block">Unit</label>
                  <select
                    value={intervalUnit}
                    onChange={(e) => setIntervalUnit(e.target.value as 'minutes' | 'hours' | 'days')}
                    className="w-full rounded border border-border px-3 py-2"
                  >
                    <option value="minutes">Minutes</option>
                    <option value="hours">Hours</option>
                    <option value="days">Days</option>
                  </select>
                </div>
              </div>
            )}

            {/* Cron Schedule */}
            {scheduleType === 'cron' && (
              <div>
                <label className="text-sm font-medium mb-1 block">Cron Expression</label>
                <input
                  type="text"
                  value={cronExpression}
                  onChange={(e) => setCronExpression(e.target.value)}
                  placeholder="0 0 * * *"
                  className="w-full rounded border border-border px-3 py-2 font-mono"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Format: minute hour day month weekday
                </p>
              </div>
            )}

            {/* Once Schedule */}
            {scheduleType === 'once' && (
              <div>
                <label className="text-sm font-medium mb-1 block">Scheduled Time</label>
                <input
                  type="datetime-local"
                  value={scheduledTime}
                  onChange={(e) => setScheduledTime(e.target.value)}
                  className="w-full rounded border border-border px-3 py-2"
                />
              </div>
            )}

            <Button onClick={handleSchedule} disabled={isScheduling}>
              {isScheduling ? 'Scheduling...' : 'Create Schedule'}
            </Button>
          </div>
        </div>

        {/* Active Schedules */}
        <div>
          <h3 className="font-semibold mb-3">Active Schedules</h3>
          {isLoading ? (
            <div className="text-center py-8 text-muted-foreground">Loading schedules...</div>
          ) : schedules?.schedules?.length > 0 ? (
            <div className="space-y-2">
              {schedules.schedules.map((schedule: any) => (
                <div
                  key={schedule.id}
                  className="p-3 border border-border rounded-lg flex items-center justify-between"
                >
                  <div>
                    <div className="font-semibold">
                      {schedule.type === 'interval'
                        ? `Every ${schedule.interval} ${schedule.unit}`
                        : schedule.type === 'cron'
                        ? schedule.cron_expression
                        : new Date(schedule.scheduled_time).toLocaleString()}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Next run: {schedule.next_run ? new Date(schedule.next_run).toLocaleString() : 'N/A'}
                    </div>
                  </div>
                  <Button
                    onClick={() => handleCancel(schedule.id)}
                    disabled={cancellingScheduleId === schedule.id}
                    size="sm"
                    variant="outline"
                  >
                    {cancellingScheduleId === schedule.id ? 'Cancelling...' : 'Cancel'}
                  </Button>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-sm text-muted-foreground">No active schedules</div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'

interface Integration {
  id: string
  name: string
  description: string
  status: 'connected' | 'disconnected'
  type: string
}

export default function IntegrationSettings() {
  const integrations: Integration[] = [
    {
      id: '1',
      name: 'Helpdesk',
      description: 'Sync support tickets from external helpdesk systems',
      status: 'connected',
      type: 'helpdesk',
    },
    {
      id: '2',
      name: 'Slack',
      description: 'Send notifications to Slack channels',
      status: 'disconnected',
      type: 'notification',
    },
    {
      id: '3',
      name: 'Email',
      description: 'Email notifications for alerts and updates',
      status: 'connected',
      type: 'notification',
    },
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>Integrations</CardTitle>
        <CardDescription>Manage external service integrations</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {integrations.map((integration) => (
          <div
            key={integration.id}
            className="flex items-center justify-between p-4 rounded-lg border border-border hover:bg-accent transition-colors"
          >
            <div className="flex items-center gap-4 flex-1">
              <div className={cn(
                'h-3 w-3 rounded-full',
                integration.status === 'connected' ? 'bg-green-500' : 'bg-gray-400'
              )} />
              <div className="flex-1">
                <div className="font-medium">{integration.name}</div>
                <div className="text-sm text-muted-foreground">{integration.description}</div>
              </div>
            </div>
            <Button variant="outline" size="sm">
              {integration.status === 'connected' ? 'Configure' : 'Connect'}
            </Button>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}

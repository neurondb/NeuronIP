'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'
import Tooltip from '@/components/ui/Tooltip'
import HelpText from '@/components/ui/HelpText'
import { InformationCircleIcon } from '@heroicons/react/24/outline'

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
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Integrations</CardTitle>
            <CardDescription>Manage external service integrations</CardDescription>
          </div>
          <HelpText
            variant="tooltip"
            content={
              <div>
                <p className="font-medium mb-1">Integrations</p>
                <p className="text-xs mb-2">
                  Connect external services to extend NeuronIP's capabilities. Integrations enable
                  notifications, data syncing, and workflow automation.
                </p>
                <p className="text-xs">
                  <a
                    href="/docs/integrations/custom-integrations"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline"
                  >
                    Learn more →
                  </a>
                </p>
              </div>
            }
          />
        </div>
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
                <div className="flex items-center gap-2">
                  <div className="font-medium">{integration.name}</div>
                  <Tooltip
                    content={`${integration.name} integration: ${integration.description}`}
                    variant="info"
                  >
                    <InformationCircleIcon className="h-4 w-4 text-muted-foreground cursor-help" />
                  </Tooltip>
                </div>
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

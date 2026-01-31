'use client'

import { useState } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'

export default function CRMAutomation() {
  const [crmType, setCrmType] = useState('salesforce')
  const [eventType, setEventType] = useState('contact_created')

  const createHook = async () => {
    try {
      await fetch('/api/v1/integrations/crm/hooks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          crm_type: crmType,
          event_type: eventType,
          trigger_config: {},
          action_config: {},
        }),
      })
      alert('CRM automation hook created')
    } catch (error) {
      console.error('Failed to create hook:', error)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>CRM Automation Hooks</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-2">CRM Type</label>
          <select
            value={crmType}
            onChange={(e) => setCrmType(e.target.value)}
            className="w-full p-2 border rounded"
          >
            <option value="salesforce">Salesforce</option>
            <option value="hubspot">HubSpot</option>
            <option value="dynamics">Dynamics</option>
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium mb-2">Event Type</label>
          <select
            value={eventType}
            onChange={(e) => setEventType(e.target.value)}
            className="w-full p-2 border rounded"
          >
            <option value="contact_created">Contact Created</option>
            <option value="deal_updated">Deal Updated</option>
            <option value="account_created">Account Created</option>
          </select>
        </div>
        <Button onClick={createHook}>Create Hook</Button>
      </CardContent>
    </Card>
  )
}

'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useAppStore } from '@/lib/store/useAppStore'

export default function SecuritySettings() {
  const { notificationsEnabled, setNotificationsEnabled } = useAppStore()

  return (
    <Card>
      <CardHeader>
        <CardTitle>Security & Authentication</CardTitle>
        <CardDescription>Manage security and authentication settings</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={notificationsEnabled}
              onChange={(e) => setNotificationsEnabled(e.target.checked)}
              className="rounded border-border"
            />
            <span className="text-sm">Enable Two-Factor Authentication</span>
          </label>
        </div>

        <div>
          <label className="text-sm font-medium mb-3 block">Password Policy</label>
          <select className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
            <option>Standard (8+ characters)</option>
            <option>Strong (12+ characters, mixed case)</option>
            <option>Very Strong (16+ characters, special chars)</option>
          </select>
        </div>

        <div>
          <label className="text-sm font-medium mb-3 block">Session Timeout</label>
          <select className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
            <option>15 minutes</option>
            <option>30 minutes</option>
            <option>1 hour</option>
            <option>8 hours</option>
          </select>
        </div>

        <Button variant="outline">Change Password</Button>
      </CardContent>
    </Card>
  )
}

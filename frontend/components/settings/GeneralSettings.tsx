'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useAppStore } from '@/lib/store/useAppStore'

export default function GeneralSettings() {
  const { theme, setTheme } = useAppStore()

  return (
    <Card>
      <CardHeader>
        <CardTitle>General Settings</CardTitle>
        <CardDescription>Application-wide settings and preferences</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div>
          <label className="text-sm font-medium mb-3 block">Theme</label>
          <div className="flex gap-2">
            <Button
              variant={theme === 'light' ? 'primary' : 'outline'}
              size="sm"
              onClick={() => setTheme('light')}
            >
              Light
            </Button>
            <Button
              variant={theme === 'dark' ? 'primary' : 'outline'}
              size="sm"
              onClick={() => setTheme('dark')}
            >
              Dark
            </Button>
            <Button
              variant={theme === 'system' ? 'primary' : 'outline'}
              size="sm"
              onClick={() => setTheme('system')}
            >
              System
            </Button>
          </div>
        </div>

        <div>
          <label className="text-sm font-medium mb-3 block">Language</label>
          <select className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
            <option>English</option>
            <option>Spanish</option>
            <option>French</option>
            <option>German</option>
          </select>
        </div>

        <div>
          <label className="text-sm font-medium mb-3 block">Date Format</label>
          <select className="w-full rounded-lg border border-border bg-background px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
            <option>MM/DD/YYYY</option>
            <option>DD/MM/YYYY</option>
            <option>YYYY-MM-DD</option>
          </select>
        </div>
      </CardContent>
    </Card>
  )
}

'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useAppStore } from '@/lib/store/useAppStore'

export default function PreferencesPanel() {
  const { theme, setTheme, notificationsEnabled, setNotificationsEnabled } = useAppStore()

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
          <CardDescription>Customize the look and feel of the application</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-2 block">Theme</label>
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
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>Manage notification preferences</CardDescription>
        </CardHeader>
        <CardContent>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={notificationsEnabled}
              onChange={(e) => setNotificationsEnabled(e.target.checked)}
              className="rounded border-border"
            />
            <span className="text-sm">Enable notifications</span>
          </label>
        </CardContent>
      </Card>
    </div>
  )
}

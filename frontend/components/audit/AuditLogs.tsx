'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import { useAuditEvents, useAuditActivity, useSearchAuditEvents } from '@/lib/api/queries'
import { MagnifyingGlassIcon, EyeIcon, ClockIcon } from '@heroicons/react/24/outline'
import { format } from 'date-fns'

export default function AuditLogs() {
  const [view, setView] = useState<'events' | 'activity' | 'search'>('events')
  const [searchQuery, setSearchQuery] = useState('')
  const [filters, setFilters] = useState<Record<string, unknown>>({})

  const { data: eventsData, isLoading: eventsLoading } = useAuditEvents(filters)
  const { data: activityData, isLoading: activityLoading } = useAuditActivity(filters)
  const searchMutation = useSearchAuditEvents()

  const events = eventsData?.events || []
  const activity = activityData?.timeline || []
  const searchResults = searchMutation.data?.events || []

  const displayData = view === 'search' ? searchResults : view === 'activity' ? activity : events
  const isLoading = view === 'search' ? searchMutation.isPending : view === 'activity' ? activityLoading : eventsLoading

  const handleSearch = async () => {
    if (!searchQuery.trim()) return
    await searchMutation.mutateAsync({ query: searchQuery, limit: 100 })
  }

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Audit Logs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2 mb-4">
            <Button
              variant={view === 'events' ? 'primary' : 'outline'}
              onClick={() => setView('events')}
              size="sm"
            >
              Events
            </Button>
            <Button
              variant={view === 'activity' ? 'primary' : 'outline'}
              onClick={() => setView('activity')}
              size="sm"
            >
              Activity Timeline
            </Button>
            <Button
              variant={view === 'search' ? 'primary' : 'outline'}
              onClick={() => setView('search')}
              size="sm"
            >
              Search
            </Button>
          </div>

          {view === 'search' && (
            <div className="flex gap-2 mb-4">
              <div className="flex-1 relative">
                <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                <input
                  type="text"
                  placeholder="Search audit logs..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                  className="w-full pl-10 pr-4 py-2 rounded-lg border border-border bg-background focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </div>
              <Button onClick={handleSearch} disabled={!searchQuery.trim() || searchMutation.isPending}>
                Search
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>
            {view === 'events' ? 'Audit Events' : view === 'activity' ? 'Activity Timeline' : 'Search Results'} (
            {displayData.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-center text-muted-foreground py-8">Loading...</p>
          ) : displayData.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">No logs found</p>
          ) : (
            <div className="space-y-3 max-h-[600px] overflow-y-auto">
              {displayData.map((event: any, index: number) => (
                <div
                  key={event.id || index}
                  className="p-4 border border-border rounded-lg hover:bg-muted/20 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <EyeIcon className="h-4 w-4 text-muted-foreground" />
                        <span className="font-semibold text-sm">{event.action_type || event.action}</span>
                        {event.resource_type && (
                          <span className="text-xs px-2 py-0.5 bg-muted rounded">{event.resource_type}</span>
                        )}
                      </div>
                      {event.user_id && (
                        <p className="text-sm text-muted-foreground">User: {event.user_id}</p>
                      )}
                      {event.resource_id && (
                        <p className="text-sm text-muted-foreground">Resource: {event.resource_id}</p>
                      )}
                      {event.details && (
                        <p className="text-xs text-muted-foreground mt-2">{JSON.stringify(event.details)}</p>
                      )}
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <ClockIcon className="h-4 w-4" />
                      {event.timestamp || event.created_at
                        ? format(new Date(event.timestamp || event.created_at), 'PPpp')
                        : 'N/A'}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
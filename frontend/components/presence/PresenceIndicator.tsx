'use client'

import { usePresence } from '@/lib/websocket/hooks'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/Avatar'
import { cn } from '@/lib/utils/cn'

interface PresenceIndicatorProps {
  roomId: string
  className?: string
  maxVisible?: number
}

export function PresenceIndicator({ roomId, className, maxVisible = 3 }: PresenceIndicatorProps) {
  const users = usePresence(roomId)

  if (users.length === 0) {
    return null
  }

  const visibleUsers = users.slice(0, maxVisible)
  const remainingCount = users.length - maxVisible

  return (
    <div className={cn('flex items-center -space-x-2', className)}>
      {visibleUsers.map((user) => (
        <Avatar key={user.id} className="border-2 border-background">
          <AvatarImage src={user.avatar} alt={user.name} />
          <AvatarFallback>{user.name.charAt(0).toUpperCase()}</AvatarFallback>
        </Avatar>
      ))}
      {remainingCount > 0 && (
        <div className="flex h-10 w-10 items-center justify-center rounded-full border-2 border-background bg-muted text-xs font-medium">
          +{remainingCount}
        </div>
      )}
    </div>
  )
}

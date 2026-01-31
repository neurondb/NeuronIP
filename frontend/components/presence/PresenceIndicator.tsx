'use client'

import { motion } from 'framer-motion'

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/Avatar'
import { cn } from '@/lib/utils/cn'
import { usePresence } from '@/lib/websocket/hooks'


interface PresenceIndicatorProps {
  roomId: string
  className?: string
  maxVisible?: number
  showCount?: boolean
  variant?: 'compact' | 'expanded'
}

export function PresenceIndicator({
  roomId,
  className,
  maxVisible = 3,
  showCount = true,
  variant = 'compact',
}: PresenceIndicatorProps) {
  const users = usePresence(roomId)

  if (users.length === 0) {
    return (
      <div className={cn('text-xs text-muted-foreground', className)}>
        {variant === 'expanded' && 'No one else is viewing'}
      </div>
    )
  }

  const visibleUsers = users.slice(0, maxVisible)
  const remainingCount = users.length - maxVisible

  if (variant === 'expanded') {
    return (
      <div className={cn('flex items-center gap-2', className)}>
        <div className="flex items-center -space-x-2">
          {visibleUsers.map((user, index) => (
            <motion.div
              key={user.id}
              initial={{ opacity: 0, scale: 0 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: index * 0.1 }}
            >
              <Avatar className="border-2 border-background">
                <AvatarImage src={user.avatar} alt={user.name} />
                <AvatarFallback>{user.name.charAt(0).toUpperCase()}</AvatarFallback>
              </Avatar>
            </motion.div>
          ))}
          {remainingCount > 0 && (
            <div className="flex h-10 w-10 items-center justify-center rounded-full border-2 border-background bg-muted text-xs font-medium">
              +{remainingCount}
            </div>
          )}
        </div>
        {showCount && (
          <span className="text-sm text-muted-foreground">
            {users.length} {users.length === 1 ? 'user' : 'users'} viewing
          </span>
        )}
      </div>
    )
  }

  return (
    <div className={cn('flex items-center -space-x-2', className)}>
      {visibleUsers.map((user, index) => (
        <motion.div
          key={user.id}
          initial={{ opacity: 0, scale: 0 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: index * 0.1 }}
        >
          <Avatar className="border-2 border-background">
            <AvatarImage src={user.avatar} alt={user.name} />
            <AvatarFallback>{user.name.charAt(0).toUpperCase()}</AvatarFallback>
          </Avatar>
        </motion.div>
      ))}
      {remainingCount > 0 && (
        <div className="flex h-10 w-10 items-center justify-center rounded-full border-2 border-background bg-muted text-xs font-medium">
          +{remainingCount}
        </div>
      )}
    </div>
  )
}

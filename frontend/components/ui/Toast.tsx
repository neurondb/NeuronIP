'use client'

import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { XMarkIcon, CheckCircleIcon, ExclamationCircleIcon, InformationCircleIcon } from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'

export type ToastType = 'success' | 'error' | 'info' | 'warning'

interface Toast {
  id: string
  message: string
  type: ToastType
  duration?: number
}

interface ToastContextType {
  toasts: Toast[]
  showToast: (message: string, type?: ToastType, duration?: number) => void
  removeToast: (id: string) => void
}

// Global toast state (simple implementation, can be enhanced with Context API)
let toastListeners: ((toasts: Toast[]) => void)[] = []
let toastState: Toast[] = []

function notifyListeners() {
  toastListeners.forEach((listener) => listener([...toastState]))
}

export const toastManager = {
  subscribe: (listener: (toasts: Toast[]) => void) => {
    toastListeners.push(listener)
    return () => {
      toastListeners = toastListeners.filter((l) => l !== listener)
    }
  },
  show: (message: string, type: ToastType = 'info', duration = 3000) => {
    const id = Math.random().toString(36).substr(2, 9)
    const newToast: Toast = { id, message, type, duration }
    toastState.push(newToast)
    notifyListeners()

    if (duration > 0) {
      setTimeout(() => {
        toastManager.remove(id)
      }, duration)
    }
  },
  remove: (id: string) => {
    toastState = toastState.filter((t) => t.id !== id)
    notifyListeners()
  },
}

// Toast Component
export default function Toast({ toast }: { toast: Toast }) {
  const icons = {
    success: CheckCircleIcon,
    error: ExclamationCircleIcon,
    info: InformationCircleIcon,
    warning: ExclamationCircleIcon,
  }

  const colors = {
    success: 'bg-green-50 text-green-800 border-green-200 dark:bg-green-900/20 dark:text-green-400',
    error: 'bg-red-50 text-red-800 border-red-200 dark:bg-red-900/20 dark:text-red-400',
    info: 'bg-blue-50 text-blue-800 border-blue-200 dark:bg-blue-900/20 dark:text-blue-400',
    warning: 'bg-yellow-50 text-yellow-800 border-yellow-200 dark:bg-yellow-900/20 dark:text-yellow-400',
  }

  const Icon = icons[toast.type]

  return (
    <motion.div
      initial={{ opacity: 0, y: -20, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: -20, scale: 0.95 }}
      className={cn(
        'flex items-center gap-3 rounded-lg border p-4 shadow-lg',
        colors[toast.type]
      )}
    >
      <Icon className="h-5 w-5 flex-shrink-0" />
      <p className="flex-1 text-sm font-medium">{toast.message}</p>
      <button
        onClick={() => toastManager.remove(toast.id)}
        className="flex-shrink-0 rounded opacity-70 hover:opacity-100"
      >
        <XMarkIcon className="h-4 w-4" />
      </button>
    </motion.div>
  )
}

// Toast Container
export function ToastContainer() {
  const [toasts, setToasts] = useState<Toast[]>([])

  useEffect(() => {
    const unsubscribe = toastManager.subscribe(setToasts)
    return unsubscribe
  }, [])

  return (
    <div className="fixed top-4 right-4 z-[100] flex flex-col gap-2">
      <AnimatePresence>
        {toasts.map((toast) => (
          <Toast key={toast.id} toast={toast} />
        ))}
      </AnimatePresence>
    </div>
  )
}

// Convenience function
export function showToast(message: string, type?: ToastType, duration?: number) {
  toastManager.show(message, type, duration)
}

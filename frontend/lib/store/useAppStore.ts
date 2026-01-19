import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AppState {
  // UI State
  sidebarCollapsed: boolean
  theme: 'light' | 'dark' | 'system'
  
  // User preferences
  notificationsEnabled: boolean
  
  // Actions
  setSidebarCollapsed: (collapsed: boolean) => void
  setTheme: (theme: 'light' | 'dark' | 'system') => void
  setNotificationsEnabled: (enabled: boolean) => void
  toggleSidebar: () => void
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      // Initial state
      sidebarCollapsed: false,
      theme: 'system',
      notificationsEnabled: true,

      // Actions
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
      setTheme: (theme) => set({ theme }),
      setNotificationsEnabled: (enabled) => set({ notificationsEnabled: enabled }),
      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
    }),
    {
      name: 'neuronip-app-store',
      partialize: (state) => ({
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
        notificationsEnabled: state.notificationsEnabled,
      }),
    }
  )
)

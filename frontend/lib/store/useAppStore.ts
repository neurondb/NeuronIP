import { create } from 'zustand'
import { persist } from 'zustand/middleware'

const MAX_RECENT_PATHS = 8

interface AppState {
  // UI State
  sidebarCollapsed: boolean
  theme: 'light' | 'dark' | 'system'
  expandedNavGroups: Record<string, boolean>

  // Notion-like navigation
  recentPaths: string[]
  pinnedPaths: string[]

  // User preferences
  notificationsEnabled: boolean

  // Actions
  setSidebarCollapsed: (collapsed: boolean) => void
  setTheme: (theme: 'light' | 'dark' | 'system') => void
  setNotificationsEnabled: (enabled: boolean) => void
  toggleSidebar: () => void
  toggleNavGroup: (group: string) => void
  addRecentPath: (path: string) => void
  togglePinnedPath: (path: string) => void
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      sidebarCollapsed: false,
      theme: 'system',
      notificationsEnabled: true,
      expandedNavGroups: {
        data: true,
        ai: true,
        ops: false,
        admin: false,
      },
      recentPaths: [],
      pinnedPaths: [],

      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),
      setTheme: (theme) => set({ theme }),
      setNotificationsEnabled: (enabled) => set({ notificationsEnabled: enabled }),
      toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
      toggleNavGroup: (group) =>
        set((state) => ({
          expandedNavGroups: {
            ...state.expandedNavGroups,
            [group]: !state.expandedNavGroups[group],
          },
        })),
      addRecentPath: (path) =>
        set((state) => {
          const normalized = path.replace(/\/$/, '') || '/'
          const filtered = state.recentPaths.filter((p) => p !== normalized)
          const next = [normalized, ...filtered].slice(0, MAX_RECENT_PATHS)
          return { recentPaths: next }
        }),
      togglePinnedPath: (path) =>
        set((state) => {
          const normalized = path.replace(/\/$/, '') || '/'
          const has = state.pinnedPaths.includes(normalized)
          const next = has
            ? state.pinnedPaths.filter((p) => p !== normalized)
            : [...state.pinnedPaths, normalized]
          return { pinnedPaths: next }
        }),
    }),
    {
      name: 'neuronip-app-store',
      storage:
        typeof window !== 'undefined'
          ? {
              getItem: (name) => {
                const str = localStorage.getItem(name)
                return str ? JSON.parse(str) : null
              },
              setItem: (name, value) => {
                localStorage.setItem(name, JSON.stringify(value))
              },
              removeItem: (name) => {
                localStorage.removeItem(name)
              },
            }
          : undefined,
      // @ts-expect-error - partialize returns a subset of AppState
      partialize: (state) => ({
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
        notificationsEnabled: state.notificationsEnabled,
        expandedNavGroups: state.expandedNavGroups,
        recentPaths: state.recentPaths,
        pinnedPaths: state.pinnedPaths,
      }),
    }
  )
)

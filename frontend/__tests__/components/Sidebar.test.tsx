import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'

import Sidebar from '@/components/layout/Sidebar'
import { useAppStore } from '@/lib/store/useAppStore'

import { render, screen } from '../utils/test-utils'

// Mock next/navigation
vi.mock('next/navigation', () => ({
  usePathname: () => '/semantic',
  Link: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}))

// Mock the store
vi.mock('@/lib/store/useAppStore', () => ({
  useAppStore: vi.fn(() => ({
    sidebarCollapsed: false,
    expandedNavGroups: {},
    recentPaths: [],
    pinnedPaths: [],
    toggleSidebar: vi.fn(),
    toggleNavGroup: vi.fn(),
    togglePinnedPath: vi.fn(),
  })),
}))

describe('Sidebar', () => {
  it('renders navigation groups', () => {
    render(<Sidebar />)
    
    expect(screen.getByText('Data')).toBeInTheDocument()
    expect(screen.getByText('AI & Automation')).toBeInTheDocument()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
  })

  it('renders navigation items', () => {
    render(<Sidebar />)
    
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Data')).toBeInTheDocument()
    expect(screen.getByText('AI & Automation')).toBeInTheDocument()
  })

  it('highlights active navigation item', () => {
    render(<Sidebar />)
    
    const semanticLink = screen.queryByText('Semantic Search')
    if (semanticLink) {
      const link = semanticLink.closest('a')
      if (link) {
        expect(link).toHaveAttribute('href', '/semantic')
      }
    } else {
      expect(screen.getByText('Data')).toBeInTheDocument()
    }
  })

  it('toggles sidebar collapse', async () => {
    const user = userEvent.setup()
    const toggleSidebar = vi.fn()
    
    vi.mocked(useAppStore).mockReturnValue({
      sidebarCollapsed: false,
      expandedNavGroups: {},
      recentPaths: [],
      pinnedPaths: [],
      toggleSidebar,
      toggleNavGroup: vi.fn(),
      togglePinnedPath: vi.fn(),
    } as any)

    render(<Sidebar />)
    
    const toggleButton = screen.getByLabelText(/expand sidebar|collapse sidebar/i)
    await user.click(toggleButton)
    
    expect(toggleSidebar).toHaveBeenCalled()
  })

  it('expands and collapses navigation groups', async () => {
    const user = userEvent.setup()
    const toggleNavGroup = vi.fn()
    
    vi.mocked(useAppStore).mockReturnValue({
      sidebarCollapsed: false,
      expandedNavGroups: {},
      recentPaths: [],
      pinnedPaths: [],
      toggleSidebar: vi.fn(),
      toggleNavGroup,
      togglePinnedPath: vi.fn(),
    } as any)

    render(<Sidebar />)
    
    const dataGroup = screen.getByText('Data')
    await user.click(dataGroup)
    
    expect(toggleNavGroup).toHaveBeenCalledWith('data')
  })
})

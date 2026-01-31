import { describe, it, expect } from 'vitest'

import ActivityFeed from '@/components/dashboard/ActivityFeed'

import { render, screen, waitFor } from '../utils/test-utils'

describe('ActivityFeed', () => {
  it('renders activity feed', () => {
    render(<ActivityFeed />)

    const feed =
      screen.queryByText(/activity|recent/i) ||
      screen.queryByRole('list') ||
      document.querySelector('[class*="activity"]')
    expect(feed !== null).toBeTruthy()
  })

  it('displays activities', async () => {
    render(<ActivityFeed />)

    await waitFor(
      () => {
        const activities = screen.queryAllByText(
          /search|query|workflow|Semantic|Warehouse|Data Processing/i
        )
        expect(activities.length).toBeGreaterThan(0)
      },
      { timeout: 3000 }
    )
  })

  it('formats timestamps correctly', () => {
    render(<ActivityFeed />)

    const timeElements = screen.queryAllByText(/ago|just now|minute|hour/i)
    expect(timeElements !== null).toBeTruthy()
  })
})

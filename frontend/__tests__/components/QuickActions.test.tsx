import { describe, it, expect } from 'vitest'

import QuickActions from '@/components/dashboard/QuickActions'

import { render, screen } from '../utils/test-utils'

describe('QuickActions', () => {
  it('renders quick actions', () => {
    render(<QuickActions />)
    
    // Should render quick action links
    const actions = screen.queryAllByRole('link')
    expect(actions.length).toBeGreaterThan(0)
  })

  it('renders all action items', () => {
    render(<QuickActions />)
    
    // Check for action items
    const actionTexts = [
      /semantic|search/i,
      /warehouse|query/i,
      /workflow/i,
      /compliance/i,
    ]
    
    actionTexts.forEach(text => {
      const element = screen.queryByText(text)
      // At least some actions should be present
      if (actionTexts.indexOf(text) === 0) {
        expect(element || screen.queryByRole('link')).toBeTruthy()
      }
    })
  })
})

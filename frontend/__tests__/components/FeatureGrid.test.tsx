import { describe, it, expect } from 'vitest'

import FeatureGrid from '@/components/features/FeatureGrid'

import { render, screen } from '../utils/test-utils'

describe('FeatureGrid', () => {
  it('renders feature grid', () => {
    render(<FeatureGrid />)
    
    // Should render at least some features - use getAllByText since there might be multiple
    const features = screen.getAllByText(/Dashboard|Semantic Search|Warehouse/i)
    expect(features.length).toBeGreaterThan(0)
  })

  it('renders all feature cards', () => {
    render(<FeatureGrid />)
    
    // Check for multiple features - use getAllByText to handle duplicates
    const dashboardElements = screen.getAllByText(/Dashboard/i)
    expect(dashboardElements.length).toBeGreaterThan(0)
    
    const semanticElements = screen.getAllByText(/Semantic Search/i)
    expect(semanticElements.length).toBeGreaterThan(0)
  })

  it('renders feature links', () => {
    const { container } = render(<FeatureGrid />)
    
    // Features should be clickable links
    const links = container.querySelectorAll('a[href^="/"]')
    expect(links.length).toBeGreaterThan(0)
  })
})

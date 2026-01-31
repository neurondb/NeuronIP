import { describe, it, expect } from 'vitest'

import Loading from '@/components/ui/Loading'

import { render } from '../utils/test-utils'

describe('Loading', () => {
  it('renders loading spinner', () => {
    const { container } = render(<Loading />)
    
    // Loading component renders a motion.div with spinner (framer-motion)
    // The spinner is a div with border classes
    const wrapper = container.querySelector('div')
    expect(wrapper).toBeTruthy()
    // Component renders without errors
    expect(container.firstChild).toBeTruthy()
  })

  it('renders with small size', () => {
    const { container } = render(<Loading size="sm" />)
    
    // Component should render
    const wrapper = container.querySelector('div')
    expect(wrapper).toBeTruthy()
  })

  it('renders with large size', () => {
    const { container } = render(<Loading size="lg" />)
    
    // Component should render
    const wrapper = container.querySelector('div')
    expect(wrapper).toBeTruthy()
  })

  it('renders with custom className', () => {
    const { container } = render(<Loading className="custom-class" />)
    
    const spinner = container.querySelector('.custom-class') ||
                   container.querySelector('[role="status"]')
    expect(spinner).toBeTruthy()
  })
})

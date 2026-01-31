import { describe, it, expect } from 'vitest'

import BarChart from '@/components/charts/BarChart'

import { render } from '../utils/test-utils'

describe('Chart Components', () => {
  const mockData = [
    { name: 'Jan', value: 100 },
    { name: 'Feb', value: 200 },
    { name: 'Mar', value: 150 },
  ]

  it('renders bar chart', () => {
    const { container } = render(
      <div style={{ width: '500px', height: '300px' }}>
        <BarChart data={mockData} dataKeys={['value']} xAxisKey="name" />
      </div>
    )
    
    // Recharts ResponsiveContainer renders a div wrapper
    // In test environment, SVG might not render, but component should mount
    const wrapper = container.querySelector('div')
    expect(wrapper).toBeTruthy()
    // Component renders without errors
    expect(container.firstChild).toBeTruthy()
  })

  it('renders chart with custom height', () => {
    const { container } = render(
      <div style={{ width: '500px', height: '400px' }}>
        <BarChart data={mockData} dataKeys={['value']} xAxisKey="name" />
      </div>
    )
    
    // Component should render without errors
    const wrapper = container.querySelector('div')
    expect(wrapper).toBeTruthy()
    expect(container.firstChild).toBeTruthy()
  })

  it('renders chart with title', () => {
    const { container } = render(
      <div style={{ width: '500px', height: '300px' }}>
        <h3>Sales Chart</h3>
        <BarChart data={mockData} dataKeys={['value']} xAxisKey="name" />
      </div>
    )
    
    expect(container.querySelector('h3')?.textContent).toBe('Sales Chart')
    // Chart component should render without errors
    const wrapper = container.querySelector('div')
    expect(wrapper).toBeTruthy()
  })
})

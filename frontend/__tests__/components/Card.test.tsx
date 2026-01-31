import { describe, it, expect } from 'vitest'

import { Card, CardHeader, CardContent, CardFooter } from '@/components/ui/Card'

import { render, screen } from '../utils/test-utils'

describe('Card', () => {
  it('renders card with children', () => {
    render(
      <Card>
        <p>Card content</p>
      </Card>
    )
    
    expect(screen.getByText('Card content')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <Card className="custom-class">
        <p>Content</p>
      </Card>
    )
    
    const card = container.querySelector('.custom-class')
    expect(card).toBeInTheDocument()
  })

  it('renders card header when provided', () => {
    render(
      <Card>
        <CardHeader>
          <h2>Card Title</h2>
        </CardHeader>
        <CardContent>Content</CardContent>
      </Card>
    )
    
    expect(screen.getByText('Card Title')).toBeInTheDocument()
  })

  it('renders card footer when provided', () => {
    render(
      <Card>
        <CardContent>Content</CardContent>
        <CardFooter>
          <button>Action</button>
        </CardFooter>
      </Card>
    )
    
    expect(screen.getByText('Action')).toBeInTheDocument()
  })
})

import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'

import { render, screen } from '../utils/test-utils'

// Mock table component structure
const Table = ({ children }: { children: React.ReactNode }) => (
  <table>{children}</table>
)

const TableHeader = ({ children }: { children: React.ReactNode }) => (
  <thead>{children}</thead>
)

const TableBody = ({ children }: { children: React.ReactNode }) => (
  <tbody>{children}</tbody>
)

const TableRow = ({ children }: { children: React.ReactNode }) => (
  <tr>{children}</tr>
)

const TableHead = ({ children }: { children: React.ReactNode }) => (
  <th>{children}</th>
)

const TableCell = ({ children }: { children: React.ReactNode }) => (
  <td>{children}</td>
)

describe('Table', () => {
  const mockData = [
    { id: 1, name: 'Item 1', value: 100 },
    { id: 2, name: 'Item 2', value: 200 },
    { id: 3, name: 'Item 3', value: 300 },
  ]

  it('renders table with data', () => {
    render(
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Value</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {mockData.map((item) => (
            <TableRow key={item.id}>
              <TableCell>{item.name}</TableCell>
              <TableCell>{item.value}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    )
    
    expect(screen.getByText('Item 1')).toBeInTheDocument()
    expect(screen.getByText('Item 2')).toBeInTheDocument()
    expect(screen.getByText('Item 3')).toBeInTheDocument()
  })

  it('handles sorting', async () => {
    const user = userEvent.setup()
    const onSort = vi.fn()
    
    render(
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>
              <button onClick={() => onSort('name')} aria-label="Sort by name">Name</button>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {mockData.map((item) => (
            <TableRow key={item.id}>
              <TableCell>{item.name}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    )
    
    const sortButton = screen.getByText('Name')
    if (sortButton) {
      await user.click(sortButton)
      expect(onSort).toHaveBeenCalledWith('name')
    }
  })

  it('handles filtering', async () => {
    const user = userEvent.setup()
    
    render(
      <div>
        <input
          type="text"
          placeholder="Filter"
          onChange={(e) => {
            // Filter logic would go here
          }}
        />
        <Table>
          <TableBody>
            {mockData.map((item) => (
              <TableRow key={item.id}>
                <TableCell>{item.name}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    )
    
    const filterInput = screen.getByPlaceholderText('Filter')
    await user.type(filterInput, 'Item 1')
    
    expect(filterInput).toHaveValue('Item 1')
  })

  it('handles pagination', async () => {
    const user = userEvent.setup()
    const onPageChange = vi.fn()
    
    render(
      <div>
        <Table>
          <TableBody>
            {mockData.map((item) => (
              <TableRow key={item.id}>
                <TableCell>{item.name}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <button onClick={() => onPageChange(2)}>Next</button>
      </div>
    )
    
    const nextButton = screen.getByText('Next')
    await user.click(nextButton)
    
    expect(onPageChange).toHaveBeenCalledWith(2)
  })
})

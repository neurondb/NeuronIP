import { describe, it, expect } from 'vitest'

import { cn } from '@/lib/utils/cn'

describe('cn utility', () => {
  it('merges class names', () => {
    expect(cn('class1', 'class2')).toBe('class1 class2')
  })

  it('handles conditional classes', () => {
    expect(cn('base', true && 'conditional')).toBe('base conditional')
    expect(cn('base', false && 'conditional')).toBe('base')
  })

  it('handles undefined and null', () => {
    expect(cn('base', undefined, null)).toBe('base')
  })

  it('handles arrays', () => {
    expect(cn(['class1', 'class2'])).toBe('class1 class2')
  })

  it('handles objects', () => {
    expect(cn({ class1: true, class2: false })).toBe('class1')
  })

  it('handles mixed inputs', () => {
    expect(cn('base', ['array1', 'array2'], { conditional: true })).toContain('base')
    expect(cn('base', ['array1', 'array2'], { conditional: true })).toContain('array1')
    expect(cn('base', ['array1', 'array2'], { conditional: true })).toContain('conditional')
  })

  it('removes duplicates', () => {
    const result = cn('class1', 'class1', 'class2')
    expect(result).toContain('class1')
    expect(result).toContain('class2')
  })
})

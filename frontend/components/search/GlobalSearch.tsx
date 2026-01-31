'use client'

import { MagnifyingGlassIcon, ClockIcon, XMarkIcon, DocumentTextIcon } from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useRouter } from 'next/navigation'
import { useState, useEffect, useRef, useMemo } from 'react'

import { ROUTE_ARCHETYPE_MAP } from '@/lib/pageArchetypes'
import { slideUp, transition } from '@/lib/animations/variants'
import { useDebounce } from '@/lib/hooks/useDebounce'
import { useSearchHistory } from '@/lib/hooks/useSearchHistory'
import { cn } from '@/lib/utils/cn'

interface SearchSuggestion {
  id: string
  text: string
  type: 'history' | 'suggestion' | 'module' | 'page'
  module?: string
  href?: string
}

interface GlobalSearchProps {
  onSelect?: (query: string) => void
  placeholder?: string
  className?: string
}

export default function GlobalSearch({
  onSelect,
  placeholder = 'Search...',
  className,
}: GlobalSearchProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [suggestions, setSuggestions] = useState<SearchSuggestion[]>([])
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const inputRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const debouncedQuery = useDebounce(query, 300)
  const { history, addToHistory } = useSearchHistory()
  const router = useRouter()

  const pageSuggestions = useMemo(
    () =>
      Object.values(ROUTE_ARCHETYPE_MAP).map((meta) => ({
        id: `page-${meta.path}`,
        text: meta.title,
        type: 'page' as const,
        href: meta.path,
      })),
    []
  )

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  useEffect(() => {
    const q = debouncedQuery.trim().toLowerCase()
    if (!q) {
      const historyItems = history.slice(0, 4).map((item) => ({
        id: item.id,
        text: item.query,
        type: 'history' as const,
        module: item.module,
      }))
      const pages = pageSuggestions.slice(0, 6)
      setSuggestions([...historyItems, ...pages])
      return
    }

    const filteredHistory = history
      .filter((item) => item.query.toLowerCase().includes(q))
      .slice(0, 3)
      .map((item) => ({
        id: item.id,
        text: item.query,
        type: 'history' as const,
        module: item.module,
      }))

    const filteredPages = pageSuggestions.filter((p) => p.text.toLowerCase().includes(q)).slice(0, 5)
    const mockModule: SearchSuggestion[] = [
      { id: 'm1', text: `${debouncedQuery} in Semantic Search`, type: 'module', module: 'semantic' },
      { id: 'm2', text: `${debouncedQuery} in Warehouse`, type: 'module', module: 'warehouse' },
      { id: 'm3', text: `${debouncedQuery} in Workflows`, type: 'module', module: 'workflows' },
    ]
    setSuggestions([...filteredHistory, ...filteredPages, ...mockModule])
  }, [debouncedQuery, history, pageSuggestions])

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelectedIndex((prev) => Math.min(prev + 1, suggestions.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelectedIndex((prev) => Math.max(prev - 1, -1))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (selectedIndex >= 0 && suggestions[selectedIndex]) {
        const s = suggestions[selectedIndex]
        if (s.type === 'page' && s.href) {
          setIsOpen(false)
          router.push(s.href)
          setQuery('')
          inputRef.current?.blur()
        } else {
          handleSelect(s.text)
        }
      } else if (query.trim()) {
        handleSelect(query)
      }
    } else if (e.key === 'Escape') {
      setIsOpen(false)
      inputRef.current?.blur()
    }
  }

  const handleSelect = (selectedQuery: string) => {
    setQuery(selectedQuery)
    setIsOpen(false)
    addToHistory(selectedQuery)
    onSelect?.(selectedQuery)
    inputRef.current?.blur()
  }

  const handleChange = (value: string) => {
    setQuery(value)
    setIsOpen(true)
    setSelectedIndex(-1)
  }

  return (
    <div ref={containerRef} className={cn('relative w-full max-w-md', className)}>
      <div className="relative">
        <MagnifyingGlassIcon className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-muted-foreground" />
        <input
          ref={inputRef}
          type="text"
          placeholder={placeholder}
          value={query}
          onChange={(e) => handleChange(e.target.value)}
          onFocus={() => setIsOpen(true)}
          onKeyDown={handleKeyDown}
          className={cn(
            'w-full rounded-lg border border-border bg-background py-2 pl-10 pr-4',
            'text-sm placeholder:text-muted-foreground',
            'focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2'
          )}
        />
        {query && (
          <button
            onClick={() => {
              setQuery('')
              setIsOpen(false)
              inputRef.current?.focus()
            }}
            className="absolute right-3 top-1/2 -translate-y-1/2 rounded p-1 hover:bg-accent"
          >
            <XMarkIcon className="h-4 w-4" />
          </button>
        )}
      </div>

      <AnimatePresence>
        {isOpen && suggestions.length > 0 && (
          <motion.div
            variants={slideUp}
            initial="hidden"
            animate="visible"
            exit="hidden"
            transition={transition}
            className="absolute top-full z-50 mt-2 w-full rounded-lg border border-border bg-card shadow-lg"
          >
            <div className="max-h-96 overflow-y-auto p-2">
              {suggestions.map((suggestion, index) => (
                <button
                  key={suggestion.id}
                  onClick={() => {
                    if (suggestion.type === 'page' && suggestion.href) {
                      setIsOpen(false)
                      router.push(suggestion.href)
                      setQuery('')
                      inputRef.current?.blur()
                    } else {
                      handleSelect(suggestion.text)
                    }
                  }}
                  onMouseEnter={() => setSelectedIndex(index)}
                  className={cn(
                    'flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left text-sm transition-colors',
                    selectedIndex === index
                      ? 'bg-accent text-accent-foreground'
                      : 'hover:bg-accent hover:text-accent-foreground'
                  )}
                >
                  {suggestion.type === 'history' && <ClockIcon className="h-4 w-4 shrink-0" />}
                  {suggestion.type === 'page' && <DocumentTextIcon className="h-4 w-4 shrink-0" />}
                  <span className="flex-1 truncate">{suggestion.text}</span>
                  {suggestion.module && (
                    <span className="text-xs text-muted-foreground capitalize shrink-0">{suggestion.module}</span>
                  )}
                  {suggestion.type === 'page' && (
                    <span className="text-xs text-muted-foreground shrink-0">Page</span>
                  )}
                </button>
              ))}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

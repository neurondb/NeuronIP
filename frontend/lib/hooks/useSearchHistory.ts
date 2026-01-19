import { useState, useEffect } from 'react'
import { useLocalStorage } from './useLocalStorage'

interface SearchHistoryItem {
  id: string
  query: string
  timestamp: Date
  module?: string
}

export function useSearchHistory(maxItems: number = 10) {
  const [history, setHistory] = useLocalStorage<SearchHistoryItem[]>('neuronip-search-history', [])

  const addToHistory = (query: string, module?: string) => {
    if (!query.trim()) return

    const newItem: SearchHistoryItem = {
      id: Date.now().toString(),
      query: query.trim(),
      timestamp: new Date(),
      module,
    }

    setHistory((prev) => {
      // Remove duplicates
      const filtered = prev.filter((item) => item.query.toLowerCase() !== query.toLowerCase())
      // Add new item at the beginning
      const updated = [newItem, ...filtered]
      // Keep only max items
      return updated.slice(0, maxItems)
    })
  }

  const removeFromHistory = (id: string) => {
    setHistory((prev) => prev.filter((item) => item.id !== id))
  }

  const clearHistory = () => {
    setHistory([])
  }

  return {
    history,
    addToHistory,
    removeFromHistory,
    clearHistory,
  }
}

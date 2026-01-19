'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { TrashIcon, EyeIcon, EyeSlashIcon } from '@heroicons/react/24/outline'
import { useAPIKeys, useDeleteAPIKey } from '@/lib/api/queries'
import { showToast } from '@/components/ui/Toast'
import { cn } from '@/lib/utils/cn'
import { slideUp, transition } from '@/lib/animations/variants'

interface APIKey {
  id: string
  name: string
  key: string
  rate_limit: number
  created_at: Date
}

export default function APIKeyList() {
  const [visibleKeys, setVisibleKeys] = useState<Set<string>>(new Set())
  const { data: apiKeys, isLoading } = useAPIKeys()
  const { mutate: deleteKey } = useDeleteAPIKey()

  const toggleVisibility = (id: string) => {
    setVisibleKeys((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const handleDelete = (id: string, name: string) => {
    if (confirm(`Are you sure you want to delete API key "${name}"?`)) {
      deleteKey(id, {
        onSuccess: () => {
          showToast('API key deleted successfully', 'success')
        },
        onError: () => {
          showToast('Failed to delete API key', 'error')
        },
      })
    }
  }

  const maskKey = (key: string) => {
    if (key.length <= 8) return '••••••••'
    return key.substring(0, 4) + '••••••••' + key.substring(key.length - 4)
  }

  const keyList: APIKey[] = apiKeys?.keys || [
    {
      id: '1',
      name: 'Production API Key',
      key: 'sk_live_1234567890abcdef',
      rate_limit: 1000,
      created_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30),
    },
    {
      id: '2',
      name: 'Development API Key',
      key: 'sk_test_abcdef1234567890',
      rate_limit: 100,
      created_at: new Date(Date.now() - 1000 * 60 * 60 * 24 * 7),
    },
  ]

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <CardTitle className="text-base sm:text-lg">API Keys</CardTitle>
        <CardDescription className="text-xs">Manage your API keys and rate limits</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 overflow-y-auto min-h-0">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loading size="md" />
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>Rate Limit</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keyList.map((apiKey, index) => (
                <motion.tr
                  key={apiKey.id}
                  variants={slideUp}
                  initial="hidden"
                  animate="visible"
                  transition={{ ...transition, delay: index * 0.03 }}
                  className="hover:bg-muted/50 transition-colors"
                >
                  <TableCell className="font-medium">{apiKey.name}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <code className="text-xs font-mono bg-muted px-2 py-1 rounded">
                        {visibleKeys.has(apiKey.id) ? apiKey.key : maskKey(apiKey.key)}
                      </code>
                      <button
                        onClick={() => toggleVisibility(apiKey.id)}
                        className="p-1 hover:bg-accent rounded transition-colors"
                      >
                        {visibleKeys.has(apiKey.id) ? (
                          <EyeSlashIcon className="h-4 w-4" />
                        ) : (
                          <EyeIcon className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  </TableCell>
                  <TableCell className="text-sm">{apiKey.rate_limit}/hour</TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(apiKey.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => handleDelete(apiKey.id, apiKey.name)}
                      className="text-destructive hover:text-destructive"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </motion.tr>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}

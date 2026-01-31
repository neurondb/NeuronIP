'use client'

import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'

import AIWritingAssistant from '@/components/ai/AIWritingAssistant'
import { LazyBlockEditor } from '@/components/blocks'
import LiveCursors from '@/components/collaboration/LiveCursors'
import DatabaseView from '@/components/database/DatabaseView'
import { PresenceIndicator } from '@/components/presence/PresenceIndicator'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Loading from '@/components/ui/Loading'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import { showToast } from '@/components/ui/Toast'
import { staggerContainer, slideUp } from '@/lib/animations/variants'
import { useBlocks, useCreateBlock, useUpdateBlock, useDeleteBlock } from '@/lib/hooks/useBlocks'
import { useKeyboardShortcuts } from '@/lib/hooks/useKeyboardShortcuts'

// Mock data for database view
const mockDatabaseData = [
  {
    id: '1',
    title: 'Project Alpha',
    status: 'In Progress',
    assignee: 'John Doe',
    dueDate: '2024-02-15',
    priority: 'High',
  },
  {
    id: '2',
    title: 'Project Beta',
    status: 'Todo',
    assignee: 'Jane Smith',
    dueDate: '2024-02-20',
    priority: 'Medium',
  },
  {
    id: '3',
    title: 'Project Gamma',
    status: 'Done',
    assignee: 'Bob Johnson',
    dueDate: '2024-02-10',
    priority: 'Low',
  },
]

const mockColumns = [
  { id: 'title', name: 'Title', type: 'text' as const },
  { id: 'status', name: 'Status', type: 'select' as const, options: ['Todo', 'In Progress', 'Done'] },
  { id: 'assignee', name: 'Assignee', type: 'person' as const },
  { id: 'dueDate', name: 'Due Date', type: 'date' as const },
  { id: 'priority', name: 'Priority', type: 'select' as const, options: ['Low', 'Medium', 'High'] },
]

export default function NotionUIPage() {
  const [activeTab, setActiveTab] = useState<'editor' | 'database'>('editor')
  const [editorContent, setEditorContent] = useState('')
  const [databaseData, setDatabaseData] = useState(mockDatabaseData)
  const [editorInstance, setEditorInstance] = useState<unknown>(null)
  const [pageId] = useState<string>('demo-page-1') // In production, get from route or create new

  // Blocks API hooks
  const { data: blocks, isLoading: blocksLoading, error: blocksError, refetch: _refetchBlocks } = useBlocks(pageId)
  const createBlock = useCreateBlock()
  const _updateBlock = useUpdateBlock()
  const _deleteBlock = useDeleteBlock()

  // Load blocks content into editor
  useEffect(() => {
    if (blocks && blocks.length > 0 && editorInstance) {
      // Convert blocks to HTML content
      const htmlContent = blocks
        .sort((a, b) => a.order - b.order)
        .map((block) => {
          if (typeof block.content === 'string') {
            return block.content
          }
          return JSON.stringify(block.content)
        })
        .join('')
      setEditorContent(htmlContent)
    }
  }, [blocks, editorInstance])

  // Enable keyboard shortcuts for editor
  useKeyboardShortcuts(
    [
      {
        key: 'k',
        metaKey: true,
        action: () => {
          // Trigger command palette
          const event = new CustomEvent('open-command-palette')
          window.dispatchEvent(event)
        },
      },
      {
        key: 'b',
        metaKey: true,
        action: () => {
          // Trigger bold in editor
          const event = new CustomEvent('editor-format', { detail: { format: 'bold' } })
          window.dispatchEvent(event)
        },
      },
      {
        key: 'i',
        metaKey: true,
        action: () => {
          // Trigger italic in editor
          const event = new CustomEvent('editor-format', { detail: { format: 'italic' } })
          window.dispatchEvent(event)
        },
      },
    ],
    true
  )

  const handleDatabaseUpdate = (rowId: string, columnId: string, value: unknown) => {
    setDatabaseData((prev) =>
      prev.map((row) => (row.id === rowId ? { ...row, [columnId]: value } : row))
    )
  }

  const handleDatabaseDelete = (rowId: string) => {
    setDatabaseData((prev) => prev.filter((row) => row.id !== rowId))
  }

  const handleDatabaseCreate = (data: Record<string, unknown>) => {
    const newRow: typeof mockDatabaseData[0] = {
      id: String(Date.now()),
      title: String(data.title ?? 'Untitled'),
      status: String(data.status ?? 'Todo'),
      assignee: String(data.assignee ?? ''),
      dueDate: String(data.dueDate ?? ''),
      priority: String(data.priority ?? 'Medium'),
    }
    setDatabaseData((prev) => [...prev, newRow])
  }

  const handleEditorChange = (content: string) => {
    setEditorContent(content)
    // Auto-save could be implemented here with debouncing
  }

  const _handleBlockCreate = async () => {
    try {
      await createBlock.mutateAsync({
        page_id: pageId,
        type: 'paragraph',
        content: { text: '' },
      })
      showToast('Block created', 'success')
    } catch (error) {
      showToast('Failed to create block', 'error')
    }
  }

  return (
    <motion.div
      variants={staggerContainer}
      initial="hidden"
      animate="visible"
      className="space-y-4 lg:space-y-6 flex flex-col h-full"
    >
      {/* Page Header */}
      <motion.div variants={slideUp} className="flex-shrink-0 pb-2">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground">
              Block-based Editor
            </h1>
            <p className="text-sm sm:text-base text-muted-foreground mt-1">
              Block-based editing with slash commands, database views, and real-time collaboration
            </p>
          </div>
          <PresenceIndicator roomId="notion-ui-page" variant="expanded" />
        </div>
      </motion.div>

      {/* Main Content */}
      <div className="flex-1 min-h-0">
        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as typeof activeTab)}>
          <TabsList className="mb-4">
            <TabsTrigger value="editor">Block Editor</TabsTrigger>
            <TabsTrigger value="database">Database Views</TabsTrigger>
          </TabsList>

          <TabsContent value="editor" className="flex-1 min-h-0 flex flex-col">
            <Card className="flex-1 flex flex-col min-h-0">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>Block-Based Editor</CardTitle>
                  <div className="flex items-center gap-2">
                    <AIWritingAssistant
                      editor={editorInstance}
                      onSuggestion={() => {}}
                    />
                    <PresenceIndicator roomId="notion-ui-editor" maxVisible={3} />
                  </div>
                </div>
              </CardHeader>
              <CardContent className="flex-1 flex flex-col min-h-0">
                {blocksLoading ? (
                  <div className="flex items-center justify-center h-full">
                    <Loading />
                  </div>
                ) : blocksError ? (
                  <div className="flex items-center justify-center h-full text-destructive">
                    Error loading blocks. Using local editor.
                  </div>
                ) : (
                  <div className="flex-1 min-h-0 relative">
                    <LiveCursors roomId="notion-ui-editor" currentUserId="current-user" />
                    <LazyBlockEditor
                      content={editorContent}
                      onChange={handleEditorChange}
                      placeholder="Type '/' for commands, or start writing..."
                      editable={true}
                      showToolbar={true}
                      className="h-full"
                      onEditorReady={setEditorInstance}
                    />
                  </div>
                )}
                <div className="mt-4 text-xs text-muted-foreground">
                  <p>Try typing &quot;/&quot; to see slash commands, or use keyboard shortcuts:</p>
                  <ul className="list-disc list-inside mt-1 space-y-1">
                    <li>Cmd/Ctrl + K: Open command palette</li>
                    <li>Cmd/Ctrl + /: Open slash commands</li>
                    <li>Cmd/Ctrl + B: Bold text</li>
                    <li>Cmd/Ctrl + I: Italic text</li>
                  </ul>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="database" className="flex-1 min-h-0 flex flex-col">
            <Card className="flex-1 flex flex-col min-h-0">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>Database Views</CardTitle>
                  <PresenceIndicator roomId="notion-ui-database" maxVisible={3} />
                </div>
              </CardHeader>
              <CardContent className="flex-1 flex flex-col min-h-0">
                <div className="flex-1 min-h-0">
                  <DatabaseView
                    data={databaseData}
                    columns={mockColumns}
                    onUpdate={handleDatabaseUpdate}
                    onDelete={handleDatabaseDelete}
                    onCreate={handleDatabaseCreate}
                    defaultView="table"
                    className="h-full"
                  />
                </div>
                <div className="mt-4 text-xs text-muted-foreground">
                  <p>Switch between different view types: Table, Kanban, Calendar, Gallery, and List</p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </motion.div>
  )
}

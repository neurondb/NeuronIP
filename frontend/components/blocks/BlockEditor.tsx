'use client'

import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import Table from '@tiptap/extension-table'
import TableCell from '@tiptap/extension-table-cell'
import TableHeader from '@tiptap/extension-table-header'
import TableRow from '@tiptap/extension-table-row'
import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'
import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import { useState, useEffect, useCallback, useRef } from 'react'

import { cn } from '@/lib/utils/cn'

import { getCommandById } from '../commands/CommandRegistry'
import SlashCommandMenu from '../commands/SlashCommandMenu'

interface BlockEditorProps {
  content?: string
  onChange?: (content: string) => void
  placeholder?: string
  className?: string
  editable?: boolean
  showToolbar?: boolean
  onEditorReady?: (editor: any) => void
}

export function BlockEditor({
  content = '',
  onChange,
  placeholder = 'Type "/" for commands...',
  className,
  editable = true,
  showToolbar = false,
  onEditorReady,
}: BlockEditorProps) {
  const [slashCommandQuery, setSlashCommandQuery] = useState('')
  const [slashCommandPosition, setSlashCommandPosition] = useState<{
    top: number
    left: number
  } | null>(null)

  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: {
          levels: [1, 2, 3, 4, 5, 6],
        },
      }),
      Placeholder.configure({
        placeholder,
      }),
      Image.configure({
        inline: true,
        allowBase64: true,
      }),
      Link.configure({
        openOnClick: false,
        HTMLAttributes: {
          class: 'text-primary underline',
        },
      }),
      Table.configure({
        resizable: true,
      }),
      TableRow,
      TableHeader,
      TableCell,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
    ],
    content,
    editable,
    onUpdate: ({ editor }) => {
      onChange?.(editor.getHTML())
    },
    editorProps: {
      handleKeyDown: (view, event) => {
        // Handle slash command
        if (event.key === '/') {
          const { selection } = view.state
          const { $from } = selection
          const textBefore = $from.nodeBefore?.textContent || ''
          
          // Only trigger if at start of line or after space
          if (textBefore === '' || textBefore.endsWith(' ')) {
            const coords = view.coordsAtPos($from.pos)
            setSlashCommandPosition({
              top: coords.top + 20,
              left: coords.left,
            })
            setSlashCommandQuery('')
            return true
          }
        }

        // Handle escape to close slash menu
        if (event.key === 'Escape' && slashCommandPosition) {
          setSlashCommandPosition(null)
          setSlashCommandQuery('')
          return true
        }

        return false
      },
    },
  })

  // Update slash command query as user types
  useEffect(() => {
    if (!editor || !slashCommandPosition) return

    const updateSlashQuery = () => {
      const { selection } = editor.state
      const { $from } = selection
      const textBefore = $from.nodeBefore?.textContent || ''
      
      if (textBefore.startsWith('/')) {
        const query = textBefore.slice(1).trim()
        setSlashCommandQuery(query)
      } else {
        setSlashCommandPosition(null)
        setSlashCommandQuery('')
      }
    }

    editor.on('selectionUpdate', updateSlashQuery)
    editor.on('update', updateSlashQuery)

    return () => {
      editor.off('selectionUpdate', updateSlashQuery)
      editor.off('update', updateSlashQuery)
    }
  }, [editor, slashCommandPosition])

  const handleSlashCommandSelect = useCallback(
    (commandId: string) => {
      if (!editor || !commandId) return

      const command = getCommandById(commandId)
      if (!command) return

      // Remove the slash command text
      const { selection } = editor.state
      const { $from } = selection
      const textBefore = $from.nodeBefore?.textContent || ''
      
      if (textBefore.startsWith('/')) {
        const deleteFrom = $from.pos - textBefore.length
        editor.chain().focus().deleteRange({ from: deleteFrom, to: $from.pos }).run()
      }

      // Execute the command
      command.action(editor)

      // Close the menu
      setSlashCommandPosition(null)
      setSlashCommandQuery('')
    },
    [editor]
  )

  useEffect(() => {
    if (editor && onEditorReady) {
      onEditorReady(editor)
    }
  }, [editor, onEditorReady])

  // Handle keyboard shortcuts for editor
  useEffect(() => {
    if (!editor) return

    const handleKeyDown = (event: KeyboardEvent) => {
      // Handle Cmd/Ctrl + B for bold
      if ((event.metaKey || event.ctrlKey) && event.key === 'b') {
        event.preventDefault()
        editor.chain().focus().toggleBold().run()
      }
      // Handle Cmd/Ctrl + I for italic
      if ((event.metaKey || event.ctrlKey) && event.key === 'i') {
        event.preventDefault()
        editor.chain().focus().toggleItalic().run()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [editor])

  if (!editor) {
    return null
  }

  return (
    <div className={cn('relative', className)}>
      <div className="border rounded-lg overflow-hidden bg-background">
        {editable && showToolbar && (
          <div className="border-b p-2 flex gap-1 flex-wrap bg-muted/50">
            <button
              onClick={() => editor.chain().focus().toggleBold().run()}
              className={cn(
                'px-2 py-1 rounded text-sm transition-colors',
                editor.isActive('bold')
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent/50 text-muted-foreground'
              )}
            >
              <strong>B</strong>
            </button>
            <button
              onClick={() => editor.chain().focus().toggleItalic().run()}
              className={cn(
                'px-2 py-1 rounded text-sm transition-colors',
                editor.isActive('italic')
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent/50 text-muted-foreground'
              )}
            >
              <em>I</em>
            </button>
            <div className="w-px bg-border mx-1" />
            <button
              onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
              className={cn(
                'px-2 py-1 rounded text-sm transition-colors',
                editor.isActive('heading', { level: 1 })
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent/50 text-muted-foreground'
              )}
            >
              H1
            </button>
            <button
              onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
              className={cn(
                'px-2 py-1 rounded text-sm transition-colors',
                editor.isActive('heading', { level: 2 })
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent/50 text-muted-foreground'
              )}
            >
              H2
            </button>
            <button
              onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
              className={cn(
                'px-2 py-1 rounded text-sm transition-colors',
                editor.isActive('heading', { level: 3 })
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent/50 text-muted-foreground'
              )}
            >
              H3
            </button>
            <div className="w-px bg-border mx-1" />
            <button
              onClick={() => editor.chain().focus().toggleBulletList().run()}
              className={cn(
                'px-2 py-1 rounded text-sm transition-colors',
                editor.isActive('bulletList')
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent/50 text-muted-foreground'
              )}
            >
              •
            </button>
            <button
              onClick={() => editor.chain().focus().toggleOrderedList().run()}
              className={cn(
                'px-2 py-1 rounded text-sm transition-colors',
                editor.isActive('orderedList')
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent/50 text-muted-foreground'
              )}
            >
              1.
            </button>
            <button
              onClick={() => editor.chain().focus().toggleCodeBlock().run()}
              className={cn(
                'px-2 py-1 rounded text-sm transition-colors',
                editor.isActive('codeBlock')
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-accent/50 text-muted-foreground'
              )}
            >
              {'</>'}
            </button>
          </div>
        )}
        <EditorContent
          editor={editor}
          className={cn(
            'prose prose-sm sm:prose-base lg:prose-lg xl:prose-2xl mx-auto focus:outline-none min-h-[200px] p-4',
            'prose-headings:font-bold prose-headings:text-foreground',
            'prose-p:text-foreground prose-p:leading-relaxed',
            'prose-code:text-foreground prose-code:bg-muted prose-code:px-1 prose-code:rounded',
            'prose-pre:bg-muted prose-pre:border prose-pre:border-border',
            'prose-blockquote:border-l-primary prose-blockquote:text-muted-foreground',
            '[&_.ProseMirror]:outline-none'
          )}
        />
      </div>
      {slashCommandPosition && (
        <SlashCommandMenu
          editor={editor}
          query={slashCommandQuery}
          onSelect={handleSlashCommandSelect}
          position={slashCommandPosition}
        />
      )}
    </div>
  )
}

import {
  DocumentTextIcon as Heading1,
  ListBulletIcon,
  ListBulletIcon as ListOrderedIcon,
  CheckCircleIcon,
  CodeBracketIcon,
  ChatBubbleLeftRightIcon,
  MinusIcon,
  PhotoIcon,
  LinkIcon,
  TableCellsIcon,
  CircleStackIcon,
} from '@heroicons/react/24/outline'

// Use Heading1 for all heading levels since Heading2/3 don't exist
const Heading2 = Heading1
const Heading3 = Heading1
import type { SlashCommand } from '../blocks/BlockTypes'

export const slashCommands: SlashCommand[] = [
  // Text blocks
  {
    id: 'heading1',
    label: 'Heading 1',
    description: 'Large section heading',
    icon: Heading1,
    keywords: ['h1', 'heading', 'title'],
    group: 'text',
    action: (editor) => {
      editor.chain().focus().toggleHeading({ level: 1 }).run()
    },
  },
  {
    id: 'heading2',
    label: 'Heading 2',
    description: 'Medium section heading',
    icon: Heading2,
    keywords: ['h2', 'heading', 'subtitle'],
    group: 'text',
    action: (editor) => {
      editor.chain().focus().toggleHeading({ level: 2 }).run()
    },
  },
  {
    id: 'heading3',
    label: 'Heading 3',
    description: 'Small section heading',
    icon: Heading3,
    keywords: ['h3', 'heading'],
    group: 'text',
    action: (editor) => {
      editor.chain().focus().toggleHeading({ level: 3 }).run()
    },
  },
  {
    id: 'paragraph',
    label: 'Text',
    description: 'Just start typing with plain text',
    keywords: ['text', 'paragraph', 'p'],
    group: 'text',
    action: (editor) => {
      editor.chain().focus().setParagraph().run()
    },
  },
  {
    id: 'quote',
    label: 'Quote',
    description: 'Capture a quote',
    icon: ChatBubbleLeftRightIcon,
    keywords: ['quote', 'blockquote'],
    group: 'text',
    action: (editor) => {
      editor.chain().focus().toggleBlockquote().run()
    },
  },
  {
    id: 'code',
    label: 'Code',
    description: 'Capture a code snippet',
    icon: CodeBracketIcon,
    keywords: ['code', 'codeblock'],
    group: 'text',
    action: (editor) => {
      editor.chain().focus().toggleCodeBlock().run()
    },
  },
  {
    id: 'divider',
    label: 'Divider',
    description: 'Visually divide blocks',
    icon: MinusIcon,
    keywords: ['divider', 'hr', 'horizontal'],
    group: 'text',
    action: (editor) => {
      editor.chain().focus().setHorizontalRule().run()
    },
  },
  // List blocks
  {
    id: 'bullet',
    label: 'Bulleted List',
    description: 'Create a simple bulleted list',
    icon: ListBulletIcon,
    keywords: ['bullet', 'ul', 'unordered'],
    group: 'lists',
    action: (editor) => {
      editor.chain().focus().toggleBulletList().run()
    },
  },
  {
    id: 'number',
    label: 'Numbered List',
    description: 'Create a list with numbering',
    icon: ListOrderedIcon,
    keywords: ['number', 'ol', 'ordered'],
    group: 'lists',
    action: (editor) => {
      editor.chain().focus().toggleOrderedList().run()
    },
  },
  {
    id: 'todo',
    label: 'To-do List',
    description: 'Track tasks with a to-do list',
    icon: CheckCircleIcon,
    keywords: ['todo', 'task', 'checkbox'],
    group: 'lists',
    action: (editor) => {
      editor.chain().focus().toggleTaskList().run()
    },
  },
  // Media blocks
  {
    id: 'image',
    label: 'Image',
    description: 'Upload or embed an image',
    icon: PhotoIcon,
    keywords: ['image', 'img', 'picture', 'photo'],
    group: 'media',
    action: (editor) => {
      // Trigger image upload
      const input = document.createElement('input')
      input.type = 'file'
      input.accept = 'image/*'
      input.onchange = (e) => {
        const file = (e.target as HTMLInputElement).files?.[0]
        if (file) {
          const reader = new FileReader()
          reader.onload = (event) => {
            const src = event.target?.result as string
            editor.chain().focus().setImage({ src }).run()
          }
          reader.readAsDataURL(file)
        }
      }
      input.click()
    },
  },
  {
    id: 'embed',
    label: 'Embed',
    description: 'Embed a link, video, or file',
    icon: LinkIcon,
    keywords: ['embed', 'link', 'video', 'iframe'],
    group: 'media',
    action: (editor) => {
      const url = prompt('Enter URL to embed:')
      if (url) {
        // Create a link
        editor.chain().focus().setLink({ href: url, target: '_blank' }).run()
      }
    },
  },
  // Advanced blocks
  {
    id: 'table',
    label: 'Table',
    description: 'Insert a table',
    icon: TableCellsIcon,
    keywords: ['table', 'grid'],
    group: 'advanced',
    action: (editor) => {
      editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
    },
  },
  {
    id: 'database',
    label: 'Database',
    description: 'Create a database view',
    icon: CircleStackIcon,
    keywords: ['database', 'db', 'data'],
    group: 'advanced',
    action: (editor) => {
      // This would trigger database creation UI
      console.log('Database creation not yet implemented')
    },
  },
]

export function searchCommands(query: string): SlashCommand[] {
  const lowerQuery = query.toLowerCase()
  return slashCommands.filter((cmd) => {
    const matchesLabel = cmd.label.toLowerCase().includes(lowerQuery)
    const matchesDescription = cmd.description.toLowerCase().includes(lowerQuery)
    const matchesKeywords = cmd.keywords.some((kw) => kw.toLowerCase().includes(lowerQuery))
    return matchesLabel || matchesDescription || matchesKeywords
  })
}

export function getCommandById(id: string): SlashCommand | undefined {
  return slashCommands.find((cmd) => cmd.id === id)
}

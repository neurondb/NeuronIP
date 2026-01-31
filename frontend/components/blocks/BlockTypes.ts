export type BlockType =
  | 'paragraph'
  | 'heading1'
  | 'heading2'
  | 'heading3'
  | 'heading4'
  | 'heading5'
  | 'heading6'
  | 'bulletList'
  | 'orderedList'
  | 'checklist'
  | 'codeBlock'
  | 'quote'
  | 'divider'
  | 'image'
  | 'embed'
  | 'table'
  | 'database'

export interface Block {
  id: string
  type: BlockType
  content: string | Record<string, any>
  order: number
  parentId?: string
  children?: Block[]
  metadata?: Record<string, any>
}

export interface SlashCommand {
  id: string
  label: string
  description: string
  icon?: React.ComponentType<{ className?: string }>
  keywords: string[]
  action: (editor: any) => void
  group: 'text' | 'lists' | 'media' | 'advanced'
}

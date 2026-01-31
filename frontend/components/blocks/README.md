# Block Editor Components

This directory contains components for implementing a block-based editing experience in NeuronIP.

## Components

### BlockEditor
The main block-based editor component with slash commands support.

```tsx
import { BlockEditor } from '@/components/blocks'

<BlockEditor
  content={htmlContent}
  onChange={(content) => setContent(content)}
  placeholder="Type '/' for commands..."
  editable={true}
  showToolbar={true}
  onEditorReady={(editor) => {
    // Access TipTap editor instance
  }}
/>
```

### SlashCommandMenu
Inline menu that appears when typing `/` in the editor.

### BlockDragHandle
Drag handle for reordering blocks (used with @dnd-kit).

### BlockMenu
Context menu for block actions (delete, duplicate, move, comments).

### BlockComments
Block-level commenting system with resolve/delete functionality.

## Database Views

### DatabaseView
Main component for switching between different database view types.

```tsx
import DatabaseView from '@/components/database/DatabaseView'

<DatabaseView
  data={rows}
  columns={columns}
  onUpdate={(rowId, columnId, value) => {}}
  onDelete={(rowId) => {}}
  onCreate={(data) => {}}
  defaultView="table"
/>
```

### View Types
- **TableView**: Standard table view with sorting and filtering
- **KanbanView**: Kanban board with drag-and-drop between columns
- **CalendarView**: Calendar view using react-big-calendar
- **GalleryView**: Card-based gallery layout
- **ListView**: Detailed list view with expandable cards

## Collaboration Features

### LiveCursors
Real-time cursor tracking showing other users' positions.

```tsx
import LiveCursors from '@/components/collaboration/LiveCursors'

<LiveCursors
  roomId="page-123"
  currentUserId="user-456"
/>
```

### PresenceIndicator
Shows active users viewing the page.

```tsx
import { PresenceIndicator } from '@/components/presence/PresenceIndicator'

<PresenceIndicator
  roomId="page-123"
  variant="expanded"
  showCount={true}
/>
```

## AI Integration

### AIWritingAssistant
AI-powered writing suggestions and block generation.

```tsx
import AIWritingAssistant from '@/components/ai/AIWritingAssistant'

<AIWritingAssistant
  editor={editorInstance}
  onSuggestion={(suggestion) => {}}
/>
```

## Keyboard Shortcuts

Use the `useKeyboardShortcuts` hook to enable block-based shortcuts:

```tsx
import { useKeyboardShortcuts } from '@/lib/hooks/useKeyboardShortcuts'

useKeyboardShortcuts([
  {
    key: 'k',
    metaKey: true,
    action: () => openCommandPalette(),
  },
  {
    key: 'b',
    metaKey: true,
    action: () => editor.chain().focus().toggleBold().run(),
  },
], true)
```

## API Integration

### Blocks API
```tsx
import { useBlocks, useCreateBlock, useUpdateBlock, useDeleteBlock } from '@/lib/hooks/useBlocks'

const { data: blocks } = useBlocks(pageId)
const createBlock = useCreateBlock()
const updateBlock = useUpdateBlock()
const deleteBlock = useDeleteBlock()
```

### Database API
```tsx
import { useDatabase, useUpdateRow, useCreateRow, useDeleteRow } from '@/lib/hooks/useDatabase'

const { data: database } = useDatabase(databaseId)
const updateRow = useUpdateRow()
const createRow = useCreateRow()
const deleteRow = useDeleteRow()
```

## Usage Example

These components can be integrated into any page where block-based editing or database views are needed. The components are designed to work independently and can be combined as needed for your use case.

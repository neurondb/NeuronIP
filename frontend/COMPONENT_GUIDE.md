# Component Guide

Complete reference for all UI components in the NeuronIP frontend.

## Table of Contents

1. [Form Components](#form-components)
2. [Layout Components](#layout-components)
3. [Data Display](#data-display)
4. [Feedback Components](#feedback-components)
5. [Navigation Components](#navigation-components)
6. [Overlay Components](#overlay-components)
7. [Advanced Components](#advanced-components)

## Form Components

### Button

A versatile button component with multiple variants and sizes.

```tsx
import { Button } from '@/components/ui'

<Button variant="primary" size="md">Click me</Button>
<Button variant="outline" size="sm">Outline</Button>
<Button variant="destructive" isLoading>Loading</Button>
```

**Props:**
- `variant`: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive'
- `size`: 'sm' | 'md' | 'lg'
- `isLoading`: boolean
- `icon`: ReactNode
- `iconPosition`: 'left' | 'right'

### Input

Text input with label, error, and helper text support.

```tsx
import { Input } from '@/components/ui'

<Input label="Email" type="email" required />
<Input label="Password" type="password" error="Invalid password" />
```

### FormBuilder

Dynamic form generator with Zod validation.

```tsx
import { FormBuilder, type FormFieldConfig } from '@/components/ui'

const fields: FormFieldConfig[] = [
  {
    name: 'email',
    label: 'Email',
    type: 'email',
    required: true,
    validation: z.string().email(),
  },
]

<FormBuilder fields={fields} onSubmit={handleSubmit} />
```

### Select, Checkbox, Switch, RadioGroup

Standard form controls with consistent styling.

## Layout Components

### Card

Container component with header, content, and footer.

```tsx
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui'

<Card>
  <CardHeader>
    <CardTitle>Title</CardTitle>
  </CardHeader>
  <CardContent>Content</CardContent>
</Card>
```

### Separator

Visual divider component.

```tsx
<Separator orientation="horizontal" />
```

## Data Display

### DataTable

Advanced table with sorting, filtering, pagination.

```tsx
import { DataTable } from '@/components/ui'
import { ColumnDef } from '@tanstack/react-table'

const columns: ColumnDef<Data>[] = [
  { accessorKey: 'name', header: 'Name' },
  { accessorKey: 'email', header: 'Email' },
]

<DataTable columns={columns} data={data} searchKey="name" />
```

### Table

Basic table components for custom implementations.

```tsx
import { Table, TableHeader, TableBody, TableRow, TableCell } from '@/components/ui'

<Table>
  <TableHeader>
    <TableRow>
      <TableCell>Header</TableCell>
    </TableRow>
  </TableHeader>
  <TableBody>
    <TableRow>
      <TableCell>Data</TableCell>
    </TableRow>
  </TableBody>
</Table>
```

### Badge, Avatar, Skeleton

Display components for status, users, and loading states.

## Feedback Components

### Toast (Sonner)

Toast notifications via Sonner.

```tsx
import { toast } from 'sonner'

toast.success('Success!')
toast.error('Error occurred')
toast.info('Information')
```

### Progress

Progress indicator.

```tsx
<Progress value={75} />
```

### Loading

Loading spinner component.

## Navigation Components

### Tabs

Tab navigation component.

```tsx
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui'

<Tabs defaultValue="tab1">
  <TabsList>
    <TabsTrigger value="tab1">Tab 1</TabsTrigger>
    <TabsTrigger value="tab2">Tab 2</TabsTrigger>
  </TabsList>
  <TabsContent value="tab1">Content 1</TabsContent>
  <TabsContent value="tab2">Content 2</TabsContent>
</Tabs>
```

### Accordion

Collapsible content sections.

```tsx
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from '@/components/ui'

<Accordion>
  <AccordionItem value="item1">
    <AccordionTrigger>Title</AccordionTrigger>
    <AccordionContent>Content</AccordionContent>
  </AccordionItem>
</Accordion>
```

## Overlay Components

### Dialog

Modal dialog component.

```tsx
import { Dialog, DialogTrigger, DialogContent, DialogHeader, DialogTitle } from '@/components/ui'

<Dialog>
  <DialogTrigger>Open</DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Title</DialogTitle>
    </DialogHeader>
    Content
  </DialogContent>
</Dialog>
```

### Sheet

Slide-over panel.

```tsx
import { Sheet, SheetTrigger, SheetContent, SheetHeader, SheetTitle } from '@/components/ui'

<Sheet>
  <SheetTrigger>Open</SheetTrigger>
  <SheetContent side="right">
    <SheetHeader>
      <SheetTitle>Title</SheetTitle>
    </SheetHeader>
    Content
  </SheetContent>
</Sheet>
```

### Popover

Popover component.

```tsx
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui'

<Popover>
  <PopoverTrigger>Trigger</PopoverTrigger>
  <PopoverContent>Content</PopoverContent>
</Popover>
```

## Advanced Components

### CommandPalette

Cmd+K search interface.

```tsx
import { CommandDialog, CommandInput, CommandList, CommandItem } from '@/components/ui'

<CommandDialog open={open} onOpenChange={setOpen}>
  <CommandInput placeholder="Search..." />
  <CommandList>
    <CommandItem>Item 1</CommandItem>
  </CommandList>
</CommandDialog>
```

### RichTextEditor

WYSIWYG editor with Tiptap.

```tsx
import { RichTextEditor } from '@/components/ui'

<RichTextEditor
  content={content}
  onChange={setContent}
  placeholder="Start typing..."
/>
```

### CodeEditor

Monaco editor for code editing.

```tsx
import { CodeEditor } from '@/components/ui'

<CodeEditor
  value={code}
  onChange={setCode}
  language="typescript"
  theme="vs-dark"
/>
```

### FileUpload

Drag-and-drop file upload.

```tsx
import { FileUpload } from '@/components/ui'

<FileUpload
  onFilesSelected={(files) => console.log(files)}
  maxSize={10 * 1024 * 1024}
  maxFiles={5}
/>
```

### MultiSelect

Multi-select dropdown with search.

```tsx
import { MultiSelect } from '@/components/ui'

<MultiSelect
  options={options}
  selected={selected}
  onChange={setSelected}
/>
```

### DatePicker

Date picker with calendar.

```tsx
import { DatePicker } from '@/components/ui'

<DatePicker
  date={date}
  onDateChange={setDate}
  placeholder="Pick a date"
/>
```

## Hooks

### useDebounce

Debounce a value.

```tsx
import { useDebounce } from '@/lib/hooks'

const debouncedValue = useDebounce(value, 500)
```

### useLocalStorage

Sync state with localStorage.

```tsx
import { useLocalStorage } from '@/lib/hooks'

const [value, setValue] = useLocalStorage('key', 'default')
```

### useWebSocket

WebSocket connection hook.

```tsx
import { useWebSocket } from '@/lib/hooks'

const { status, send, subscribe } = useWebSocket()
```

### useRealtimeQuery

Real-time data synchronization.

```tsx
import { useRealtimeQuery } from '@/lib/hooks'

const { data } = useRealtimeQuery(['key'], 'message-type', initialData)
```

## Styling

All components use Tailwind CSS with CSS variables for theming. Customize colors in `tailwind.config.js` and `globals.css`.

## Accessibility

All components follow WCAG AA guidelines:
- Proper ARIA labels
- Keyboard navigation
- Focus management
- Screen reader support

## TypeScript

All components are fully typed. Import types as needed:

```tsx
import type { FormFieldConfig, Option } from '@/components/ui'
```

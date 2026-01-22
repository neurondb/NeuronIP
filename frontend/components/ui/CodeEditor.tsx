'use client'

import dynamic from 'next/dynamic'
import { cn } from '@/lib/utils/cn'

const MonacoEditor = dynamic(() => import('@monaco-editor/react'), { ssr: false })

interface CodeEditorProps {
  value?: string
  onChange?: (value: string | undefined) => void
  language?: string
  theme?: 'vs-dark' | 'light'
  height?: string
  className?: string
  readOnly?: boolean
}

export function CodeEditor({
  value = '',
  onChange,
  language = 'typescript',
  theme = 'vs-dark',
  height = '400px',
  className,
  readOnly = false,
}: CodeEditorProps) {
  return (
    <div className={cn('border rounded-lg overflow-hidden', className)}>
      <MonacoEditor
        height={height}
        language={language}
        theme={theme}
        value={value}
        onChange={onChange}
        options={{
          readOnly,
          minimap: { enabled: false },
          fontSize: 14,
          wordWrap: 'on',
          automaticLayout: true,
        }}
      />
    </div>
  )
}

'use client'

import { useState, useRef, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { PaperAirplaneIcon } from '@heroicons/react/24/outline'
import { Card } from '@/components/ui/Card'
import Button from '@/components/ui/Button'
import Loading from '@/components/ui/Loading'
import { useSemanticRAG } from '@/lib/api/queries'
import { cn } from '@/lib/utils/cn'
import ReactMarkdown from 'react-markdown'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
}

export default function ChatInterface() {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [isStreaming, setIsStreaming] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const { mutate: ragSearch, isPending } = useSemanticRAG()

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!input.trim() || isPending || isStreaming) return

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      timestamp: new Date(),
    }

    setMessages((prev) => [...prev, userMessage])
    setInput('')
    setIsStreaming(true)

    // Simulate streaming response (in production, use actual streaming API)
    const assistantMessageId = (Date.now() + 1).toString()
    const assistantMessage: Message = {
      id: assistantMessageId,
      role: 'assistant',
      content: '',
      timestamp: new Date(),
    }

    setMessages((prev) => [...prev, assistantMessage])

    // Call RAG API
    ragSearch(
      {
        query: userMessage.content,
        context_limit: 5,
      },
      {
        onSuccess: (data) => {
          setMessages((prev) =>
            prev.map((msg) =>
              msg.id === assistantMessageId
                ? { ...msg, content: data.response || data.content || 'Response received' }
                : msg
            )
          )
          setIsStreaming(false)
        },
        onError: (error) => {
          setMessages((prev) =>
            prev.map((msg) =>
              msg.id === assistantMessageId
                ? { ...msg, content: 'Sorry, an error occurred. Please try again.' }
                : msg
            )
          )
          setIsStreaming(false)
        },
      }
    )
  }

  return (
    <Card className="h-full flex flex-col">
      <div className="flex-1 overflow-y-auto p-4 sm:p-5 space-y-3">
        {messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center text-muted-foreground">
            <h3 className="text-lg font-semibold mb-2">Start a conversation</h3>
            <p className="text-sm">Ask questions about your documents or data</p>
          </div>
        ) : (
          <AnimatePresence>
            {messages.map((message) => (
              <motion.div
                key={message.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                className={cn(
                  'flex gap-3',
                  message.role === 'user' ? 'justify-end' : 'justify-start'
                )}
              >
                {message.role === 'assistant' && (
                  <div className="flex-shrink-0 h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                    AI
                  </div>
                )}
                <div
                  className={cn(
                    'max-w-[80%] rounded-lg p-4',
                    message.role === 'user'
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-foreground'
                  )}
                >
                  {message.role === 'assistant' ? (
                    <div className="prose prose-sm dark:prose-invert max-w-none">
                      <ReactMarkdown>{message.content}</ReactMarkdown>
                    </div>
                  ) : (
                    <p className="text-sm whitespace-pre-wrap">{message.content}</p>
                  )}
                </div>
              </motion.div>
            ))}
          </AnimatePresence>
        )}
        {isPending && (
          <div className="flex gap-3">
            <div className="flex-shrink-0 h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
              <Loading size="sm" />
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <form onSubmit={handleSubmit} className="border-t border-border p-3 sm:p-4 flex-shrink-0">
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ask a question..."
            className={cn(
              'flex-1 rounded-lg border border-border bg-background px-4 py-2',
              'text-sm focus:outline-none focus:ring-2 focus:ring-ring'
            )}
            disabled={isPending || isStreaming}
          />
          <Button type="submit" disabled={!input.trim() || isPending || isStreaming}>
            <PaperAirplaneIcon className="h-5 w-5" />
          </Button>
        </div>
      </form>
    </Card>
  )
}

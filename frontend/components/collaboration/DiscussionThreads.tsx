'use client'

import { useState, useEffect } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import Input from '@/components/ui/Input'

interface Thread {
  id: string
  title: string
  status: string
  posts: ThreadPost[]
  created_at: string
}

interface ThreadPost {
  id: string
  author_id: string
  content: string
  created_at: string
}

interface DiscussionThreadsProps {
  resourceType: string
  resourceId: string
}

export default function DiscussionThreads({ resourceType, resourceId }: DiscussionThreadsProps) {
  const [threads, setThreads] = useState<Thread[]>([])
  const [selectedThread, setSelectedThread] = useState<Thread | null>(null)
  const [newPost, setNewPost] = useState('')
  const [newThreadTitle, setNewThreadTitle] = useState('')

  useEffect(() => {
    fetchThreads()
  }, [resourceType, resourceId])

  const fetchThreads = async () => {
    try {
      const response = await fetch(`/api/v1/collaboration/threads?resource_type=${resourceType}&resource_id=${resourceId}`)
      const data = await response.json()
      setThreads(data)
    } catch (error) {
      console.error('Failed to fetch threads:', error)
    }
  }

  const createThread = async () => {
    try {
      await fetch('/api/v1/collaboration/threads', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          resource_type: resourceType,
          resource_id: resourceId,
          title: newThreadTitle,
          initial_post: { content: newPost },
        }),
      })
      setNewThreadTitle('')
      setNewPost('')
      fetchThreads()
    } catch (error) {
      console.error('Failed to create thread:', error)
    }
  }

  return (
    <div className="grid grid-cols-2 gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Discussion Threads</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2 mb-4">
            <Input
              placeholder="Thread title"
              value={newThreadTitle}
              onChange={(e) => setNewThreadTitle(e.target.value)}
            />
            <Input
              placeholder="Initial post"
              value={newPost}
              onChange={(e) => setNewPost(e.target.value)}
            />
            <Button onClick={createThread} size="sm">Create Thread</Button>
          </div>
          <div className="space-y-2">
            {threads.map((thread) => (
              <div
                key={thread.id}
                className="p-3 border rounded cursor-pointer hover:bg-gray-50"
                onClick={() => setSelectedThread(thread)}
              >
                <div className="font-medium">{thread.title}</div>
                <div className="text-sm text-gray-500">{thread.status}</div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
      {selectedThread && (
        <Card>
          <CardHeader>
            <CardTitle>{selectedThread.title}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {selectedThread.posts.map((post) => (
                <div key={post.id} className="p-3 bg-gray-50 rounded">
                  <div className="text-sm font-medium">{post.author_id}</div>
                  <div>{post.content}</div>
                  <div className="text-xs text-gray-500">{new Date(post.created_at).toLocaleString()}</div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

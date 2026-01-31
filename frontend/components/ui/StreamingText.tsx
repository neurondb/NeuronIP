'use client'

import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'

interface StreamingTextProps {
  text: string
  speed?: number // Characters per second
  onComplete?: () => void
  className?: string
}

export default function StreamingText({
  text,
  speed = 30,
  onComplete,
  className = '',
}: StreamingTextProps) {
  const [displayedText, setDisplayedText] = useState('')
  const [isComplete, setIsComplete] = useState(false)

  useEffect(() => {
    if (!text) {
      setDisplayedText('')
      setIsComplete(false)
      return
    }

    setDisplayedText('')
    setIsComplete(false)

    let currentIndex = 0
    const interval = 1000 / speed // Time per character

    const timer = setInterval(() => {
      if (currentIndex < text.length) {
        setDisplayedText(text.slice(0, currentIndex + 1))
        currentIndex++
      } else {
        setIsComplete(true)
        clearInterval(timer)
        onComplete?.()
      }
    }, interval)

    return () => clearInterval(timer)
  }, [text, speed, onComplete])

  return (
    <span className={className}>
      {displayedText}
      {!isComplete && (
        <motion.span
          animate={{ opacity: [1, 0] }}
          transition={{ duration: 0.8, repeat: Infinity, repeatType: 'reverse' }}
          className="inline-block w-0.5 h-4 bg-current ml-0.5 align-middle"
        />
      )}
    </span>
  )
}

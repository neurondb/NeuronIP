'use client'

import { InformationCircleIcon } from '@heroicons/react/24/outline'
import { useState } from 'react'

import Tooltip from './Tooltip'

interface FeatureHintProps {
  text: string
  className?: string
}

export default function FeatureHint({ text, className = '' }: FeatureHintProps) {
  return (
    <Tooltip content={text} variant="info">
      <InformationCircleIcon className={`h-4 w-4 text-muted-foreground cursor-help ${className}`} />
    </Tooltip>
  )
}

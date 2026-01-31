'use client'

import { useState, useEffect } from 'react'

const ONBOARDING_STORAGE_KEY = 'neuronip_onboarding'
const ONBOARDING_COMPLETED_KEY = 'onboarding_completed'

export interface OnboardingState {
  isCompleted: boolean
  currentStep: string | null
  skipped: boolean
  data: Record<string, any>
}

export function useOnboarding() {
  const [state, setState] = useState<OnboardingState>({
    isCompleted: false,
    currentStep: null,
    skipped: false,
    data: {},
  })

  useEffect(() => {
    // Load from localStorage
    if (typeof window !== 'undefined') {
      const completed = localStorage.getItem(ONBOARDING_COMPLETED_KEY) === 'true'
      const stored = localStorage.getItem(ONBOARDING_STORAGE_KEY)
      
      setState({
        isCompleted: completed,
        currentStep: stored ? JSON.parse(stored).currentStep : null,
        skipped: stored ? JSON.parse(stored).skipped : false,
        data: stored ? JSON.parse(stored).data : {},
      })
    }
  }, [])

  const markCompleted = () => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(ONBOARDING_COMPLETED_KEY, 'true')
      localStorage.removeItem(ONBOARDING_STORAGE_KEY)
    }
    setState(prev => ({ ...prev, isCompleted: true, currentStep: null }))
  }

  const markSkipped = () => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(ONBOARDING_STORAGE_KEY, JSON.stringify({
        skipped: true,
        currentStep: null,
        data: {},
      }))
    }
    setState(prev => ({ ...prev, skipped: true, currentStep: null }))
  }

  const updateStep = (step: string | null) => {
    if (typeof window !== 'undefined') {
      const current = localStorage.getItem(ONBOARDING_STORAGE_KEY)
      const data = current ? JSON.parse(current) : {}
      localStorage.setItem(ONBOARDING_STORAGE_KEY, JSON.stringify({
        ...data,
        currentStep: step,
      }))
    }
    setState(prev => ({ ...prev, currentStep: step }))
  }

  const updateData = (newData: Record<string, any>) => {
    if (typeof window !== 'undefined') {
      const current = localStorage.getItem(ONBOARDING_STORAGE_KEY)
      const existing = current ? JSON.parse(current) : {}
      const updated = {
        ...existing,
        data: { ...existing.data, ...newData },
      }
      localStorage.setItem(ONBOARDING_STORAGE_KEY, JSON.stringify(updated))
    }
    setState(prev => ({
      ...prev,
      data: { ...prev.data, ...newData },
    }))
  }

  const reset = () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem(ONBOARDING_COMPLETED_KEY)
      localStorage.removeItem(ONBOARDING_STORAGE_KEY)
    }
    setState({
      isCompleted: false,
      currentStep: null,
      skipped: false,
      data: {},
    })
  }

  return {
    ...state,
    markCompleted,
    markSkipped,
    updateStep,
    updateData,
    reset,
  }
}

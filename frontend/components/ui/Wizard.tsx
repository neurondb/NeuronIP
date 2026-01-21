'use client'

import { useState, ReactNode, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './Card'
import Button from './Button'
import {
  ChevronLeftIcon,
  ChevronRightIcon,
  XMarkIcon,
  CheckCircleIcon,
} from '@heroicons/react/24/outline'
import { cn } from '@/lib/utils/cn'

export interface WizardStep {
  id: string
  title: string
  description?: string
  component: React.ComponentType<WizardStepProps>
  validate?: (data: any) => Promise<boolean> | boolean
  canSkip?: boolean
  isOptional?: boolean
}

export interface WizardStepProps {
  data: any
  updateData: (data: any) => void
  goToStep: (stepId: string) => void
  nextStep: () => void
  previousStep: () => void
  isFirstStep: boolean
  isLastStep: boolean
  currentStep: number
  totalSteps: number
}

export interface WizardProps {
  steps: WizardStep[]
  title?: string
  description?: string
  onComplete: (data: any) => void | Promise<void>
  onCancel?: () => void
  initialData?: any
  showProgress?: boolean
  className?: string
  allowSkip?: boolean
}

export default function Wizard({
  steps,
  title,
  description,
  onComplete,
  onCancel,
  initialData = {},
  showProgress = true,
  className,
  allowSkip = false,
}: WizardProps) {
  const [currentStepIndex, setCurrentStepIndex] = useState(0)
  const [data, setData] = useState<any>(initialData)
  const [isValidating, setIsValidating] = useState(false)
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({})
  const [completedSteps, setCompletedSteps] = useState<Set<string>>(new Set())
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)

  const currentStep = steps[currentStepIndex]
  const isFirstStep = currentStepIndex === 0
  const isLastStep = currentStepIndex === steps.length - 1
  const totalSteps = steps.length

  const updateData = useCallback((newData: any) => {
    setData((prev: any) => ({ ...prev, ...newData }))
    setHasUnsavedChanges(true)
  }, [])

  const goToStep = useCallback((stepIndex: number) => {
    if (stepIndex >= 0 && stepIndex < steps.length) {
      setCurrentStepIndex(stepIndex)
      setValidationErrors({})
    }
  }, [steps.length])

  const goToStepById = useCallback((stepId: string) => {
    const index = steps.findIndex((s) => s.id === stepId)
    if (index >= 0) {
      goToStep(index)
    }
  }, [steps, goToStep])

  const validateCurrentStep = useCallback(async (): Promise<boolean> => {
    if (!currentStep.validate) {
      return true
    }

    setIsValidating(true)
    setValidationErrors({})

    try {
      const isValid = await currentStep.validate(data)
      if (!isValid) {
        setValidationErrors({
          [currentStep.id]: 'Validation failed. Please check your input.',
        })
      }
      return isValid
    } catch (error) {
      setValidationErrors({
        [currentStep.id]: error instanceof Error ? error.message : 'Validation error occurred',
      })
      return false
    } finally {
      setIsValidating(false)
    }
  }, [currentStep, data])

  const nextStep = useCallback(async () => {
    const isValid = await validateCurrentStep()
    if (!isValid && !currentStep.canSkip) {
      return
    }

    if (isValid) {
      setCompletedSteps((prev) => new Set([...prev, currentStep.id]))
    }

    if (!isLastStep) {
      goToStep(currentStepIndex + 1)
    } else {
      // Complete wizard
      try {
        await onComplete(data)
        setHasUnsavedChanges(false)
      } catch (error) {
        console.error('Error completing wizard:', error)
      }
    }
  }, [currentStep, currentStepIndex, isLastStep, validateCurrentStep, goToStep, data, onComplete])

  const previousStep = useCallback(() => {
    if (!isFirstStep) {
      goToStep(currentStepIndex - 1)
    }
  }, [isFirstStep, currentStepIndex, goToStep])

  const handleCancel = useCallback(() => {
    if (hasUnsavedChanges) {
      const confirmed = window.confirm(
        'You have unsaved changes. Are you sure you want to cancel?'
      )
      if (!confirmed) {
        return
      }
    }
    if (onCancel) {
      onCancel()
    }
  }, [hasUnsavedChanges, onCancel])

  const stepProps: WizardStepProps = {
    data,
    updateData,
    goToStep: goToStepById,
    nextStep,
    previousStep,
    isFirstStep,
    isLastStep,
    currentStep: currentStepIndex + 1,
    totalSteps,
  }

  const StepComponent = currentStep.component

  return (
    <div className={cn('flex flex-col h-full', className)}>
      {/* Header */}
      {(title || description) && (
        <div className="mb-6">
          {title && <h2 className="text-2xl font-bold">{title}</h2>}
          {description && <p className="text-muted-foreground mt-1">{description}</p>}
        </div>
      )}

      {/* Progress Indicator */}
      {showProgress && totalSteps > 1 && (
        <div className="mb-6">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium">
              Step {currentStepIndex + 1} of {totalSteps}
            </span>
            <span className="text-sm text-muted-foreground">{currentStep.title}</span>
          </div>
          <div className="w-full bg-muted rounded-full h-2">
            <motion.div
              className="bg-primary h-2 rounded-full"
              initial={{ width: 0 }}
              animate={{ width: `${((currentStepIndex + 1) / totalSteps) * 100}%` }}
              transition={{ duration: 0.3 }}
            />
          </div>
          {/* Step indicators */}
          <div className="flex items-center justify-between mt-4">
            {steps.map((step, index) => (
              <div key={step.id} className="flex items-center flex-1">
                <button
                  type="button"
                  onClick={() => {
                    // Only allow going to completed steps or next step
                    if (completedSteps.has(step.id) || index === currentStepIndex) {
                      goToStep(index)
                    }
                  }}
                  className={cn(
                    'flex items-center justify-center w-8 h-8 rounded-full border-2 transition-colors',
                    index < currentStepIndex || completedSteps.has(step.id)
                      ? 'bg-primary border-primary text-primary-foreground'
                      : index === currentStepIndex
                      ? 'border-primary text-primary bg-background'
                      : 'border-muted text-muted-foreground bg-background',
                    (completedSteps.has(step.id) || index === currentStepIndex) &&
                      'cursor-pointer hover:border-primary/80'
                  )}
                  disabled={!completedSteps.has(step.id) && index !== currentStepIndex}
                >
                  {index < currentStepIndex || completedSteps.has(step.id) ? (
                    <CheckCircleIcon className="h-5 w-5" />
                  ) : (
                    <span className="text-sm font-medium">{index + 1}</span>
                  )}
                </button>
                {index < steps.length - 1 && (
                  <div
                    className={cn(
                      'flex-1 h-0.5 mx-2',
                      index < currentStepIndex ? 'bg-primary' : 'bg-muted'
                    )}
                  />
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Step Content */}
      <Card className="flex-1 flex flex-col min-h-0">
        <CardHeader className="flex-shrink-0">
          <div className="flex items-start justify-between">
            <div>
              <CardTitle>{currentStep.title}</CardTitle>
              {currentStep.description && (
                <CardDescription className="mt-1">{currentStep.description}</CardDescription>
              )}
            </div>
            {onCancel && (
              <button
                type="button"
                onClick={handleCancel}
                className="text-muted-foreground hover:text-foreground transition-colors"
                aria-label="Close"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            )}
          </div>
        </CardHeader>
        <CardContent className="flex-1 min-h-0 overflow-y-auto">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentStep.id}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
            >
              {validationErrors[currentStep.id] && (
                <div className="mb-4 p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
                  {validationErrors[currentStep.id]}
                </div>
              )}
              <StepComponent {...stepProps} />
            </motion.div>
          </AnimatePresence>
        </CardContent>

        {/* Navigation */}
        <div className="flex items-center justify-between p-6 border-t border-border">
          <div>
            {!isFirstStep && (
              <Button onClick={previousStep} variant="outline" disabled={isValidating}>
                <ChevronLeftIcon className="h-4 w-4 mr-2" />
                Previous
              </Button>
            )}
          </div>
          <div className="flex gap-2">
            {onCancel && (
              <Button onClick={handleCancel} variant="ghost" disabled={isValidating}>
                Cancel
              </Button>
            )}
            <Button
              onClick={nextStep}
              disabled={isValidating}
              isLoading={isValidating}
            >
              {isLastStep ? 'Complete' : 'Next'}
              {!isLastStep && <ChevronRightIcon className="h-4 w-4 ml-2" />}
            </Button>
          </div>
        </div>
      </Card>
    </div>
  )
}

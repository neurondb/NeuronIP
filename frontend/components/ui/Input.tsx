'use client'

import { cva, type VariantProps } from 'class-variance-authority'
import { forwardRef, InputHTMLAttributes } from 'react'

import { cn } from '@/lib/utils/cn'

const inputVariants = cva(
  'flex h-10 w-full rounded-lg border bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-colors duration-fast',
  {
    variants: {
      variant: {
        default: 'border-input focus-visible:ring-ring',
        error: 'border-destructive focus-visible:ring-destructive',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

interface InputProps extends InputHTMLAttributes<HTMLInputElement>, VariantProps<typeof inputVariants> {
  label?: string
  error?: string
  helperText?: string
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, helperText, variant, ...props }, ref) => {
    const effectiveVariant = variant ?? (error ? 'error' : 'default')
    return (
      <div className="w-full">
        {label && (
          <label className="block text-sm font-medium text-foreground mb-1.5">
            {label}
            {props.required && <span className="text-destructive ml-1">*</span>}
          </label>
        )}
        <input
          ref={ref}
          className={cn(inputVariants({ variant: effectiveVariant }), className)}
          aria-invalid={!!error}
          aria-describedby={error ? 'input-error' : helperText ? 'input-helper' : undefined}
          {...props}
        />
        {error && (
          <p id="input-error" className="mt-1.5 text-sm text-destructive" role="alert">
            {error}
          </p>
        )}
        {helperText && !error && (
          <p id="input-helper" className="mt-1.5 text-sm text-muted-foreground">
            {helperText}
          </p>
        )}
      </div>
    )
  }
)

Input.displayName = 'Input'

export { inputVariants }
export default Input

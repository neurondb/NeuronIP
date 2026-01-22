'use client'

import * as React from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from './Form'
import Input from './Input'
import { Button } from './Button'
import { Textarea } from './Textarea'
import Select from './Select'
import { Checkbox } from './Checkbox'
import { RadioGroup, RadioGroupItem } from './RadioGroup'
import { Switch } from './Switch'
import { DatePicker } from './DatePicker'
import { MultiSelect } from './MultiSelect'

export type FieldType =
  | 'text'
  | 'email'
  | 'password'
  | 'number'
  | 'textarea'
  | 'select'
  | 'checkbox'
  | 'radio'
  | 'switch'
  | 'date'
  | 'multiselect'

export interface FormFieldConfig {
  name: string
  label: string
  type: FieldType
  placeholder?: string
  description?: string
  required?: boolean
  defaultValue?: any
  options?: { label: string; value: string }[]
  validation?: z.ZodTypeAny
}

interface FormBuilderProps {
  fields: FormFieldConfig[]
  onSubmit: (data: Record<string, any>) => void | Promise<void>
  submitLabel?: string
  className?: string
}

export function FormBuilder({ fields, onSubmit, submitLabel = 'Submit', className }: FormBuilderProps) {
  const schema = z.object(
    fields.reduce((acc, field) => {
      let fieldSchema: z.ZodTypeAny
      
      if (field.validation) {
        fieldSchema = field.validation
      } else if (field.type === 'email') {
        fieldSchema = field.required 
          ? z.string().email('Invalid email address').min(1, `${field.label} is required`)
          : z.string().email('Invalid email address').optional()
      } else if (field.type === 'number') {
        fieldSchema = field.required ? z.number() : z.number().optional()
      } else if (field.type === 'checkbox' || field.type === 'switch') {
        fieldSchema = field.required ? z.boolean() : z.boolean().optional()
      } else if (field.type === 'multiselect') {
        fieldSchema = field.required
          ? z.array(z.string()).min(1, `${field.label} is required`)
          : z.array(z.string()).optional()
      } else {
        fieldSchema = field.required
          ? z.string().min(1, `${field.label} is required`)
          : z.string().optional()
      }
      
      acc[field.name] = fieldSchema
      return acc
    }, {} as Record<string, z.ZodTypeAny>)
  )

  const form = useForm({
    resolver: zodResolver(schema),
    defaultValues: fields.reduce((acc, field) => {
      acc[field.name] = field.defaultValue ?? (field.type === 'checkbox' || field.type === 'switch' ? false : field.type === 'multiselect' ? [] : '')
      return acc
    }, {} as Record<string, any>),
  })

  const handleSubmit = form.handleSubmit(async (data) => {
    await onSubmit(data)
  })

  const renderField = (field: FormFieldConfig) => {
    return (
      <FormField
        key={field.name}
        control={form.control}
        name={field.name}
        render={({ field: formField }) => {
          switch (field.type) {
            case 'textarea':
              return (
                <FormItem>
                  <FormLabel>{field.label}</FormLabel>
                  <FormControl>
                    <Textarea {...formField} placeholder={field.placeholder} />
                  </FormControl>
                  {field.description && <FormDescription>{field.description}</FormDescription>}
                  <FormMessage />
                </FormItem>
              )
            case 'select':
              return (
                <FormItem>
                  <FormLabel>{field.label}</FormLabel>
                  <FormControl>
                    <Select {...formField} options={field.options || []} />
                  </FormControl>
                  {field.description && <FormDescription>{field.description}</FormDescription>}
                  <FormMessage />
                </FormItem>
              )
            case 'checkbox':
              return (
                <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                  <FormControl>
                    <Checkbox checked={formField.value} onCheckedChange={formField.onChange} />
                  </FormControl>
                  <div className="space-y-1 leading-none">
                    <FormLabel>{field.label}</FormLabel>
                    {field.description && <FormDescription>{field.description}</FormDescription>}
                  </div>
                  <FormMessage />
                </FormItem>
              )
            case 'radio':
              return (
                <FormItem>
                  <FormLabel>{field.label}</FormLabel>
                  <FormControl>
                    <RadioGroup value={formField.value} onValueChange={formField.onChange}>
                      {field.options?.map((option) => (
                        <div key={option.value} className="flex items-center space-x-2">
                          <RadioGroupItem value={option.value} id={option.value} />
                          <label htmlFor={option.value}>{option.label}</label>
                        </div>
                      ))}
                    </RadioGroup>
                  </FormControl>
                  {field.description && <FormDescription>{field.description}</FormDescription>}
                  <FormMessage />
                </FormItem>
              )
            case 'switch':
              return (
                <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                  <div className="space-y-0.5">
                    <FormLabel>{field.label}</FormLabel>
                    {field.description && <FormDescription>{field.description}</FormDescription>}
                  </div>
                  <FormControl>
                    <Switch checked={formField.value} onCheckedChange={formField.onChange} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )
            case 'date':
              return (
                <FormItem>
                  <FormLabel>{field.label}</FormLabel>
                  <FormControl>
                    <DatePicker date={formField.value} onDateChange={formField.onChange} />
                  </FormControl>
                  {field.description && <FormDescription>{field.description}</FormDescription>}
                  <FormMessage />
                </FormItem>
              )
            case 'multiselect':
              return (
                <FormItem>
                  <FormLabel>{field.label}</FormLabel>
                  <FormControl>
                    <MultiSelect
                      options={field.options || []}
                      selected={formField.value || []}
                      onChange={formField.onChange}
                    />
                  </FormControl>
                  {field.description && <FormDescription>{field.description}</FormDescription>}
                  <FormMessage />
                </FormItem>
              )
            default:
              return (
                <FormItem>
                  <FormLabel>{field.label}</FormLabel>
                  <FormControl>
                    <Input
                      {...formField}
                      type={field.type}
                      placeholder={field.placeholder}
                    />
                  </FormControl>
                  {field.description && <FormDescription>{field.description}</FormDescription>}
                  <FormMessage />
                </FormItem>
              )
          }
        }}
      />
    )
  }

  return (
    <Form {...form}>
      <form onSubmit={handleSubmit} className={className}>
        <div className="space-y-4">
          {fields.map((field) => renderField(field))}
        </div>
        <Button type="submit" className="mt-6">
          {submitLabel}
        </Button>
      </form>
    </Form>
  )
}

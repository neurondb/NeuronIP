# Wizard Development Guide

This guide covers how to create wizards, wizard component patterns, and best practices for multi-step flows in NeuronIP.

## Overview

Wizards provide guided, step-by-step experiences for complex operations. NeuronIP includes a reusable `Wizard` component that handles navigation, validation, and progress tracking.

## Base Wizard Component

### Importing

```typescript
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
```

### Basic Usage

```typescript
function MyWizard() {
  const steps: WizardStep[] = [
    {
      id: 'step1',
      title: 'Step 1',
      description: 'First step description',
      component: Step1Component,
    },
    {
      id: 'step2',
      title: 'Step 2',
      component: Step2Component,
    },
  ]

  const handleComplete = async (data: any) => {
    // Process completed wizard data
    console.log('Wizard completed:', data)
  }

  return (
    <Wizard
      steps={steps}
      title="My Wizard"
      description="Complete this wizard step by step"
      onComplete={handleComplete}
      onCancel={() => console.log('Cancelled')}
      showProgress={true}
    />
  )
}
```

## Wizard Step Structure

### Step Definition

```typescript
interface WizardStep {
  id: string                    // Unique step identifier
  title: string                 // Step title displayed in header
  description?: string          // Optional step description
  component: React.ComponentType<WizardStepProps>  // Step component
  validate?: (data: any) => Promise<boolean> | boolean  // Optional validation
  canSkip?: boolean             // Whether step can be skipped
  isOptional?: boolean          // Whether step is optional
}
```

### Step Component Props

```typescript
interface WizardStepProps {
  data: any                     // Current wizard data
  updateData: (data: any) => void  // Update wizard data
  goToStep: (stepId: string) => void  // Navigate to specific step
  nextStep: () => void          // Go to next step
  previousStep: () => void      // Go to previous step
  isFirstStep: boolean          // Whether this is the first step
  isLastStep: boolean           // Whether this is the last step
  currentStep: number           // Current step number (1-indexed)
  totalSteps: number            // Total number of steps
}
```

## Creating a Step Component

### Basic Step Component

```typescript
function MyStep({ data, updateData }: WizardStepProps) {
  const [value, setValue] = useState(data.value || '')

  return (
    <div className="space-y-4">
      <label className="block text-sm font-medium">Value</label>
      <input
        value={value}
        onChange={(e) => {
          setValue(e.target.value)
          updateData({ value: e.target.value })
        }}
        className="w-full px-3 py-2 border rounded-lg"
      />
    </div>
  )
}
```

### Step with Validation

```typescript
function MyStep({ data, updateData }: WizardStepProps) {
  const [value, setValue] = useState(data.value || '')
  const [error, setError] = useState('')

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value
    setValue(newValue)
    
    // Client-side validation
    if (newValue.length < 3) {
      setError('Value must be at least 3 characters')
    } else {
      setError('')
      updateData({ value: newValue })
    }
  }

  return (
    <div className="space-y-4">
      <label className="block text-sm font-medium">Value</label>
      <input
        value={value}
        onChange={handleChange}
        className="w-full px-3 py-2 border rounded-lg"
      />
      {error && <p className="text-sm text-red-600">{error}</p>}
    </div>
  )
}
```

## Complete Wizard Example

### Example: User Creation Wizard

```typescript
import { useState } from 'react'
import Wizard, { WizardStep, WizardStepProps } from '@/components/ui/Wizard'
import { Input } from '@/components/ui/Input'
import { useCreateUser } from '@/lib/api/queries'

interface UserWizardData {
  name: string
  email: string
  role: string
}

// Step 1: Basic Information
function BasicInfoStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as UserWizardData) || { name: '', email: '' }

  return (
    <div className="space-y-4">
      <Input
        label="Name *"
        value={wizardData.name || ''}
        onChange={(e) => updateData({ name: e.target.value })}
        required
      />
      <Input
        label="Email *"
        type="email"
        value={wizardData.email || ''}
        onChange={(e) => updateData({ email: e.target.value })}
        required
      />
    </div>
  )
}

// Step 2: Role Selection
function RoleStep({ data, updateData }: WizardStepProps) {
  const wizardData = (data as UserWizardData) || { role: '' }
  const roles = ['admin', 'user', 'viewer']

  return (
    <div className="space-y-4">
      <label className="block text-sm font-medium">Role *</label>
      {roles.map((role) => (
        <label key={role} className="flex items-center gap-2">
          <input
            type="radio"
            name="role"
            value={role}
            checked={wizardData.role === role}
            onChange={(e) => updateData({ role: e.target.value })}
          />
          <span className="capitalize">{role}</span>
        </label>
      ))}
    </div>
  )
}

// Step 3: Review
function ReviewStep({ data }: WizardStepProps) {
  const wizardData = data as UserWizardData

  return (
    <div className="space-y-4">
      <h3 className="font-semibold">Review Information</h3>
      <div>
        <strong>Name:</strong> {wizardData.name}
      </div>
      <div>
        <strong>Email:</strong> {wizardData.email}
      </div>
      <div>
        <strong>Role:</strong> {wizardData.role}
      </div>
    </div>
  )
}

// Main Wizard Component
export default function CreateUserWizard() {
  const { mutate: createUser } = useCreateUser()

  const handleComplete = async (data: UserWizardData) => {
    createUser(data, {
      onSuccess: () => {
        showToast('User created successfully', 'success')
      },
      onError: (error: any) => {
        showToast(error?.message || 'Failed to create user', 'error')
      },
    })
  }

  const validateBasicInfo = (data: any): boolean => {
    const wizardData = data as UserWizardData
    return !!(wizardData.name?.trim() && wizardData.email?.trim())
  }

  const validateRole = (data: any): boolean => {
    const wizardData = data as UserWizardData
    return !!wizardData.role
  }

  const steps: WizardStep[] = [
    {
      id: 'basic',
      title: 'Basic Information',
      description: 'Enter name and email',
      component: BasicInfoStep,
      validate: validateBasicInfo,
    },
    {
      id: 'role',
      title: 'Select Role',
      description: 'Choose user role',
      component: RoleStep,
      validate: validateRole,
    },
    {
      id: 'review',
      title: 'Review & Create',
      description: 'Review information before creating',
      component: ReviewStep,
    },
  ]

  return (
    <Wizard
      steps={steps}
      title="Create User"
      description="Set up a new user account"
      onComplete={handleComplete}
      showProgress={true}
    />
  )
}
```

## Wizard Patterns

### Pattern 1: Progressive Disclosure

Break complex operations into logical steps:

```typescript
const steps = [
  { id: 'basic', title: 'Basic Info' },
  { id: 'advanced', title: 'Advanced Settings' },
  { id: 'review', title: 'Review' },
]
```

### Pattern 2: Conditional Steps

Use `canSkip` and conditional logic:

```typescript
const steps = [
  { id: 'basic', component: BasicStep },
  {
    id: 'advanced',
    component: AdvancedStep,
    canSkip: true,  // User can skip if not needed
  },
]
```

### Pattern 3: Template Selection

Let users choose a starting point:

```typescript
function TemplateStep({ data, updateData }: WizardStepProps) {
  const templates = ['basic', 'advanced', 'custom']
  
  return (
    <div className="grid grid-cols-3 gap-4">
      {templates.map((template) => (
        <Card
          key={template}
          onClick={() => updateData({ template })}
          className={data.template === template ? 'ring-2 ring-primary' : ''}
        >
          {template}
        </Card>
      ))}
    </div>
  )
}
```

### Pattern 4: Dynamic Step Generation

Generate steps based on data:

```typescript
function getSteps(data: any): WizardStep[] {
  const baseSteps = [
    { id: 'start', component: StartStep },
  ]
  
  if (data.type === 'advanced') {
    baseSteps.push({
      id: 'advanced',
      component: AdvancedStep,
    })
  }
  
  baseSteps.push({
    id: 'review',
    component: ReviewStep,
  })
  
  return baseSteps
}
```

## Best Practices

### 1. Clear Step Titles and Descriptions

```typescript
// Good
{
  id: 'credentials',
  title: 'Provide Credentials',
  description: 'Enter your API key and secret. These will be securely stored.',
  component: CredentialsStep,
}

// Bad
{
  id: 'step2',
  title: 'Step 2',
  component: Step2Component,
}
```

### 2. Validate Before Proceeding

```typescript
const validateEmail = (data: any): boolean => {
  const email = data.email || ''
  return email.includes('@') && email.includes('.')
}

const steps = [
  {
    id: 'email',
    component: EmailStep,
    validate: validateEmail,  // Prevents proceeding with invalid data
  },
]
```

### 3. Provide Clear Feedback

```typescript
function TestConnectionStep({ data }: WizardStepProps) {
  const [status, setStatus] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')

  const handleTest = async () => {
    setStatus('testing')
    try {
      await testConnection(data)
      setStatus('success')
    } catch (error) {
      setStatus('error')
    }
  }

  return (
    <div>
      <Button onClick={handleTest} disabled={status === 'testing'}>
        Test Connection
      </Button>
      {status === 'success' && <p className="text-green-600">✓ Connected</p>}
      {status === 'error' && <p className="text-red-600">✗ Connection failed</p>}
    </div>
  )
}
```

### 4. Save Progress (Optional)

```typescript
const handleComplete = async (data: any) => {
  // Save to local storage as draft
  localStorage.setItem('wizard_draft', JSON.stringify(data))
  
  // Complete wizard
  await submitData(data)
  localStorage.removeItem('wizard_draft')
}
```

### 5. Use Review Step

Always include a review step before completion:

```typescript
function ReviewStep({ data }: WizardStepProps) {
  return (
    <Card>
      <CardContent>
        <h3>Review Your Configuration</h3>
        {/* Display summary of all entered data */}
      </CardContent>
    </Card>
  )
}
```

### 6. Handle Errors Gracefully

```typescript
const handleComplete = async (data: any) => {
  try {
    await createResource(data)
    showToast('Created successfully', 'success')
  } catch (error) {
    const formatted = formatError(error)
    showToast(formatted.message, 'error')
    // Don't close wizard on error - let user fix and retry
  }
}
```

### 7. Make Steps Optional When Appropriate

```typescript
const steps = [
  {
    id: 'required',
    component: RequiredStep,
    validate: validateRequired,
  },
  {
    id: 'optional',
    component: OptionalStep,
    canSkip: true,  // User can skip
    isOptional: true,
  },
]
```

## Integration with Pages

### Modal Integration

```typescript
import Modal from '@/components/ui/Modal'
import MyWizard from '@/components/wizards/MyWizard'

function MyPage() {
  const [showWizard, setShowWizard] = useState(false)

  return (
    <>
      <Button onClick={() => setShowWizard(true)}>Open Wizard</Button>
      
      <Modal
        open={showWizard}
        onOpenChange={setShowWizard}
        size="xl"
        title="Wizard Title"
      >
        <MyWizard
          onComplete={() => setShowWizard(false)}
          onCancel={() => setShowWizard(false)}
        />
      </Modal>
    </>
  )
}
```

### Full Page Integration

```typescript
function MyPage() {
  return (
    <div className="container mx-auto p-6">
      <MyWizard />
    </div>
  )
}
```

## Styling

Wizards use Tailwind CSS classes. Customize as needed:

```typescript
<Wizard
  steps={steps}
  className="max-w-4xl mx-auto"  // Custom container styling
  showProgress={true}
/>
```

## Accessibility

The Wizard component includes:
- Keyboard navigation (Tab, Enter, Escape)
- ARIA labels for screen readers
- Focus management
- Progress announcements

Ensure step components are accessible:
- Use semantic HTML
- Add labels to form inputs
- Provide error messages
- Use ARIA attributes when needed

## Testing

### Unit Testing Steps

```typescript
describe('EmailStep', () => {
  it('updates wizard data on input change', () => {
    const updateData = jest.fn()
    render(<EmailStep data={{}} updateData={updateData} />)
    
    const input = screen.getByLabelText('Email')
    fireEvent.change(input, { target: { value: 'test@example.com' } })
    
    expect(updateData).toHaveBeenCalledWith({ email: 'test@example.com' })
  })
})
```

### Integration Testing

```typescript
describe('CreateUserWizard', () => {
  it('completes wizard flow', async () => {
    render(<CreateUserWizard />)
    
    // Step 1
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'John' } })
    fireEvent.click(screen.getByText('Next'))
    
    // Step 2
    fireEvent.click(screen.getByLabelText('Admin'))
    fireEvent.click(screen.getByText('Next'))
    
    // Complete
    fireEvent.click(screen.getByText('Complete'))
    
    await waitFor(() => {
      expect(mockCreateUser).toHaveBeenCalled()
    })
  })
})
```

## Examples

See existing wizard implementations:
- `WorkflowCreationWizard` - Complex multi-step workflow setup
- `IntegrationSetupWizard` - Integration configuration
- `AgentCreationWizard` - AI agent setup
- `OnboardingWizard` - First-time user onboarding

## Related Documentation

- [Validation Guide](./validation.md)
- [Error Handling Guide](./error-handling.md)
- [Request Flow Guide](../architecture/request-flow.md)

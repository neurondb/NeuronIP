# NeuronIP Frontend

A feature-rich, production-ready frontend built with Next.js 14, React 18, TypeScript, and modern tooling.

## Features

### 🎨 Advanced UI Components
- 25+ production-ready components built with Radix UI
- Fully accessible with ARIA support
- Dark mode support
- Responsive design
- TypeScript typed

### ⚡ Real-Time Features
- WebSocket client with auto-reconnection
- Real-time data synchronization
- Live notifications
- Presence indicators

### 📊 Data Visualization
- Interactive charts with zoom/pan
- Multiple chart types (Bar, Line, Pie, Sankey, Treemap)
- Export functionality (PNG, SVG, CSV)
- Chart configuration UI

### 🚀 Performance
- Code splitting and lazy loading
- PWA support with Service Worker
- Optimized bundle sizes
- Memoization utilities

### 🧪 Testing
- Vitest for unit testing
- React Testing Library for components
- Playwright for E2E testing
- MSW for API mocking

### 🌍 Internationalization
- next-intl integration
- Language switcher
- RTL support ready

### ♿ Accessibility
- WCAG AA compliant
- Keyboard navigation
- Screen reader support
- Focus management

## Getting Started

### Prerequisites
- Node.js 18+ 
- npm or yarn

### Installation

```bash
# Install dependencies
npm install

# Set up git hooks
npm run prepare

# Set up environment variables
cp .env.example .env.local
# Edit .env.local with your configuration
```

### Development

```bash
# Start development server
npm run dev

# Run tests
npm run test

# Run E2E tests
npm run test:e2e

# Start Storybook
npm run storybook

# Lint and format
npm run lint
npm run format
```

### Building

```bash
# Build for production
npm run build

# Start production server
npm start
```

## Project Structure

```
frontend/
├── app/                    # Next.js app directory
├── components/            # React components
│   ├── ui/               # UI component library
│   ├── charts/           # Chart components
│   ├── layout/           # Layout components
│   └── ...
├── lib/                  # Utilities and helpers
│   ├── hooks/            # Custom React hooks
│   ├── utils/            # Utility functions
│   ├── websocket/        # WebSocket client
│   └── ...
├── __tests__/            # Test files
├── .storybook/           # Storybook configuration
└── public/               # Static assets
```

## Key Technologies

- **Framework**: Next.js 14 (App Router)
- **UI Library**: React 18
- **Styling**: Tailwind CSS
- **UI Components**: Radix UI
- **State Management**: Zustand, TanStack Query
- **Forms**: React Hook Form + Zod
- **Charts**: Recharts, Nivo, D3
- **Testing**: Vitest, Playwright, React Testing Library
- **Real-time**: WebSocket (socket.io-client)
- **i18n**: next-intl
- **PWA**: next-pwa

## Component Usage

### Basic Example

```tsx
import { Button } from '@/components/ui'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui'

export function MyComponent() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Hello World</CardTitle>
      </CardHeader>
      <CardContent>
        <Button>Click me</Button>
      </CardContent>
    </Card>
  )
}
```

### Form with Validation

```tsx
import { FormBuilder, type FormFieldConfig } from '@/components/ui'

const fields: FormFieldConfig[] = [
  {
    name: 'email',
    label: 'Email',
    type: 'email',
    required: true,
  },
  {
    name: 'password',
    label: 'Password',
    type: 'password',
    required: true,
  },
]

export function MyForm() {
  return (
    <FormBuilder
      fields={fields}
      onSubmit={(data) => console.log(data)}
    />
  )
}
```

### Real-Time Data

```tsx
import { useRealtimeQuery } from '@/lib/hooks'

export function LiveData() {
  const { data } = useRealtimeQuery(
    ['live-data'],
    'data-update',
    { value: 0 }
  )

  return <div>Value: {data?.value}</div>
}
```

## Environment Variables

See `lib/utils/env.ts` for required variables:

- `NEXT_PUBLIC_API_URL` - API endpoint URL
- `NEXT_PUBLIC_WS_URL` - WebSocket URL (optional)
- `NEXT_PUBLIC_ENABLE_ANALYTICS` - Enable analytics (true/false)
- `NEXT_PUBLIC_ENABLE_SENTRY` - Enable Sentry (true/false)
- `NEXT_PUBLIC_SENTRY_DSN` - Sentry DSN (if enabled)

## Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint
- `npm run lint:fix` - Fix ESLint errors
- `npm run format` - Format code with Prettier
- `npm run format:check` - Check code formatting
- `npm run type-check` - Run TypeScript type checking
- `npm run test` - Run unit tests
- `npm run test:ui` - Run tests with UI
- `npm run test:coverage` - Run tests with coverage
- `npm run test:e2e` - Run E2E tests
- `npm run test:e2e:ui` - Run E2E tests with UI
- `npm run storybook` - Start Storybook
- `npm run build-storybook` - Build Storybook

## Contributing

1. Follow the code style (ESLint + Prettier)
2. Write tests for new features
3. Update documentation
4. Follow conventional commits
5. Ensure accessibility compliance

## License

See LICENSE file in project root.

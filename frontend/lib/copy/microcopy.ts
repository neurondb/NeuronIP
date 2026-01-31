/**
 * Centralized microcopy library for friendly, conversational language
 * Replaces technical jargon with human-centered messaging
 */

export const microcopy = {
  // Navigation & Actions
  actions: {
    create: 'Create',
    save: 'Save',
    cancel: 'Cancel',
    delete: 'Delete',
    edit: 'Edit',
    share: 'Share',
    copy: 'Copy',
    run: 'Run',
    execute: 'Run this',
    search: 'Search',
    filter: 'Filter',
    clear: 'Clear',
    close: 'Close',
    back: 'Back',
    next: 'Next',
    skip: 'Skip for now',
    continue: 'Continue',
    getStarted: "Let's go",
    tryIt: 'Try it now',
    explore: 'Explore',
    learnMore: 'Learn more',
    seeItInAction: 'See it in action',
  },

  // Common phrases
  phrases: {
    welcome: "Welcome! Let's get you started",
    youCanSkip: "You can skip this and explore first",
    optional: 'Optional',
    required: 'Required',
    comingSoon: 'Coming soon',
    inProgress: 'Working on it...',
    almostDone: 'Almost there!',
    allSet: "You're all set!",
    tryAgain: 'Try again',
    somethingWentWrong: "Oops! Something went wrong",
    needHelp: 'Need help?',
    getHelp: 'Get help',
  },

  // Onboarding
  onboarding: {
    welcome: {
      title: "Welcome to NeuronIP!",
      subtitle: "Your AI-powered data workspace",
      description: "Let's get you started. You can explore first, or set up a few things to get the most out of NeuronIP.",
      skipAndExplore: "Skip setup, explore first",
      startSetup: "Start setup",
      whatYoullSetUp: "What you'll set up:",
      apiKey: "Get your API key",
      dataSource: "Connect your data (optional)",
      semanticSearch: "Set up search (optional)",
      exploreFeatures: "Explore what's possible",
    },
    apiKey: {
      title: "Get your API key",
      description: "This lets you connect NeuronIP to your apps and tools",
      nameLabel: "What should we call this key?",
      namePlaceholder: "e.g., My first key",
      createButton: "Get my API key",
      created: "Here's your API key!",
      saveNotice: "Save this now - we'll only show it once",
      copyButton: "Copy to clipboard",
      copied: "Copied!",
    },
    dataSource: {
      title: "Connect your data",
      description: "Link your databases to ask questions and get insights",
      optional: "You can always do this later",
      connectButton: "Connect data source",
      skipButton: "Skip for now",
    },
    semanticSearch: {
      title: "Set up semantic search",
      description: "Search your documents by meaning, not just keywords",
      optional: "You can always do this later",
      collectionNameLabel: "What should we call this collection?",
      collectionNamePlaceholder: "e.g., My knowledge base",
      createButton: "Create collection",
      skipButton: "Skip for now",
    },
    features: {
      title: "Explore what's possible",
      description: "Here are some things you can do with NeuronIP",
      semanticSearch: {
        title: "Semantic Search",
        description: "Find information by meaning, not just keywords",
      },
      warehouse: {
        title: "Ask Questions",
        description: "Get answers about your data in plain English",
      },
      agents: {
        title: "AI Agents",
        description: "Build workflows that think for themselves",
      },
    },
    complete: {
      title: "You're all set!",
      description: "Ready to start? Head to your dashboard to explore.",
      goToDashboard: "Go to dashboard",
    },
  },

  // Dashboard
  dashboard: {
    title: "Dashboard",
    subtitle: "Overview of your NeuronIP workspace",
    quickStart: {
      title: "Quick Start",
      subtitle: "Try these to get started",
      searchData: "Search your data",
      askQuestion: "Ask a question",
      exploreFeatures: "Explore features",
    },
    quickActions: {
      title: "Quick Actions",
      semanticSearch: "Search",
      warehouseQuery: "Ask a question",
      runWorkflow: "Run workflow",
      complianceCheck: "Check compliance",
    },
  },

  // Warehouse / Query
  warehouse: {
    title: "Ask questions about your data",
    subtitle: "Type your question in plain English, or write SQL if you prefer",
    queryTab: "Ask",
    historyTab: "History",
    queryPlaceholder: "Ask a question about your data...",
    sqlPlaceholder: "Or write SQL here...",
    runQuery: "Run",
    running: "Running...",
    noResults: "No results found",
    error: "Something went wrong. Try rephrasing your question.",
    history: {
      title: "Recent questions",
      empty: "No recent questions yet",
    },
  },

  // Semantic Search
  semantic: {
    title: "Semantic Search",
    subtitle: "Find information by meaning, not just keywords",
    chatTitle: "Ask anything",
    chatPlaceholder: "Ask a question about your documents...",
    startConversation: "Start a conversation",
    askQuestions: "Ask questions about your documents or data",
    tryAsking: "Try asking:",
    example1: "What documents discuss customer support?",
    example2: "Explain the refund policy",
    example3: "Find information about API authentication",
    noDocuments: "No documents yet",
    addDocuments: "Add documents to get started",
  },

  // Common UI elements
  ui: {
    loading: "Loading...",
    saving: "Saving...",
    deleting: "Deleting...",
    searching: "Searching...",
    noResults: "No results found",
    emptyState: "Nothing here yet",
    error: "Something went wrong",
    success: "Done!",
    cancel: "Cancel",
    confirm: "Confirm",
    close: "Close",
  },

  // Settings & Configuration
  settings: {
    title: "Settings",
    preferences: "Preferences",
    account: "Account",
    security: "Security",
    notifications: "Notifications",
    appearance: "Appearance",
  },

  // Sharing
  sharing: {
    share: "Share",
    shareLink: "Share link",
    copyLink: "Copy link",
    linkCopied: "Link copied!",
    permissions: "Who can access",
    viewOnly: "View only",
    canEdit: "Can edit",
    fullAccess: "Full access",
    expires: "Expires",
    never: "Never",
  },

  // Collaboration
  collaboration: {
    comment: "Add a comment",
    reply: "Reply",
    mention: "Mention someone",
    react: "React",
    reactions: "Reactions",
    typing: "is typing...",
    viewing: "people viewing",
  },

  // Errors
  errors: {
    generic: "Something went wrong. Please try again.",
    notFound: "We couldn't find what you're looking for.",
    unauthorized: "You don't have permission to do that.",
    network: "Connection problem. Check your internet and try again.",
    validation: "Please check your input and try again.",
  },
} as const

/**
 * Get microcopy by path
 * Example: getCopy('onboarding.welcome.title') => "Welcome to NeuronIP!"
 */
export function getCopy(path: string): string {
  const keys = path.split('.')
  let value: any = microcopy
  
  for (const key of keys) {
    if (value && typeof value === 'object' && key in value) {
      value = value[key as keyof typeof value]
    } else {
      return path // Return path if not found (for debugging)
    }
  }
  
  return typeof value === 'string' ? value : path
}

/**
 * Get microcopy with fallback
 */
export function getCopyWithFallback(path: string, fallback: string): string {
  const copy = getCopy(path)
  return copy === path ? fallback : copy
}

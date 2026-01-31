'use client'

import {
  KeyIcon,
  ServerIcon,
  CheckCircleIcon,
  XCircleIcon,
  ArrowPathIcon,
  SparklesIcon,
} from '@heroicons/react/24/outline'
import { motion, AnimatePresence } from 'framer-motion'
import { useRouter } from 'next/navigation'
import { useState, useEffect } from 'react'

import Button from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import Input from '@/components/ui/Input'
import { slideUp, staggerContainer } from '@/lib/animations/variants'
import { testDatabaseConnection, type DatabaseConnection } from '@/lib/api/database'
import { login, register, checkAuth } from '@/lib/auth'


export default function LoginPage() {
  const router = useRouter()
  const [isSignup, setIsSignup] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [checkingAuth, setCheckingAuth] = useState(true)

  // PostgreSQL connection settings
  const [dbConfig, setDbConfig] = useState<DatabaseConnection>({
    host: 'localhost',
    port: '5432',
    database: 'neurondb',
    user: 'neurondb',
    password: 'neurondb',
    sslMode: 'disable',
  })
  
  const [connectionTested, setConnectionTested] = useState(false)
  const [testingConnection, setTestingConnection] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState<{
    success: boolean
    message: string
    latency?: number
    version?: string
  } | null>(null)
  const [showDbSettings, setShowDbSettings] = useState(false)
  const [apiServerStatus, setApiServerStatus] = useState<'checking' | 'online' | 'offline'>('checking')
  const [mounted, setMounted] = useState(false)

  // Set mounted flag to prevent hydration mismatch
  useEffect(() => {
    setMounted(true)
  }, [])

  // Check if already logged in
  useEffect(() => {
    if (!mounted) return
    
    const checkAuthStatus = async () => {
      try {
        const token = typeof window !== 'undefined' ? localStorage.getItem('api_token') : null
        if (token) {
          const isAuthenticated = await checkAuth()
          if (isAuthenticated) {
            router.push('/')
            return
          }
        }
      } catch (error) {
        console.error('Auth check failed:', error)
      } finally {
        setCheckingAuth(false)
      }
    }
    checkAuthStatus()
  }, [router, mounted])

  // Check API server status on mount
  useEffect(() => {
    const checkApiServer = async () => {
      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://127.0.0.1:8082/api/v1'
        const response = await fetch(`${apiUrl.replace('/api/v1', '')}/health`, {
          method: 'GET',
          signal: AbortSignal.timeout(3000),
        })
        if (response.ok) {
          setApiServerStatus('online')
        } else {
          setApiServerStatus('offline')
        }
      } catch (error) {
        setApiServerStatus('offline')
      }
    }
    checkApiServer()
  }, [])

  // Load saved database config from localStorage
  useEffect(() => {
    if (!mounted) return
    
    const saved = localStorage.getItem('db_connection_config')
    if (saved) {
      try {
        const config = JSON.parse(saved)
        setDbConfig(config)
        // If config was saved, assume it was tested before
        setConnectionTested(true)
        setShowDbSettings(false)
      } catch (e) {
        // Ignore parse errors
      }
    }
  }, [mounted])

  const handleTestConnection = async () => {
    setTestingConnection(true)
    setConnectionStatus(null)
    setError('')

    try {
      const result = await testDatabaseConnection(dbConfig)
      setConnectionStatus({
        success: result.success,
        message: result.message,
        latency: result.latency_ms,
        version: result.version,
      })
      
      if (result.success) {
        setConnectionTested(true)
        // Save successful config
        localStorage.setItem('db_connection_config', JSON.stringify(dbConfig))
      } else {
        setConnectionTested(false)
      }
    } catch (err: unknown) {
      console.error('Connection test error:', err)
      const errorMessage = err instanceof Error ? err.message : 'Connection test failed'
      setConnectionStatus({
        success: false,
        message: errorMessage,
      })
      setConnectionTested(false)
    } finally {
      setTestingConnection(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!connectionTested) {
      setError('Please test the database connection first')
      return
    }

    if (!username.trim() || !password.trim()) {
      setError('Username and password are required')
      return
    }

    if (isSignup && password.length < 6) {
      setError('Password must be at least 6 characters')
      return
    }

    setLoading(true)

    try {
      if (isSignup) {
        await register(username, password, dbConfig.database)
      } else {
        await login(username, password, dbConfig.database)
      }

      // Store database preference
      localStorage.setItem('selected_database', dbConfig.database)

      // Redirect to dashboard
      router.push('/')
      router.refresh()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Authentication failed')
      setLoading(false)
    }
  }

  // Don't render until mounted to prevent hydration mismatch
  if (!mounted || (checkingAuth && typeof window !== 'undefined' && localStorage.getItem('api_token'))) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 via-slate-50 to-slate-100 dark:from-slate-900 dark:via-slate-900 dark:to-slate-800">
        <div className="text-center">
          <div className="h-8 w-8 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-muted-foreground">Checking authentication...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 via-blue-50 via-purple-50 to-indigo-50 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900 px-4 py-12 relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-purple-300 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob"></div>
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-blue-300 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob animation-delay-2000"></div>
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-80 h-80 bg-indigo-300 rounded-full mix-blend-multiply filter blur-xl opacity-20 animate-blob animation-delay-4000"></div>
      </div>

      <div className="max-w-6xl w-full relative z-10">
        <div className="grid md:grid-cols-2 gap-8 items-center">
          {/* Left side - Branding */}
          <motion.div
            variants={staggerContainer}
            initial="hidden"
            animate="visible"
            className="hidden md:block"
          >
            <motion.div variants={slideUp} className="space-y-6">
              <div className="w-40 h-40 md:w-48 md:h-48 lg:w-56 lg:h-56 rounded-3xl flex items-center justify-center shadow-2xl shadow-purple-500/30 transform hover:scale-105 transition-transform overflow-hidden bg-white dark:bg-slate-800 p-3">
                <img 
                  src="/logo.png" 
                  alt="NeuronIP Logo" 
                  className="w-full h-full object-contain"
                />
              </div>
              <motion.h1
                variants={slideUp}
                className="text-5xl font-bold bg-gradient-to-r from-purple-600 via-indigo-600 to-blue-600 bg-clip-text text-transparent"
              >
                NeuronIP
              </motion.h1>
              <motion.p
                variants={slideUp}
                className="text-xl text-gray-700 dark:text-slate-300 leading-relaxed"
              >
                Enterprise Intelligence Platform
              </motion.p>
              <motion.div
                variants={slideUp}
                className="flex items-center gap-2 text-sm text-gray-600 dark:text-slate-400"
              >
                <SparklesIcon className="w-5 h-5 text-purple-500" />
                <span>AI-Native • Semantic Search • Data Warehouse Q&A</span>
              </motion.div>
            </motion.div>
          </motion.div>

          {/* Right side - Login Form */}
          <motion.div variants={slideUp} className="w-full">
            <Card className="backdrop-blur-sm bg-white/90 dark:bg-slate-900/90 border-2 shadow-2xl">
              <CardHeader className="space-y-1 pb-4">
                {/* Mobile logo - shown on small screens */}
                <div className="md:hidden flex justify-center mb-4">
                  <div className="w-20 h-20 rounded-2xl flex items-center justify-center shadow-xl shadow-purple-500/30 overflow-hidden bg-white dark:bg-slate-800 p-2">
                    <img 
                      src="/logo.png" 
                      alt="NeuronIP Logo" 
                      className="w-full h-full object-contain"
                    />
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-2xl font-bold flex items-center gap-2">
                    <KeyIcon className="h-6 w-6 text-purple-600" />
                  {isSignup ? 'Create Account' : 'Sign In'}
                </CardTitle>
                  <button
                    onClick={() => setShowDbSettings(!showDbSettings)}
                    className="text-sm text-purple-600 dark:text-purple-400 hover:text-purple-700 dark:hover:text-purple-300 transition-colors flex items-center gap-1"
                  >
                    <ServerIcon className="w-4 h-4" />
                    {showDbSettings ? 'Hide' : 'Show'} DB Settings
                  </button>
                </div>
                <CardDescription>
                  {isSignup
                    ? 'Create a new account to access NeuronIP'
                    : 'Enter your credentials to access the NeuronIP platform'}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                {/* PostgreSQL Connection Settings */}
                <AnimatePresence>
                  {showDbSettings && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      className="space-y-4 p-4 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-slate-800 dark:to-slate-700 rounded-lg border border-blue-200 dark:border-slate-600"
                    >
                      <div className="flex items-center gap-2 mb-2">
                        <ServerIcon className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                        <h3 className="font-semibold text-sm text-gray-900 dark:text-gray-100">
                          PostgreSQL Connection
                        </h3>
                      </div>
                      <div className="grid grid-cols-2 gap-3">
                        <Input
                          label="Host"
                          type="text"
                          value={dbConfig.host}
                          onChange={(e) => {
                            setDbConfig({ ...dbConfig, host: e.target.value })
                            setConnectionTested(false)
                            setConnectionStatus(null)
                          }}
                          placeholder="localhost"
                          disabled={testingConnection}
                        />
                        <Input
                          label="Port"
                          type="text"
                          value={dbConfig.port}
                          onChange={(e) => {
                            setDbConfig({ ...dbConfig, port: e.target.value })
                            setConnectionTested(false)
                            setConnectionStatus(null)
                          }}
                          placeholder="5432"
                          disabled={testingConnection}
                        />
                        <Input
                          label="Database"
                          type="text"
                          value={dbConfig.database}
                          onChange={(e) => {
                            setDbConfig({ ...dbConfig, database: e.target.value })
                            setConnectionTested(false)
                            setConnectionStatus(null)
                          }}
                          placeholder="neurondb"
                          disabled={testingConnection}
                          className="col-span-2"
                        />
                        <Input
                          label="User"
                          type="text"
                          value={dbConfig.user}
                          onChange={(e) => {
                            setDbConfig({ ...dbConfig, user: e.target.value })
                            setConnectionTested(false)
                            setConnectionStatus(null)
                          }}
                          placeholder="postgres"
                          disabled={testingConnection}
                        />
                        <Input
                          label="Password"
                          type="password"
                          value={dbConfig.password}
                          onChange={(e) => {
                            setDbConfig({ ...dbConfig, password: e.target.value })
                            setConnectionTested(false)
                            setConnectionStatus(null)
                          }}
                          placeholder="••••••••"
                          disabled={testingConnection}
                        />
                      </div>
                      
                      {/* API Server Status Indicator */}
                      {apiServerStatus === 'checking' && (
                        <div className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                          <ArrowPathIcon className="w-4 h-4 animate-spin" />
                          <span>Checking API server...</span>
                        </div>
                      )}
                      {apiServerStatus === 'offline' && (
                        <div className="flex items-center gap-2 text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-3 py-2 rounded-lg">
                          <XCircleIcon className="w-4 h-4" />
                          <span>API server is offline. Please start the server on http://127.0.0.1:8082</span>
                        </div>
                      )}
                      {apiServerStatus === 'online' && (
                        <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 px-3 py-2 rounded-lg">
                          <CheckCircleIcon className="w-4 h-4" />
                          <span>API server is online</span>
                        </div>
                      )}

                      {/* Connection Test Button and Status */}
                      <div className="space-y-2">
                        <Button
                          type="button"
                          onClick={handleTestConnection}
                          disabled={testingConnection || apiServerStatus !== 'online'}
                          className="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:opacity-50"
                        >
                          {testingConnection ? (
                            <>
                              <ArrowPathIcon className="w-4 h-4 mr-2 animate-spin" />
                              Testing Connection...
                            </>
                          ) : (
                            <>
                              <ServerIcon className="w-4 h-4 mr-2" />
                              Test Connection
                            </>
                          )}
                        </Button>
                        
                        <AnimatePresence>
                          {connectionStatus && (
                            <motion.div
                              initial={{ opacity: 0, y: -10 }}
                              animate={{ opacity: 1, y: 0 }}
                              exit={{ opacity: 0, y: -10 }}
                              className={`p-3 rounded-lg flex items-start gap-2 ${
                                connectionStatus.success
                                  ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800'
                                  : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800'
                              }`}
                            >
                              {connectionStatus.success ? (
                                <CheckCircleIcon className="w-5 h-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
                              ) : (
                                <XCircleIcon className="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
                              )}
                              <div className="flex-1">
                                <p
                                  className={`text-sm font-medium ${
                                    connectionStatus.success
                                      ? 'text-green-800 dark:text-green-300'
                                      : 'text-red-800 dark:text-red-300'
                                  }`}
                                >
                                  {connectionStatus.message}
                                </p>
                                {connectionStatus.success && connectionStatus.latency && (
                                  <p className="text-xs text-green-600 dark:text-green-400 mt-1">
                                    Latency: {connectionStatus.latency}ms
                                    {connectionStatus.version && ` • ${connectionStatus.version.split(' ')[0]} ${connectionStatus.version.split(' ')[1]}`}
                                  </p>
                                )}
                              </div>
                            </motion.div>
                          )}
                        </AnimatePresence>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>

                {/* Login Form */}
                <form onSubmit={handleSubmit} className="space-y-4">
                  <Input
                    label="Username / Email"
                    type="text"
                    value={username}
                    onChange={(e) => {
                      setUsername(e.target.value)
                      setError('')
                    }}
                    placeholder="Enter your username or email"
                    required
                    disabled={loading || !connectionTested}
                  />

                  <Input
                    label="Password"
                    type="password"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value)
                      setError('')
                    }}
                    placeholder="Enter your password"
                    required
                    disabled={loading || !connectionTested}
                    minLength={isSignup ? 6 : undefined}
                  />
                  {isSignup && (
                    <p className="text-sm text-muted-foreground">Password must be at least 6 characters</p>
                  )}

                  {error && (
                    <motion.div
                      initial={{ opacity: 0, y: -10 }}
                      animate={{ opacity: 1, y: 0 }}
                      className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 px-4 py-3 rounded-lg text-sm"
                    >
                      {error}
                    </motion.div>
                  )}

                  <Button
                    type="submit"
                    className="w-full bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 text-white font-semibold py-3 text-base shadow-lg hover:shadow-xl transition-all"
                    disabled={loading || !username.trim() || !password.trim() || !connectionTested}
                  >
                    {loading ? (
                      <>
                        <ArrowPathIcon className="w-5 h-5 mr-2 animate-spin" />
                        Please wait...
                      </>
                    ) : isSignup ? (
                      'Sign Up & Continue'
                    ) : (
                      'Sign In'
                    )}
                  </Button>
                  
                  {!connectionTested && (
                    <p className="text-sm text-amber-600 dark:text-amber-400 text-center">
                      ⚠️ Please test the database connection before signing in
                    </p>
                  )}
                </form>

                {/* Toggle Signup/Login */}
                <div className="mt-6 text-center">
                  <button
                    type="button"
                    onClick={() => {
                      setIsSignup(!isSignup)
                      setError('')
                    }}
                    className="text-sm text-purple-600 dark:text-purple-400 hover:text-purple-700 dark:hover:text-purple-300 transition-colors"
                  >
                    {isSignup ? 'Already have an account? Sign in' : "Don't have an account? Sign up"}
                  </button>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        </div>
      </div>

      {/* eslint-disable-next-line react/no-unknown-property */}
      <style jsx>{`
        @keyframes blob {
          0% {
            transform: translate(0px, 0px) scale(1);
          }
          33% {
            transform: translate(30px, -50px) scale(1.1);
          }
          66% {
            transform: translate(-20px, 20px) scale(0.9);
          }
          100% {
            transform: translate(0px, 0px) scale(1);
          }
        }
        .animate-blob {
          animation: blob 7s infinite;
        }
        .animation-delay-2000 {
          animation-delay: 2s;
        }
        .animation-delay-4000 {
          animation-delay: 4s;
        }
      `}</style>
    </div>
  )
}

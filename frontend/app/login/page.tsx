'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import Input from '@/components/ui/Input'
import Button from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { slideUp, staggerContainer } from '@/lib/animations/variants'
import { login, register, checkAuth } from '@/lib/auth'
import { CircleStackIcon, KeyIcon } from '@heroicons/react/24/outline'

export default function LoginPage() {
  const router = useRouter()
  const [isSignup, setIsSignup] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [database, setDatabase] = useState('neuronip')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  // Check if already logged in
  useEffect(() => {
    const checkAuthStatus = async () => {
      const isAuthenticated = await checkAuth()
      if (isAuthenticated) {
        router.push('/dashboard')
      }
    }
    checkAuthStatus()
  }, [router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    if (!username.trim() || !password.trim()) {
      setError('Username and password are required')
      setLoading(false)
      return
    }

    if (isSignup && password.length < 6) {
      setError('Password must be at least 6 characters')
      setLoading(false)
      return
    }

    try {
      if (isSignup) {
        await register(username, password, database)
      } else {
        await login(username, password, database)
      }

      // Store database preference
      localStorage.setItem('selected_database', database)

      // Redirect to dashboard
      router.push('/dashboard')
      router.refresh()
    } catch (err: any) {
      setError(err.message || 'Authentication failed')
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 via-slate-50 to-slate-100 dark:from-slate-900 dark:via-slate-900 dark:to-slate-800 px-4 py-12">
      <div className="max-w-7xl w-full">
        <div className="max-w-2xl mx-auto space-y-6">
          {/* Header */}
          <motion.div
            variants={staggerContainer}
            initial="hidden"
            animate="visible"
            className="text-center mb-8"
          >
            <motion.div variants={slideUp} className="w-20 h-20 bg-gradient-to-br from-purple-500 via-purple-600 to-indigo-600 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-2xl shadow-purple-500/30">
              <CircleStackIcon className="w-10 h-10 text-white" />
            </motion.div>
            <motion.h1 variants={slideUp} className="text-4xl font-bold bg-gradient-to-r from-purple-600 to-indigo-600 bg-clip-text text-transparent mb-2">
              NeuronIP
            </motion.h1>
            <motion.p variants={slideUp} className="text-gray-700 dark:text-slate-400">
              {isSignup ? 'Create your account' : 'Sign in to your account'}
            </motion.p>
          </motion.div>

          {/* Main Card */}
          <motion.div variants={slideUp}>
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <KeyIcon className="h-5 w-5" />
                  {isSignup ? 'Create Account' : 'Sign In'}
                </CardTitle>
                <CardDescription>
                  {isSignup
                    ? 'Create a new account to access NeuronIP'
                    : 'Enter your credentials to access the NeuronIP platform'}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSubmit} className="space-y-4">
                  {/* Username */}
                  <Input
                    label="Username"
                    type="text"
                    value={username}
                    onChange={(e) => {
                      setUsername(e.target.value)
                      setError('')
                    }}
                    placeholder="Enter your username"
                    required
                    disabled={loading}
                  />

                  {/* Password */}
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
                    disabled={loading}
                    minLength={isSignup ? 6 : undefined}
                  />
                  {isSignup && (
                    <p className="text-sm text-muted-foreground">Password must be at least 6 characters</p>
                  )}

                  {/* Database Selection */}
                  <div>
                    <label htmlFor="database" className="block text-sm font-medium text-foreground mb-2">
                      Database
                    </label>
                    <select
                      id="database"
                      value={database}
                      onChange={(e) => setDatabase(e.target.value)}
                      disabled={loading}
                      className="w-full px-4 py-3 bg-background border border-input rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <option value="neuronip">NeuronIP Database</option>
                      <option value="neuronai-demo">NeuronAI Demo Database</option>
                    </select>
                    <p className="mt-1.5 text-sm text-muted-foreground">
                      Select which database to use for your session
                    </p>
                  </div>

                  {/* Error Message */}
                  {error && (
                    <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 px-4 py-3 rounded-lg text-sm">
                      {error}
                    </div>
                  )}

                  {/* Submit Button */}
                  <Button
                    type="submit"
                    className="w-full"
                    disabled={loading || !username.trim() || !password.trim()}
                  >
                    {loading ? 'Please wait...' : isSignup ? 'Sign Up & Continue' : 'Sign In'}
                  </Button>
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

                {/* API Key Fallback */}
                <div className="mt-6 pt-6 border-t border-border">
                  <p className="text-sm text-muted-foreground mb-3">
                    Prefer API key authentication?
                  </p>
                  <p className="text-xs text-muted-foreground">
                    You can still use API keys for programmatic access. Set your API key in Settings after logging in.
                  </p>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        </div>
      </div>
    </div>
  )
}

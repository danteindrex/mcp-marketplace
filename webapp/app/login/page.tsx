'use client'

import { useEffect, useState, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { ArrowRight, ShieldCheck, Sparkles } from 'lucide-react'
import { Text } from '@/components/retroui/Text'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Sticker } from '@/components/ui/sticker'
import { BurstShape, LightningShape, Star5Shape } from '@/components/ui/shapes'
import { completeOAuthSignup, loginWithCredentials, signupWithCredentials, startOAuthFlow, type AppRole } from '@/lib/api'
import { getDashboardPath, setAuthSession } from '@/lib/auth-session'

type Mode = 'login' | 'signup'

function LoginForm() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [mode, setMode] = useState<Mode>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [mfaCode, setMfaCode] = useState('')
  const [showMFA, setShowMFA] = useState(false)
  const [name, setName] = useState('')
  const [tenantName, setTenantName] = useState('')
  const [role, setRole] = useState<'buyer' | 'merchant'>('buyer')
  const [rememberMe, setRememberMe] = useState(true)
  const [nextPath, setNextPath] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [oauthLoading, setOauthLoading] = useState<'google' | 'github' | null>(null)
  const [oauthSignupToken, setOauthSignupToken] = useState<string | null>(null)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    setNextPath(params.get('next'))
    
    // Handle OAuth callback success
    const oauth = params.get('oauth')
    const errorParam = params.get('error')
    
    if (oauth === 'success') {
      // Read role from cookie and redirect
      const roleCookie = document.cookie.match(/mcp_active_role=([^;]+)/)
      if (roleCookie) {
        const userRole = roleCookie[1] as AppRole
        setAuthSession(userRole)
        router.push(nextPath || getDashboardPath(userRole))
      } else {
        // Default to buyer if no role found
        setAuthSession('buyer')
        router.push(nextPath || getDashboardPath('buyer'))
      }
    } else if (oauth === 'complete') {
      setMode('signup')
      setShowMFA(false)
      setPassword('')
      setOauthSignupToken(params.get('signupToken'))
      setEmail(params.get('email') || '')
      if (params.get('name')) {
        setName(params.get('name') || '')
      }
    } else if (errorParam) {
      setError(decodeURIComponent(errorParam))
    }
  }, [router, searchParams, nextPath])

  const finalizeAuth = async (userRole: AppRole) => {
    setAuthSession(userRole)
    router.push(nextPath || getDashboardPath(userRole))
  }

  const handleOAuthLogin = async (provider: 'google' | 'github') => {
    setOauthLoading(provider)
    setError(null)
    try {
      const authUrl = await startOAuthFlow(provider, { mode })
      // Redirect to the OAuth provider
      window.location.href = authUrl
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : 'OAuth failed'
      setError(message)
      setOauthLoading(null)
    }
  }

  const handleSubmit = async () => {
    setLoading(true)
    setError(null)
    try {
      if (mode === 'signup') {
        const data = oauthSignupToken
          ? await completeOAuthSignup({ signupToken: oauthSignupToken, name, role, tenantName })
          : await signupWithCredentials({ email, password, name, role, tenantName })
        await finalizeAuth(data.user.role)
      } else {
        const data = await loginWithCredentials(email, password, showMFA ? mfaCode : undefined)
        await finalizeAuth(data.user.role)
      }
    } catch (e: any) {
      const message = e?.message || 'Authentication failed'
      if (mode === 'login' && message === 'mfa_required') {
        setShowMFA(true)
        setError('Enter your authenticator code to continue.')
      } else {
        setError(message)
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="min-h-screen grid-pattern">
      <div className="mx-auto min-h-screen max-w-7xl px-4 py-8 lg:px-6 lg:py-10">
        <div className="grid min-h-[calc(100vh-5rem)] overflow-hidden border-2 border-foreground bg-card shadow-[8px_8px_0px_hsl(var(--shadow-color))] lg:grid-cols-2">
          <section className="relative hidden overflow-hidden border-r-2 border-foreground bg-accent lg:flex lg:flex-col lg:justify-between lg:p-10">
            <div className="space-y-5">
              <Sticker
                variant="primary"
                rotation="slight-right"
                className="w-fit !bg-[var(--tertiary)] !text-[var(--tertiary-foreground)]"
              >
                MCP Marketplace
              </Sticker>
              <Text variant="h2" className="max-w-lg uppercase tracking-tight">
                Build and sell MCP servers with one secure account.
              </Text>
              <Text variant="small" className="max-w-xl text-foreground/80">
                OAuth-ready marketplace access for buyer, merchant, and admin workflows with a single login.
              </Text>
            </div>

            <div className="relative mt-8 space-y-4 border-2 border-foreground bg-background p-5 shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
              <Text variant="caption" className="uppercase tracking-wide text-muted-foreground">What you get</Text>
              <ul className="space-y-3 text-sm font-semibold">
                <li className="flex items-center gap-2"><ShieldCheck className="h-4 w-4 text-tertiary" />PKCE + OAuth-protected server access</li>
                <li className="flex items-center gap-2"><Sparkles className="h-4 w-4 text-tertiary" />Role-aware dashboards and analytics</li>
                <li className="flex items-center gap-2"><ArrowRight className="h-4 w-4 text-tertiary" />One-click marketplace onboarding</li>
              </ul>
            </div>

            <BurstShape size={82} className="absolute right-8 top-8 text-tertiary" />
            <LightningShape size={70} className="absolute bottom-12 right-14 text-tertiary" />
            <Star5Shape size={64} className="absolute bottom-28 left-10 text-secondary" />
          </section>

          <section className="flex items-center justify-center bg-background p-6 sm:p-10">
            <Card className="w-full max-w-md border-2 border-foreground bg-card p-6 shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
              <div className="mb-5 space-y-1 text-center">
                <Text as="span" variant="caption" className="inline-flex items-center gap-2 uppercase tracking-wide text-muted-foreground"><ShieldCheck className="h-4 w-4" />Secure account access</Text>
                <Text variant="h3" className="uppercase">{mode === 'login' ? 'Welcome back' : 'Create account'}</Text>
                <Text variant="small" className="text-muted-foreground">
                  {oauthSignupToken
                    ? 'Finish your social signup by choosing a role and workspace.'
                    : mode === 'login'
                      ? 'Enter your credentials to continue.'
                      : 'Create your buyer or merchant workspace.'}
                </Text>
              </div>

              {error && <div className="mb-4 rounded-none border-2 border-destructive bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</div>}

              <div className="mb-4 grid grid-cols-2 gap-2">
                <Button
                  variant={mode === 'login' ? 'default' : 'outline'}
                  disabled={Boolean(oauthSignupToken)}
                  onClick={() => {
                    setMode('login')
                  }}
                >
                  Login
                </Button>
                <Button
                  variant={mode === 'signup' ? 'default' : 'outline'}
                  onClick={() => {
                    setMode('signup')
                    setShowMFA(false)
                    setMfaCode('')
                  }}
                >
                  Signup
                </Button>
              </div>

              {!oauthSignupToken && <div className="mb-4 space-y-3">
                <div className="relative">
                  <div className="absolute inset-0 flex items-center">
                    <span className="w-full border-t" />
                  </div>
                  <div className="relative flex justify-center text-xs uppercase">
                    <span className="bg-card px-2 text-muted-foreground">
                      {mode === 'login' ? 'Or continue with' : 'Or sign up with'}
                    </span>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-2">
                  <Button
                    variant="outline"
                    onClick={() => handleOAuthLogin('google')}
                    disabled={oauthLoading !== null}
                    className="gap-2"
                  >
                    {oauthLoading === 'google' ? (
                      <span className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    ) : (
                      <svg viewBox="0 0 24 24" className="h-4 w-4" fill="currentColor">
                        <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                        <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                        <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                        <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
                      </svg>
                    )}
                    Google
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => handleOAuthLogin('github')}
                    disabled={oauthLoading !== null}
                    className="gap-2"
                  >
                    {oauthLoading === 'github' ? (
                      <span className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                    ) : (
                      <svg viewBox="0 0 24 24" className="h-4 w-4" fill="currentColor">
                        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                      </svg>
                    )}
                    GitHub
                  </Button>
                </div>
              </div>}

              <form
                className="space-y-4"
                onSubmit={event => {
                  event.preventDefault()
                  void handleSubmit()
                }}
                >
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    placeholder="you@example.com"
                    value={email}
                    onChange={e => setEmail(e.target.value)}
                    readOnly={Boolean(oauthSignupToken)}
                  />
                </div>
                {!oauthSignupToken && <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input id="password" type="password" placeholder="Enter your password" value={password} onChange={e => setPassword(e.target.value)} />
                </div>}
                {mode === 'login' && showMFA && (
                  <div className="space-y-2">
                    <Label htmlFor="mfaCode">Authenticator code</Label>
                    <Input
                      id="mfaCode"
                      placeholder="123456"
                      inputMode="numeric"
                      value={mfaCode}
                      onChange={e => setMfaCode(e.target.value)}
                    />
                  </div>
                )}

                {mode === 'signup' && (
                  <>
                    <div className="space-y-2">
                      <Label htmlFor="name">Full name</Label>
                      <Input id="name" value={name} onChange={e => setName(e.target.value)} />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="tenantName">Workspace / Tenant name</Label>
                      <Input id="tenantName" value={tenantName} onChange={e => setTenantName(e.target.value)} />
                    </div>
                    <div className="space-y-2">
                      <Label>Role</Label>
                      <div className="grid grid-cols-2 gap-2">
                        <Button type="button" variant={role === 'buyer' ? 'default' : 'outline'} onClick={() => setRole('buyer')}>Buyer</Button>
                        <Button type="button" variant={role === 'merchant' ? 'default' : 'outline'} onClick={() => setRole('merchant')}>Merchant</Button>
                      </div>
                    </div>
                  </>
                )}

                {mode === 'login' && !oauthSignupToken && (
                  <label className="flex items-center gap-2 text-sm font-medium">
                    <Checkbox checked={rememberMe} onCheckedChange={checked => setRememberMe(Boolean(checked))} />
                    Remember me
                  </label>
                )}

                <Button type="submit" disabled={loading} className="w-full">
                  {loading ? 'Please wait...' : mode === 'login' ? 'Sign In' : oauthSignupToken ? 'Complete Signup' : 'Create Account'}
                </Button>
              </form>

              <div className="mt-5 text-center">
                <Link href="/" className="font-semibold hover:text-foreground">Back to home</Link>
              </div>
            </Card>
          </section>
        </div>
      </div>
    </main>
  )
}

function LoginLoading() {
  return (
    <main className="min-h-screen grid-pattern">
      <div className="mx-auto min-h-screen max-w-7xl px-4 py-8 lg:px-6 lg:py-10">
        <div className="grid min-h-[calc(100vh-5rem)] overflow-hidden border-2 border-foreground bg-card shadow-[8px_8px_0px_hsl(var(--shadow-color))] lg:grid-cols-2">
          <section className="relative hidden overflow-hidden border-r-2 border-foreground bg-accent lg:flex lg:flex-col lg:justify-between lg:p-10" />
          <section className="flex items-center justify-center bg-background p-6 sm:p-10">
            <Card className="w-full max-w-md border-2 border-foreground bg-card p-6 shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
              <div className="flex items-center justify-center py-20">
                <span className="h-8 w-8 animate-spin rounded-full border-4 border-current border-t-transparent" />
              </div>
            </Card>
          </section>
        </div>
      </div>
    </main>
  )
}

export default function LoginPage() {
  return (
    <Suspense fallback={<LoginLoading />}>
      <LoginForm />
    </Suspense>
  )
}

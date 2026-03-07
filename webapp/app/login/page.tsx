'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowRight, ShieldCheck, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Sticker } from '@/components/ui/sticker'
import { BurstShape, LightningShape, Star5Shape } from '@/components/ui/shapes'
import { loginWithCredentials, signupWithCredentials, type AppRole } from '@/lib/api'
import { getDashboardPath, setAuthSession } from '@/lib/auth-session'

type Mode = 'login' | 'signup'

export default function LoginPage() {
  const router = useRouter()
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

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    setNextPath(params.get('next'))
  }, [])

  const finalizeAuth = async (userRole: AppRole) => {
    setAuthSession(userRole)
    router.push(nextPath || getDashboardPath(userRole))
  }

  const handleSubmit = async () => {
    setLoading(true)
    setError(null)
    try {
      if (mode === 'signup') {
        const data = await signupWithCredentials({ email, password, name, role, tenantName })
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
              <h2 className="max-w-lg text-4xl font-black uppercase tracking-tight">
                Build and sell MCP servers with one secure account.
              </h2>
              <p className="max-w-xl text-sm font-medium text-foreground/80">
                OAuth-ready marketplace access for buyer, merchant, and admin workflows with a single login.
              </p>
            </div>

            <div className="relative mt-8 space-y-4 border-2 border-foreground bg-background p-5 shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
              <p className="text-xs font-bold uppercase tracking-wide text-muted-foreground">What you get</p>
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
                <p className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground"><ShieldCheck className="h-4 w-4" />Secure account access</p>
                <h1 className="text-3xl font-black uppercase">{mode === 'login' ? 'Welcome back' : 'Create account'}</h1>
                <p className="text-sm text-muted-foreground">{mode === 'login' ? 'Enter your credentials to continue.' : 'Create your buyer or merchant workspace.'}</p>
              </div>

              {error && <div className="mb-4 rounded-none border-2 border-destructive bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</div>}

              <div className="mb-4 grid grid-cols-2 gap-2">
                <Button
                  variant={mode === 'login' ? 'default' : 'outline'}
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

              <form
                className="space-y-4"
                onSubmit={event => {
                  event.preventDefault()
                  void handleSubmit()
                }}
              >
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input id="email" placeholder="you@example.com" value={email} onChange={e => setEmail(e.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input id="password" type="password" placeholder="Enter your password" value={password} onChange={e => setPassword(e.target.value)} />
                </div>
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

                {mode === 'login' && (
                  <label className="flex items-center gap-2 text-sm font-medium">
                    <Checkbox checked={rememberMe} onCheckedChange={checked => setRememberMe(Boolean(checked))} />
                    Remember me
                  </label>
                )}

                <Button type="submit" disabled={loading} className="w-full">
                  {loading ? 'Please wait...' : mode === 'login' ? 'Sign In' : 'Create Account'}
                </Button>
              </form>

              <div className="mt-5 text-center text-sm text-muted-foreground">
                <Link href="/" className="font-semibold hover:text-foreground">Back to home</Link>
              </div>
            </Card>
          </section>
        </div>
      </div>
    </main>
  )
}

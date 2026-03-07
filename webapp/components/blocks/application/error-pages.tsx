'use client'

import { type ReactNode, useMemo, useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import {
  AlertTriangle,
  Ban,
  Clock3,
  Home,
  RefreshCcw,
  Search,
  ServerCrash,
  WifiOff,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

type GenericAction = {
  label: string
  href?: string
  onClick?: () => void
  variant?: 'default' | 'outline' | 'secondary' | 'ghost'
}

function ScreenShell({
  icon,
  title,
  description,
  children,
}: {
  icon: ReactNode
  title: string
  description: string
  children?: ReactNode
}) {
  return (
    <main className="min-h-screen bg-background grid-pattern px-4 py-12">
      <div className="mx-auto max-w-2xl">
        <Card className="border-[3px] border-foreground shadow-[6px_6px_0px_hsl(var(--shadow-color))]">
          <CardHeader className="border-b-[3px] border-foreground">
            <div className="mb-2 inline-flex h-12 w-12 items-center justify-center border-[3px] border-foreground bg-accent text-foreground shadow-[3px_3px_0px_hsl(var(--shadow-color))]">
              {icon}
            </div>
            <CardTitle className="text-2xl font-black uppercase tracking-wide">{title}</CardTitle>
            <CardDescription className="text-sm text-muted-foreground">{description}</CardDescription>
          </CardHeader>
          <CardContent className="pt-6">{children}</CardContent>
        </Card>
      </div>
    </main>
  )
}

export function Offline({ onRetry }: { onRetry?: () => void }) {
  return (
    <ScreenShell
      icon={<WifiOff className="h-6 w-6" />}
      title="You Are Offline"
      description="Please check your internet connection and try again."
    >
      <div className="flex flex-wrap gap-3">
        <Button onClick={onRetry} className="gap-2">
          <RefreshCcw className="h-4 w-4" />
          Retry
        </Button>
        <Button variant="outline" asChild>
          <Link href="/">
            <Home className="mr-2 h-4 w-4" />
            Home
          </Link>
        </Button>
      </div>
    </ScreenShell>
  )
}

export function NotFound({
  homeHref = '/',
  backHref,
  showSearch = false,
  onSearch,
}: {
  homeHref?: string
  backHref?: string
  showSearch?: boolean
  onSearch?: (query: string) => void
}) {
  const router = useRouter()
  const [query, setQuery] = useState('')

  return (
    <ScreenShell
      icon={<Search className="h-6 w-6" />}
      title="Page Not Found"
      description="The page you requested does not exist or was moved."
    >
      {showSearch && (
        <form
          className="mb-5 flex gap-2"
          onSubmit={(e) => {
            e.preventDefault()
            if (!query.trim()) return
            if (onSearch) {
              onSearch(query.trim())
              return
            }
            router.push(`/marketplace?search=${encodeURIComponent(query.trim())}`)
          }}
        >
          <Input
            placeholder="Search the marketplace..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <Button type="submit">Search</Button>
        </form>
      )}
      <div className="flex flex-wrap gap-3">
        <Button asChild>
          <Link href={homeHref}>Go Home</Link>
        </Button>
        {backHref && (
          <Button variant="outline" asChild>
            <Link href={backHref}>Go Back</Link>
          </Button>
        )}
      </div>
    </ScreenShell>
  )
}

export function ServerError({
  onRetry,
  homeHref = '/',
  errorId,
  supportEmail,
}: {
  onRetry?: () => void
  homeHref?: string
  errorId?: string
  supportEmail?: string
}) {
  return (
    <ScreenShell
      icon={<ServerCrash className="h-6 w-6" />}
      title="Server Error"
      description="Something went wrong while processing your request."
    >
      <div className="space-y-4">
        {(errorId || supportEmail) && (
          <div className="rounded-md border border-border bg-muted/40 p-3 text-sm">
            {errorId && <p className="mb-1"><span className="font-semibold">Error ID:</span> {errorId}</p>}
            {supportEmail && <p><span className="font-semibold">Support:</span> {supportEmail}</p>}
          </div>
        )}
        <div className="flex flex-wrap gap-3">
          <Button onClick={onRetry} className="gap-2">
            <RefreshCcw className="h-4 w-4" />
            Try Again
          </Button>
          <Button variant="outline" asChild>
            <Link href={homeHref}>
              <Home className="mr-2 h-4 w-4" />
              Home
            </Link>
          </Button>
        </div>
      </div>
    </ScreenShell>
  )
}

export function Forbidden({
  homeHref = '/',
  loginHref = '/login',
}: {
  homeHref?: string
  loginHref?: string
}) {
  return (
    <ScreenShell
      icon={<Ban className="h-6 w-6" />}
      title="Access Forbidden"
      description="You do not have permission to access this resource."
    >
      <div className="flex flex-wrap gap-3">
        <Button asChild>
          <Link href={loginHref}>Login</Link>
        </Button>
        <Button variant="outline" asChild>
          <Link href={homeHref}>Home</Link>
        </Button>
      </div>
    </ScreenShell>
  )
}

export function ComingSoon({
  launchDate,
  onNotify,
}: {
  launchDate?: Date | string
  onNotify?: (email: string) => void
}) {
  const [email, setEmail] = useState('')
  const launchLabel = useMemo(() => {
    if (!launchDate) return 'Launching soon'
    const d = typeof launchDate === 'string' ? new Date(launchDate) : launchDate
    if (Number.isNaN(d.getTime())) return 'Launching soon'
    return `Expected launch: ${d.toLocaleDateString()}`
  }, [launchDate])

  return (
    <ScreenShell
      icon={<Clock3 className="h-6 w-6" />}
      title="Coming Soon"
      description={launchLabel}
    >
      <form
        className="flex gap-2"
        onSubmit={(e) => {
          e.preventDefault()
          if (!email.trim() || !onNotify) return
          onNotify(email.trim())
          setEmail('')
        }}
      >
        <Input
          type="email"
          placeholder="you@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <Button type="submit">Notify Me</Button>
      </form>
    </ScreenShell>
  )
}

export function Generic({
  title = 'Oops!',
  description = 'Something went wrong.',
  actions = [],
}: {
  title?: string
  description?: string
  actions?: GenericAction[]
}) {
  return (
    <ScreenShell
      icon={<AlertTriangle className="h-6 w-6" />}
      title={title}
      description={description}
    >
      <CardFooter className="flex flex-wrap gap-3 border-t border-border px-0 pt-4">
        {actions.map((action) => {
          if (action.href) {
            return (
              <Button key={action.label} variant={action.variant ?? 'default'} asChild>
                <Link href={action.href}>{action.label}</Link>
              </Button>
            )
          }
          return (
            <Button
              key={action.label}
              variant={action.variant ?? 'default'}
              onClick={action.onClick}
            >
              {action.label}
            </Button>
          )
        })}
      </CardFooter>
    </ScreenShell>
  )
}

export const ErrorPages = {
  Offline,
  NotFound,
  ServerError,
  Forbidden,
  ComingSoon,
  Generic,
}

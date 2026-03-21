'use client'

import { ReactNode, useEffect, useMemo, useState } from 'react'
import { useTheme } from 'next-themes'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Moon, Sun, Bell, Command, Settings, LogOut } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { AppSidebar } from './sidebar'
import { SidebarProvider } from '@/components/ui/sidebar'
import { Breadcrumb, Text } from '@/components/retroui'
import { logoutAndRedirect } from '@/lib/auth-session'
import { fetchCurrentUser } from '@/lib/api-client'

export interface AppShellProps {
  children: ReactNode
  role?: 'buyer' | 'merchant' | 'admin'
  userEmail?: string
}

export function AppShell({ children, role = 'buyer', userEmail = '' }: AppShellProps) {
  const { theme, setTheme } = useTheme()
  const pathname = usePathname()
  const crumbs = pathname.split('/').filter(Boolean)
  const [resolvedEmail, setResolvedEmail] = useState(userEmail)
  const [resolvedRole, setResolvedRole] = useState<'buyer' | 'merchant' | 'admin'>(role)

  useEffect(() => {
    fetchCurrentUser().then(user => {
      if (!user) return
      if (user.email) setResolvedEmail(user.email)
      if (user.role) setResolvedRole(user.role)
    })
  }, [])

  const dashboardLabel = useMemo(() => {
    if (resolvedRole === 'admin') return 'Admin'
    if (resolvedRole === 'merchant') return 'Publisher'
    return 'Buyer'
  }, [resolvedRole])

  return (
    <SidebarProvider>
      <div className="dashboard-shell flex h-screen w-full">
        <AppSidebar role={resolvedRole} />
        <div className="flex flex-col flex-1 overflow-hidden">
          {/* Header */}
          <header
            data-shell-header
            className="min-h-16 flex items-center justify-between border-b-2 border-foreground bg-card px-6 py-2 dark:border-border dark:bg-background"
          >
            <div className="flex items-center gap-2">
              <Command className="w-5 h-5 text-primary" />
              <Text as="span" variant="small" className="font-medium">MCP Marketplace</Text>
              <Text as="span" variant="caption" className="bg-muted px-2 py-1 rounded ml-2 text-muted-foreground">
                {dashboardLabel} Dashboard
              </Text>
            </div>

            <div className="flex items-center gap-4">
              {/* Notifications */}
              <Button variant="ghost" size="icon" className="relative">
                <Bell className="w-5 h-5" />
                <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" />
              </Button>

              {/* Theme Toggle */}
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              >
                {theme === 'dark' ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
              </Button>

              {/* User Menu */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <div className="w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-medium">
                      {(resolvedEmail.trim().charAt(0) || '?').toUpperCase()}
                    </div>
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-56">
                  <DropdownMenuLabel className="flex flex-col">
                    <Text as="span" variant="small" className="font-medium">{resolvedEmail || 'Account'}</Text>
                    <Text as="span" variant="caption" className="text-muted-foreground">{resolvedRole}</Text>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem>
                    <Settings className="w-4 h-4 mr-2" />
                    Account Settings
                  </DropdownMenuItem>
                  <DropdownMenuItem>
                    <Bell className="w-4 h-4 mr-2" />
                    Notifications
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="text-red-600"
                    onSelect={event => {
                      event.preventDefault()
                      void logoutAndRedirect('/login')
                    }}
                  >
                    <LogOut className="w-4 h-4 mr-2" />
                    Sign Out
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </header>
          <div
            data-shell-breadcrumb
            className="border-b-2 border-foreground/70 bg-card/95 px-6 py-2 dark:border-border/60 dark:bg-background"
          >
            <Breadcrumb>
              <Breadcrumb.List>
                <Breadcrumb.Item>
                  <Breadcrumb.Link asChild>
                    <Link href="/">Home</Link>
                  </Breadcrumb.Link>
                </Breadcrumb.Item>
                {crumbs.map((segment, idx) => {
                  const href = `/${crumbs.slice(0, idx + 1).join('/')}`
                  const label = segment.replace(/-/g, ' ')
                  const isLast = idx === crumbs.length - 1
                  return (
                    [
                      <Breadcrumb.Separator key={`${href}-sep`} />,
                      <Breadcrumb.Item key={href}>
                        {isLast ? (
                          <Breadcrumb.Page>{label}</Breadcrumb.Page>
                        ) : (
                          <Breadcrumb.Link asChild>
                            <Link href={href}>{label}</Link>
                          </Breadcrumb.Link>
                        )}
                      </Breadcrumb.Item>
                    ]
                  )
                })}
              </Breadcrumb.List>
            </Breadcrumb>
          </div>

          {/* Main Content */}
          <main className="flex-1 overflow-auto bg-background/90 dark:bg-background">
            {children}
          </main>
        </div>
      </div>
    </SidebarProvider>
  )
}

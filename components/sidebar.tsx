'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  Home,
  Zap,
  Settings,
  ShoppingCart,
  BarChart3,
  Users,
  Shield,
  FileText,
  Package,
  GitBranch,
  Workflow,
  TrendingUp,
  Lock,
  Grid,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'

interface AppSidebarProps {
  role?: 'buyer' | 'merchant' | 'admin'
}

const buyerNavigation = [
  {
    label: 'Main',
    items: [
      { name: 'Home', href: '/', icon: Home },
      { name: 'Marketplace', href: '/marketplace', icon: ShoppingCart },
    ],
  },
  {
    label: 'My Account',
    items: [
      { name: 'Dashboard', href: '/buyer/dashboard', icon: BarChart3 },
      { name: 'Connections', href: '/buyer/connections', icon: Zap },
      { name: 'Billing', href: '/buyer/billing', icon: TrendingUp },
    ],
  },
]

const merchantNavigation = [
  {
    label: 'Publish',
    items: [
      { name: 'Dashboard', href: '/merchant/servers', icon: Home },
      { name: 'New Server', href: '/merchant/servers/new/import-docker', icon: Package },
      { name: 'Onboarding', href: '/merchant/onboarding', icon: Workflow },
    ],
  },
  {
    label: 'Server Settings',
    items: [
      { name: 'Builder', href: '/merchant/servers/[serverId]/builder', icon: Grid },
      { name: 'Deployments', href: '/merchant/servers/[serverId]/deployments', icon: GitBranch },
      { name: 'Auth & Scopes', href: '/merchant/servers/[serverId]/auth', icon: Lock },
      { name: 'Pricing', href: '/merchant/servers/[serverId]/pricing', icon: ShoppingCart },
      { name: 'Observability', href: '/merchant/servers/[serverId]/observability', icon: BarChart3 },
    ],
  },
  {
    label: 'Revenue',
    items: [
      { name: 'Dashboard', href: '/merchant/revenue', icon: TrendingUp },
    ],
  },
]

const adminNavigation = [
  {
    label: 'Administration',
    items: [
      { name: 'Dashboard', href: '/admin/dashboard', icon: Home },
      { name: 'Tenants', href: '/admin/tenants', icon: Users },
      { name: 'Security Events', href: '/admin/security', icon: Shield },
      { name: 'Audit Logs', href: '/admin/audit-logs', icon: FileText },
      { name: 'Client Compatibility', href: '/admin/client-compatibility', icon: Grid },
    ],
  },
]

export function AppSidebar({ role = 'buyer' }: AppSidebarProps) {
  const pathname = usePathname()
  const { state } = useSidebar()

  const navigationConfig =
    role === 'admin' ? adminNavigation : role === 'merchant' ? merchantNavigation : buyerNavigation

  const isActive = (href: string) => {
    // Handle dynamic routes
    if (href.includes('[')) {
      const baseHref = href.split('[')[0]
      return pathname.startsWith(baseHref)
    }
    return pathname === href
  }

  return (
    <Sidebar>
      <SidebarContent>
        {navigationConfig.map(group => (
          <SidebarGroup key={group.label}>
            <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
            <SidebarMenu>
              {group.items.map(item => {
                const Icon = item.icon
                return (
                  <SidebarMenuItem key={item.name}>
                    <SidebarMenuButton
                      asChild
                      isActive={isActive(item.href)}
                      className={cn(
                        'transition-colors',
                        isActive(item.href) && 'bg-sidebar-accent text-sidebar-accent-foreground'
                      )}
                    >
                      <Link href={item.href}>
                        <Icon className="w-4 h-4" />
                        <span>{item.name}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroup>
        ))}
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton asChild>
              <Link href="/settings">
                <Settings className="w-4 h-4" />
                <span>Settings</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  )
}

'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'
import { Home, Zap, Settings, ShoppingCart, BarChart3, Users, Shield, FileText, Package, GitBranch, Workflow, TrendingUp, Lock, Grid } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from '@/components/ui/sidebar'
import { TeamSelector, ProfileSelector, type Team, type Profile } from '@/components/kokonut'
import { apiGet } from '@/lib/api'

interface AppSidebarProps {
  role?: 'buyer' | 'merchant' | 'admin'
}

const buyerNavigation = [
  { label: 'Main', items: [{ name: 'Home', href: '/', icon: Home }, { name: 'Marketplace', href: '/marketplace', icon: ShoppingCart }] },
  { label: 'My Account', items: [{ name: 'Dashboard', href: '/buyer/dashboard', icon: BarChart3 }, { name: 'Connections', href: '/buyer/connections', icon: Zap }, { name: 'Billing', href: '/buyer/billing', icon: TrendingUp }] },
]

const merchantNavigation = [
  { label: 'Publish', items: [{ name: 'Dashboard', href: '/merchant/servers', icon: Home }, { name: 'New Server', href: '/merchant/servers/new/import-docker', icon: Package }, { name: 'Onboarding', href: '/merchant/onboarding', icon: Workflow }] },
  { label: 'Server Settings', items: [{ name: 'Builder', href: '/merchant/servers/srv_postgres/builder', icon: Grid }, { name: 'Deployments', href: '/merchant/servers/srv_postgres/deployments', icon: GitBranch }, { name: 'Auth & Scopes', href: '/merchant/servers/srv_postgres/auth', icon: Lock }, { name: 'Pricing', href: '/merchant/servers/srv_postgres/pricing', icon: ShoppingCart }, { name: 'Observability', href: '/merchant/servers/srv_postgres/observability', icon: BarChart3 }] },
  { label: 'Revenue', items: [{ name: 'Dashboard', href: '/merchant/revenue', icon: TrendingUp }] },
]

const adminNavigation = [
  { label: 'Administration', items: [{ name: 'Tenants', href: '/admin/tenants', icon: Users }, { name: 'Security Events', href: '/admin/security', icon: Shield }, { name: 'Audit Logs', href: '/admin/audit-logs', icon: FileText }, { name: 'Client Compatibility', href: '/admin/client-compatibility', icon: Grid }] },
]

export function AppSidebar({ role = 'buyer' }: AppSidebarProps) {
  const pathname = usePathname()
  const { state } = useSidebar()
  const [profile, setProfile] = useState<Profile>({ id: 'loading', name: 'Loading', email: 'loading@local', avatar: '' })
  const [teams, setTeams] = useState<Team[]>([{ id: 'team', name: 'Workspace', avatar: 'W', role: 'owner' }])

  useEffect(() => {
    apiGet<any>('/v1/me', role)
      .then(user => {
        setProfile({ id: user.id, name: user.name, email: user.email, avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${encodeURIComponent(user.name)}` })
        setTeams([{ id: user.tenantId, name: user.tenantId, avatar: 'T', role: 'owner' }])
      })
      .catch(() => {
        setProfile({ id: role, name: role.toUpperCase(), email: `${role}@local`, avatar: '' })
        setTeams([{ id: role, name: role.toUpperCase(), avatar: 'T', role: 'owner' }])
      })
  }, [role])

  const navigationConfig = role === 'admin' ? adminNavigation : role === 'merchant' ? merchantNavigation : buyerNavigation

  const isActive = (href: string) => {
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
                    <SidebarMenuButton asChild isActive={isActive(item.href)} className={cn('transition-colors', isActive(item.href) && 'bg-sidebar-accent text-sidebar-accent-foreground')}>
                      <Link href={item.href}><Icon className="w-4 h-4" /><span>{item.name}</span></Link>
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
          <SidebarMenuItem><TeamSelector teams={teams} selectedTeam={teams[0]} onTeamChange={() => {}} onCreateTeam={() => {}} /></SidebarMenuItem>
          <SidebarMenuItem><ProfileSelector profile={profile} onSettings={() => {}} onLogout={() => {}} /></SidebarMenuItem>
          <SidebarMenuItem>
            <SidebarMenuButton asChild>
              <Link href="/settings"><Settings className="w-4 h-4" /><span>Settings</span></Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  )
}
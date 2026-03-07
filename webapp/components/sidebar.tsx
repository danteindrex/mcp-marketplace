'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useTheme } from 'next-themes'
import { Home, Zap, Settings, ShoppingCart, BarChart3, Users, Shield, FileText, Package, GitBranch, Workflow, TrendingUp, Lock, Grid } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar'
import { TeamSelector, ProfileSelector, type Team, type Profile } from '@/components/kokonut'
import { apiGet } from '@/lib/api'
import { Button as RetroButton, Card as RetroCard, TableOfContents, Text } from '@/components/retroui'
import { clearAuthSession } from '@/lib/auth-session'

interface AppSidebarProps {
  role?: 'buyer' | 'merchant' | 'admin'
}

const buyerNavigation = [
  { label: 'Main', items: [{ name: 'Home', href: '/', icon: Home }, { name: 'Marketplace', href: '/marketplace', icon: ShoppingCart }] },
  { label: 'My Account', items: [{ name: 'Dashboard', href: '/buyer/dashboard', icon: BarChart3 }, { name: 'Connections', href: '/buyer/connections', icon: Zap }, { name: 'Billing', href: '/buyer/billing', icon: TrendingUp }] },
]

const merchantNavigation = [
  { label: 'Publish', items: [{ name: 'Dashboard', href: '/merchant/servers', icon: Home }, { name: 'New Server', href: '/merchant/servers/new/import-docker', icon: Package }, { name: 'Onboarding', href: '/merchant/onboarding', icon: Workflow }] },
  { label: 'Revenue', items: [{ name: 'Dashboard', href: '/merchant/revenue', icon: TrendingUp }] },
]

const adminNavigation = [
  { label: 'Administration', items: [{ name: 'Tenants', href: '/admin/tenants', icon: Users }, { name: 'Security Events', href: '/admin/security', icon: Shield }, { name: 'Audit Logs', href: '/admin/audit-logs', icon: FileText }, { name: 'Client Compatibility', href: '/admin/client-compatibility', icon: Grid }, { name: 'Payments', href: '/admin/payments', icon: TrendingUp }] },
]

function dicebearAvatar(seed: string, style: 'avataaars' | 'shapes'): string {
  return `https://api.dicebear.com/7.x/${style}/svg?seed=${encodeURIComponent(seed)}`
}

function sanitizeAvatar(url: string | undefined, seed: string, style: 'avataaars' | 'shapes'): string {
  if (url && /^https?:\/\//i.test(url)) return url
  return dicebearAvatar(seed, style)
}

export function AppSidebar({ role = 'buyer' }: AppSidebarProps) {
  const pathname = usePathname()
  const router = useRouter()
  const { resolvedTheme } = useTheme()
  const [profile, setProfile] = useState<Profile>({ id: 'loading', name: 'Loading', email: 'loading@local', avatar: dicebearAvatar('loading', 'avataaars') })
  const [teams, setTeams] = useState<Team[]>([{ id: 'team', name: 'Workspace', avatar: dicebearAvatar('workspace', 'shapes'), role: 'owner' }])
  const [merchantServerId, setMerchantServerId] = useState<string | null>(null)

  useEffect(() => {
    apiGet<any>('/v1/me', role)
      .then(user => {
        const profileAvatar = sanitizeAvatar(user.avatar, user.name || user.id, 'avataaars')
        const teamAvatar = sanitizeAvatar(user.tenantAvatar, user.tenantId || user.id, 'shapes')
        setProfile({ id: user.id, name: user.name, email: user.email, avatar: profileAvatar })
        setTeams([{ id: user.tenantId, name: user.tenantId, avatar: teamAvatar, role: 'owner' }])
      })
      .catch(() => {
        setProfile({ id: role, name: role.toUpperCase(), email: `${role}@local`, avatar: dicebearAvatar(role, 'avataaars') })
        setTeams([{ id: role, name: role.toUpperCase(), avatar: dicebearAvatar(role, 'shapes'), role: 'owner' }])
      })
  }, [role])

  useEffect(() => {
    if (role !== 'merchant') {
      setMerchantServerId(null)
      return
    }
    apiGet<{ items: Array<{ id: string }> }>('/v1/merchant/servers', 'merchant')
      .then(res => {
        setMerchantServerId(res.items?.[0]?.id || null)
      })
      .catch(() => setMerchantServerId(null))
  }, [role])

  const merchantServerSettings = merchantServerId
    ? [
        { name: 'Builder', href: `/merchant/servers/${merchantServerId}/builder`, icon: Grid },
        { name: 'Deployments', href: `/merchant/servers/${merchantServerId}/deployments`, icon: GitBranch },
        { name: 'Auth & Scopes', href: `/merchant/servers/${merchantServerId}/auth`, icon: Lock },
        { name: 'Pricing', href: `/merchant/servers/${merchantServerId}/pricing`, icon: ShoppingCart },
        { name: 'Observability', href: `/merchant/servers/${merchantServerId}/observability`, icon: BarChart3 },
      ]
    : [{ name: 'Server Settings', href: '/merchant/servers', icon: Grid }]

  const navigationConfig =
    role === 'admin'
      ? adminNavigation
      : role === 'merchant'
        ? [...merchantNavigation, { label: 'Server Settings', items: merchantServerSettings }]
        : buyerNavigation

  const isActive = (href: string) => pathname === href || pathname.startsWith(`${href}/`)
  const selectedTeam = teams[0] || { id: 'workspace', name: 'Workspace', avatar: dicebearAvatar('workspace', 'shapes'), role: 'owner' as const }

  return (
    <Sidebar>
      <SidebarContent>
        {navigationConfig.map(group => (
          <SidebarGroup key={group.label}>
            {resolvedTheme === 'light' ? (
              <RetroCard className="p-2 mb-2">
                <SidebarGroupLabel className="text-black font-black tracking-wide">{group.label}</SidebarGroupLabel>
              </RetroCard>
            ) : (
              <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
            )}
            <SidebarMenu>
              {group.items.map(item => {
                const Icon = item.icon
                return (
                  <SidebarMenuItem key={item.name}>
                    {resolvedTheme === 'light' ? (
                      <div className="mb-1">
                        <RetroButton
                          asChild
                          variant="outline"
                          className={cn(
                            'w-full justify-start gap-2 text-sm border-2',
                            isActive(item.href)
                              ? '!bg-[var(--tertiary)] !text-[var(--tertiary-foreground)] !border-black'
                              : '!bg-white !text-black !border-black',
                          )}
                        >
                          <Link href={item.href} className="flex items-center gap-2 w-full">
                            <Icon className={cn('w-4 h-4', isActive(item.href) ? 'text-[var(--tertiary-foreground)]' : 'text-black')} />
                            <span className={cn('text-sm font-bold', isActive(item.href) ? 'text-[var(--tertiary-foreground)]' : 'text-black')}>{item.name}</span>
                          </Link>
                        </RetroButton>
                      </div>
                    ) : (
                      <SidebarMenuButton asChild isActive={isActive(item.href)} className={cn('transition-colors', isActive(item.href) && 'bg-sidebar-accent text-sidebar-accent-foreground')}>
                        <Link href={item.href}><Icon className="w-4 h-4" /><span>{item.name}</span></Link>
                      </SidebarMenuButton>
                    )}
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroup>
        ))}
        {resolvedTheme === 'light' ? (
          <SidebarGroup>
            <TableOfContents className="w-full h-44">
              <Text variant="small" className="mb-2 border-b border-black pb-1">On This Page</Text>
            </TableOfContents>
          </SidebarGroup>
        ) : null}
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem><TeamSelector teams={teams} selectedTeam={selectedTeam} onTeamChange={() => {}} onCreateTeam={() => {}} /></SidebarMenuItem>
          <SidebarMenuItem><ProfileSelector profile={profile} onSettings={() => router.push('/settings')} onLogout={async () => { await clearAuthSession(); router.push('/login') }} /></SidebarMenuItem>
          <SidebarMenuItem>
            {resolvedTheme === 'light' ? (
              <RetroButton asChild variant="outline" className="w-full justify-start"><Link href="/settings"><Settings className="w-4 h-4" /><Text variant="small">Settings</Text></Link></RetroButton>
            ) : (
              <SidebarMenuButton asChild><Link href="/settings"><Settings className="w-4 h-4" /><span>Settings</span></Link></SidebarMenuButton>
            )}
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  )
}

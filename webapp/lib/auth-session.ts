import type { AppRole } from './api'

const activeRoleKey = 'mcp_active_role'

const roleDashboard: Record<AppRole, string> = {
  buyer: '/buyer/dashboard',
  merchant: '/merchant/onboarding',
  admin: '/admin/tenants',
}

export function getDashboardPath(role: AppRole): string {
  return roleDashboard[role]
}

export function setAuthSession(role: AppRole, token: string) {
  if (typeof window === 'undefined') return
  const oneWeek = 60 * 60 * 24 * 7
  document.cookie = `mcp_active_role=${role}; Path=/; Max-Age=${oneWeek}; SameSite=Lax`
  document.cookie = `mcp_access_token=${token}; Path=/; Max-Age=${oneWeek}; SameSite=Lax`
  localStorage.setItem(activeRoleKey, role)
}

export function clearAuthSession() {
  if (typeof window === 'undefined') return
  document.cookie = 'mcp_active_role=; Path=/; Max-Age=0; SameSite=Lax'
  document.cookie = 'mcp_access_token=; Path=/; Max-Age=0; SameSite=Lax'
  localStorage.removeItem(activeRoleKey)
  localStorage.removeItem('mcp_demo_tokens')
}


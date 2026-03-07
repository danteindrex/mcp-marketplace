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

export function setAuthSession(role: AppRole) {
  if (typeof window === 'undefined') return
  const oneWeek = 60 * 60 * 24 * 7
  document.cookie = `mcp_active_role=${role}; Path=/; Max-Age=${oneWeek}; SameSite=Lax`
  localStorage.setItem(activeRoleKey, role)
}

export async function clearAuthSession() {
  if (typeof window === 'undefined') return
  try {
    await fetch('/api/session/logout', {
      method: 'POST',
      credentials: 'include',
    })
  } catch {
    // Best-effort session cleanup on client logout.
  }
  document.cookie = 'mcp_active_role=; Path=/; Max-Age=0; SameSite=Lax'
  localStorage.removeItem(activeRoleKey)
}

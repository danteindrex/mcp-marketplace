import { getPublicApiBase } from './api-base'

export type AppRole = 'buyer' | 'merchant' | 'admin'

const API_BASE = getPublicApiBase()
const roleKey = 'mcp_active_role'

function readCookie(name: string): string | null {
  if (typeof window === 'undefined') return null
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`))
  return match ? decodeURIComponent(match[1]) : null
}

export function getActiveRole(): AppRole | null {
  const fromCookie = readCookie('mcp_active_role')
  if (fromCookie === 'buyer' || fromCookie === 'merchant' || fromCookie === 'admin') {
    return fromCookie
  }
  if (typeof window !== 'undefined') {
    const fromLocal = localStorage.getItem(roleKey)
    if (fromLocal === 'buyer' || fromLocal === 'merchant' || fromLocal === 'admin') {
      return fromLocal
    }
  }
  return null
}

export function setActiveRole(role: AppRole): void {
  if (typeof window === 'undefined') return
  const oneWeek = 60 * 60 * 24 * 7
  document.cookie = `mcp_active_role=${role}; Path=/; Max-Age=${oneWeek}; SameSite=Lax`
  localStorage.setItem(roleKey, role)
}

type AuthResponse = {
  user: {
    id: string
    tenantId: string
    email: string
    name: string
    role: AppRole
  }
}

export async function loginWithCredentials(
  email: string,
  password: string,
  mfaCode?: string,
): Promise<AuthResponse> {
  const res = await fetch('/api/session/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, mfaCode }),
    credentials: 'include',
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({} as { error?: string }))
    throw new Error(body.error || 'Invalid email/password')
  }
  return res.json()
}

export async function signupWithCredentials(payload: {
  email: string
  password: string
  name: string
  role: 'buyer' | 'merchant'
  tenantName: string
}): Promise<AuthResponse> {
  const res = await fetch('/api/session/signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    credentials: 'include',
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({} as { error?: string }))
    throw new Error(body.error || 'Signup failed')
  }
  return res.json()
}

export async function completeOAuthSignup(payload: {
  signupToken: string
  name: string
  role: 'buyer' | 'merchant'
  tenantName: string
}): Promise<AuthResponse> {
  const res = await fetch('/api/session/oauth/complete-signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    credentials: 'include',
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({} as { error?: string }))
    throw new Error(body.error || 'OAuth signup completion failed')
  }
  return res.json()
}

type OAuthStartResponse = {
 authorization_url?: string
 error?: string
}

export async function startOAuthFlow(
  provider: 'google' | 'github',
  options?: {
    mode?: 'login' | 'signup'
    role?: 'buyer' | 'merchant'
    name?: string
    tenantName?: string
  },
): Promise<string> {
  const params = new URLSearchParams()
  if (options?.mode) params.set('mode', options.mode)
  if (options?.role) params.set('role', options.role)
  if (options?.name) params.set('name', options.name)
  if (options?.tenantName) params.set('tenantName', options.tenantName)
  const suffix = params.toString() ? `?${params.toString()}` : ''
  const res = await fetch(`/api/auth/${provider}/start${suffix}`, {
    method: 'GET',
    credentials: 'include',
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({} as { error?: string }))
    throw new Error(body.error || `Failed to start ${provider} OAuth`)
  }
  const data = await res.json() as OAuthStartResponse
  if (!data.authorization_url) {
    throw new Error('No authorization URL returned')
  }
  return data.authorization_url
}

async function parseError(res: Response, fallback: string): Promise<Error> {
  const body = await res.json().catch(() => ({} as { error?: string }))
  return new Error(body.error || fallback)
}

async function callApi<T>(method: string, path: string, body?: unknown): Promise<T> {
  const url = `${API_BASE}${path}`
  const options: RequestInit = {
    method,
    credentials: 'include',
  }
  if (body) {
    options.headers = { 'Content-Type': 'application/json' }
    options.body = JSON.stringify(body)
  }

  let res = await fetch(url, options)

  if (res.status === 401 && !path.startsWith('/auth/refresh') && !path.startsWith('/auth/login') && !path.startsWith('/auth/signup')) {
    // Attempt to refresh the session
    const refreshRes = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })

    if (refreshRes.ok) {
      // Retry the original request
      res = await fetch(url, options)
    } else {
      // Refresh failed, possibly clear session/redirect to login
      if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/login')) {
        window.location.href = '/login?expired=1'
      }
    }
  }

  if (!res.ok) {
    throw await parseError(res, `API ${method} failed ${path} (${res.status})`)
  }

  return res.json()
}

export async function apiGet<T>(path: string, _role?: AppRole): Promise<T> {
  return callApi<T>('GET', path)
}

export async function apiPost<T>(path: string, body: unknown, _role?: AppRole): Promise<T> {
  return callApi<T>('POST', path, body)
}

export async function apiPut<T>(path: string, body: unknown, _role?: AppRole): Promise<T> {
  return callApi<T>('PUT', path, body)
}

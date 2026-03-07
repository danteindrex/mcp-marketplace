export type AppRole = 'buyer' | 'merchant' | 'admin'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'
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

async function parseError(res: Response, fallback: string): Promise<Error> {
  const body = await res.json().catch(() => ({} as { error?: string }))
  return new Error(body.error || fallback)
}

export async function apiGet<T>(path: string, _role?: AppRole): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
  })
  if (!res.ok) {
    throw await parseError(res, `API GET failed ${path} (${res.status})`)
  }
  return res.json()
}

export async function apiPost<T>(path: string, body: unknown, _role?: AppRole): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
    credentials: 'include',
  })
  if (!res.ok) {
    throw await parseError(res, `API POST failed ${path} (${res.status})`)
  }
  return res.json()
}

export async function apiPut<T>(path: string, body: unknown, _role?: AppRole): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'PUT',
    headers,
    body: JSON.stringify(body),
    credentials: 'include',
  })
  if (!res.ok) {
    throw await parseError(res, `API PUT failed ${path} (${res.status})`)
  }
  return res.json()
}

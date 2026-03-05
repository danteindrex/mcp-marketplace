export type AppRole = 'buyer' | 'merchant' | 'admin'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'
const tokenKey = 'mcp_demo_tokens'

function readTokenCache(): Record<string, string> {
  if (typeof window === 'undefined') return {}
  try {
    return JSON.parse(localStorage.getItem(tokenKey) || '{}')
  } catch {
    return {}
  }
}

function writeTokenCache(cache: Record<string, string>) {
  if (typeof window === 'undefined') return
  localStorage.setItem(tokenKey, JSON.stringify(cache))
}

function readCookie(name: string): string | null {
  if (typeof window === 'undefined') return null
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`))
  return match ? decodeURIComponent(match[1]) : null
}

type AuthResponse = {
  accessToken: string
  user: {
    id: string
    tenantId: string
    email: string
    name: string
    role: AppRole
  }
}

export async function loginWithCredentials(email: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!res.ok) throw new Error('Invalid email/password')
  return res.json()
}

export async function signupWithCredentials(payload: {
  email: string
  password: string
  name: string
  role: 'buyer' | 'merchant'
  tenantName: string
}): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/signup`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  if (!res.ok) throw new Error('Signup failed')
  return res.json()
}

export async function cacheRoleToken(role: AppRole, token: string): Promise<void> {
  const cache = readTokenCache()
  cache[role] = token
  writeTokenCache(cache)
}

export async function getRoleToken(role: AppRole): Promise<string> {
  const cache = readTokenCache()
  if (cache[role]) return cache[role]

  const cookieRole = readCookie('mcp_active_role')
  const cookieToken = readCookie('mcp_access_token')
  if (cookieRole === role && cookieToken) {
    cache[role] = cookieToken
    writeTokenCache(cache)
    return cookieToken
  }

  throw new Error('Not authenticated')
}

export async function apiGet<T>(path: string, role?: AppRole): Promise<T> {
  const headers: Record<string, string> = {}
  if (role) {
    const token = await getRoleToken(role)
    headers.Authorization = `Bearer ${token}`
  }
  const res = await fetch(`${API_BASE}${path}`, { headers })
  if (!res.ok) {
    throw new Error(`API GET failed ${path} (${res.status})`)
  }
  return res.json()
}

export async function apiPost<T>(path: string, body: unknown, role?: AppRole): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (role) {
    const token = await getRoleToken(role)
    headers.Authorization = `Bearer ${token}`
  }
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`API POST failed ${path} (${res.status})`)
  }
  return res.json()
}

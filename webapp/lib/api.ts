export type AppRole = 'buyer' | 'merchant' | 'admin'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'
const tokenKey = 'mcp_demo_tokens'

const roleEmailMap: Record<AppRole, string> = {
  buyer: 'buyer@acme.local',
  merchant: 'merchant@dataflow.local',
  admin: 'admin@platform.local',
}

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

export async function getRoleToken(role: AppRole): Promise<string> {
  const cache = readTokenCache()
  if (cache[role]) return cache[role]

  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: roleEmailMap[role] }),
  })
  if (!res.ok) throw new Error('Failed to authenticate demo role')
  const data = await res.json()
  cache[role] = data.accessToken
  writeTokenCache(cache)
  return data.accessToken
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
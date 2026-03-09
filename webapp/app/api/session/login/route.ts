import { NextResponse } from 'next/server'

const API_BASE = process.env.API_BASE_URL || process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:2024'
const SESSION_MAX_AGE_SECONDS = 60 * 60 * 8

type BackendAuthPayload = {
  accessToken: string
  user: {
    id: string
    tenantId: string
    email: string
    name: string
    role: 'buyer' | 'merchant' | 'admin'
  }
  oauth?: unknown
}

function withSessionCookies(response: NextResponse, payload: BackendAuthPayload): NextResponse {
  const secure = process.env.NODE_ENV === 'production'
  response.cookies.set('mcp_access_token', payload.accessToken, {
    httpOnly: true,
    sameSite: 'lax',
    secure,
    path: '/',
    maxAge: SESSION_MAX_AGE_SECONDS,
  })
  response.cookies.set('mcp_active_role', payload.user.role, {
    httpOnly: false,
    sameSite: 'lax',
    secure,
    path: '/',
    maxAge: SESSION_MAX_AGE_SECONDS,
  })
  return response
}

export async function POST(request: Request) {
  let body: unknown
  try {
    body = await request.json()
  } catch {
    return NextResponse.json({ error: 'invalid request body' }, { status: 400 })
  }

  const upstream = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    cache: 'no-store',
  })
  const payload = await upstream.json().catch(() => ({} as Record<string, unknown>))
  if (!upstream.ok) {
    return NextResponse.json(payload, { status: upstream.status })
  }

  const authPayload = payload as BackendAuthPayload
  if (!authPayload.accessToken || !authPayload.user?.role) {
    return NextResponse.json({ error: 'invalid auth response' }, { status: 502 })
  }

  const response = NextResponse.json(
    { user: authPayload.user, oauth: authPayload.oauth || null },
    { status: 200 },
  )
  return withSessionCookies(response, authPayload)
}


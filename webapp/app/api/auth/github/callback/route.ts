import { NextResponse } from 'next/server'
import { getServerApiBase } from '@/lib/api-base'

const API_BASE = getServerApiBase()
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
  redirect_url?: string
}

type BackendPendingOAuthSignup = {
  signupRequired?: boolean
  signupToken?: string
  profile?: {
    email?: string
    name?: string
    avatarUrl?: string
  }
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

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const code = searchParams.get('code')
  const state = searchParams.get('state')
  
  if (!code || !state) {
    return NextResponse.redirect(new URL('/login?error=missing_params', request.url))
  }

  try {
    const upstream = await fetch(
      `${API_BASE}/auth/oauth/github/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`,
      {
        method: 'GET',
        credentials: 'include',
        cache: 'no-store',
      }
    )

    if (!upstream.ok) {
      const payload = await upstream.json().catch(() => ({} as Record<string, unknown>))
      const errorMessage = (payload as { error?: string }).error || 'OAuth callback failed'
      return NextResponse.redirect(new URL(`/login?error=${encodeURIComponent(errorMessage)}`, request.url))
    }

    const authPayload = await upstream.json()
    const pending = authPayload as BackendPendingOAuthSignup
    if (pending.signupRequired && pending.signupToken) {
      const target = new URL('/login', request.url)
      target.searchParams.set('oauth', 'complete')
      target.searchParams.set('signupToken', pending.signupToken)
      if (pending.profile?.email) target.searchParams.set('email', pending.profile.email)
      if (pending.profile?.name) target.searchParams.set('name', pending.profile.name)
      return NextResponse.redirect(target)
    }

    const payload = authPayload as BackendAuthPayload

    if (!payload.accessToken || !payload.user?.role) {
      return NextResponse.redirect(new URL('/login?error=invalid_response', request.url))
    }

    const response = NextResponse.redirect(new URL('/login?oauth=success', request.url))
    return withSessionCookies(response, payload)
  } catch {
    return NextResponse.redirect(new URL('/login?error=callback_failed', request.url))
  }
}

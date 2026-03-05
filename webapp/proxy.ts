import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const protectedPrefixes = ['/buyer', '/merchant', '/admin']
const apiBase = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'

function isPublicPath(pathname: string): boolean {
  if (pathname === '/' || pathname === '/login') return true
  if (pathname === '/marketplace' || pathname.startsWith('/marketplace/')) return true
  if (pathname.startsWith('/_next') || pathname.startsWith('/favicon')) return true
  return false
}

function dashboardForRole(role: string): string {
  if (role === 'merchant') return '/merchant/onboarding'
  if (role === 'admin') return '/admin/tenants'
  return '/buyer/dashboard'
}

function roleAllowedForPath(pathname: string, role: string): boolean {
  if (pathname.startsWith('/buyer')) return role === 'buyer'
  if (pathname.startsWith('/merchant')) return role === 'merchant' || role === 'admin'
  if (pathname.startsWith('/admin')) return role === 'admin'
  return true
}

function loginRedirect(request: NextRequest, pathname: string): NextResponse {
  const loginUrl = new URL('/login', request.url)
  loginUrl.searchParams.set('next', pathname)
  return NextResponse.redirect(loginUrl)
}

function clearSessionCookies(response: NextResponse): NextResponse {
  response.cookies.delete('mcp_active_role')
  response.cookies.delete('mcp_access_token')
  return response
}

async function resolveTokenRole(token: string): Promise<string | null> {
  try {
    const res = await fetch(`${apiBase}/v1/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    })
    if (!res.ok) return null
    const user = await res.json()
    return user?.role ? String(user.role) : null
  } catch {
    return null
  }
}

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl
  if (isPublicPath(pathname)) {
    if (pathname === '/login') {
      const role = request.cookies.get('mcp_active_role')?.value
      const token = request.cookies.get('mcp_access_token')?.value
      if (!role || !token) return NextResponse.next()

      const resolvedRole = await resolveTokenRole(token)
      if (!resolvedRole) {
        return clearSessionCookies(NextResponse.next())
      }
      if (roleAllowedForPath(dashboardForRole(resolvedRole), resolvedRole)) {
        const response = NextResponse.redirect(new URL(dashboardForRole(resolvedRole), request.url))
        if (resolvedRole !== role) {
          response.cookies.set('mcp_active_role', resolvedRole, { path: '/', sameSite: 'lax' })
        }
        return response
      }
    }
    return NextResponse.next()
  }

  const isProtected = protectedPrefixes.some(prefix => pathname.startsWith(prefix))
  if (!isProtected) return NextResponse.next()

  const role = request.cookies.get('mcp_active_role')?.value
  const token = request.cookies.get('mcp_access_token')?.value
  if (!role || !token || !roleAllowedForPath(pathname, role)) {
    return clearSessionCookies(loginRedirect(request, pathname))
  }

  const resolvedRole = await resolveTokenRole(token)
  if (!resolvedRole || !roleAllowedForPath(pathname, resolvedRole)) {
    return clearSessionCookies(loginRedirect(request, pathname))
  }

  const response = NextResponse.next()
  if (resolvedRole !== role) {
    response.cookies.set('mcp_active_role', resolvedRole, { path: '/', sameSite: 'lax' })
  }
  return response
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|.*\\.png$).*)'],
}

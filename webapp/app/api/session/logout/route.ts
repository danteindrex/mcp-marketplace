import { NextResponse } from 'next/server'

export async function POST() {
  const response = NextResponse.json({ status: 'ok' }, { status: 200 })
  response.cookies.set('mcp_access_token', '', {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge: 0,
  })
  response.cookies.set('mcp_active_role', '', {
    httpOnly: false,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge: 0,
  })
  return response
}


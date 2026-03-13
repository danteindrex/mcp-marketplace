import { NextResponse } from 'next/server'
import { getServerApiBase } from '@/lib/api-base'

const API_BASE = getServerApiBase()
const SESSION_MAX_AGE_SECONDS = 60 * 60 * 8

export async function GET(request: Request) {
  try {
    const url = new URL(request.url)
    const qs = url.searchParams.toString()
    const upstream = await fetch(`${API_BASE}/auth/oauth/google/start${qs ? `?${qs}` : ''}`, {
      method: 'GET',
      credentials: 'include',
      cache: 'no-store',
    })
    
    if (!upstream.ok) {
      const payload = await upstream.json().catch(() => ({} as Record<string, unknown>))
      return NextResponse.json(payload, { status: upstream.status })
    }
    
    const data = await upstream.json()
    return NextResponse.json(data, { status: 200 })
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to initiate Google OAuth' },
      { status: 500 }
    )
  }
}

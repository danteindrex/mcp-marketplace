import { NextResponse } from 'next/server'
import { getServerApiBase } from '@/lib/api-base'

const API_BASE = getServerApiBase()

export async function GET() {
  try {
    const upstream = await fetch(`${API_BASE}/auth/oauth/github/start`, {
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
      { error: 'Failed to initiate GitHub OAuth' },
      { status: 500 }
    )
  }
}

import { NextResponse } from 'next/server'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'

export async function GET() {
  try {
    const upstream = await fetch(`${API_BASE}/auth/github/start`, {
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

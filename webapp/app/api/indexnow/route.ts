import { NextRequest, NextResponse } from 'next/server'
import { submitIndexNow } from '@/lib/indexnow'

export async function POST(request: NextRequest) {
  const authHeader = request.headers.get('authorization')
  const bearerToken = authHeader?.startsWith('Bearer ') ? authHeader.slice(7).trim() : ''
  const configuredToken = process.env.INDEXNOW_SUBMIT_TOKEN?.trim()

  if (!configuredToken || bearerToken !== configuredToken) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const body = (await request.json().catch(() => null)) as { urls?: string[] } | null
  const urls = Array.isArray(body?.urls)
    ? body.urls.filter((url): url is string => typeof url === 'string' && url.startsWith('http'))
    : []

  if (urls.length === 0) {
    return NextResponse.json({ error: 'No valid URLs supplied' }, { status: 400 })
  }

  const result = await submitIndexNow(urls)
  return NextResponse.json(result, { status: result.ok ? 200 : result.status })
}

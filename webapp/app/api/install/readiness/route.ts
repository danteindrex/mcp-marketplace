import { NextRequest, NextResponse } from 'next/server'

type FetchStatus = 'ok' | 'error'

interface FetchResult<T = any> {
  status: FetchStatus
  httpStatus: number
  url: string
  json?: T | null
  error?: string
}

const DEFAULT_TIMEOUT_MS = 7000

async function safeFetchJson(url: string): Promise<FetchResult> {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT_MS)
  try {
    const res = await fetch(url, {
      cache: 'no-store',
      headers: { Accept: 'application/json' },
      signal: controller.signal,
    })
    const json = await res.json().catch(() => null)
    return {
      status: res.ok ? 'ok' : 'error',
      httpStatus: res.status,
      url,
      json,
      error: res.ok ? undefined : `HTTP ${res.status}`,
    }
  } catch (error) {
    return {
      status: 'error',
      httpStatus: 0,
      url,
      error: error instanceof Error ? error.message : String(error),
    }
  } finally {
    clearTimeout(timeout)
  }
}

function sanitizeBaseUrl(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) return trimmed
  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

function buildUrl(base: string, path: string): string {
  try {
    const normalized = base.endsWith('/') ? base : `${base}/`
    return new URL(path.replace(/^\//, ''), normalized).toString()
  } catch {
    return `${base}${path}`
  }
}

export async function GET(req: NextRequest) {
  const resource = req.nextUrl.searchParams.get('resource')
  if (!resource) {
    return NextResponse.json({ error: 'resource query parameter required' }, { status: 400 })
  }
  let parsed: URL
  try {
    parsed = new URL(resource)
  } catch {
    return NextResponse.json({ error: 'resource must be an absolute URL' }, { status: 400 })
  }

  const baseUrl = sanitizeBaseUrl(parsed.toString())
  const cimdUrl = buildUrl(baseUrl, '/.well-known/mcp.json')
  const cimdResult = await safeFetchJson(cimdUrl)

  const authorizationServer = (cimdResult.json as any)?.authorization_server as string | undefined
  const oauthMetadataUrl = authorizationServer || buildUrl(baseUrl, '/.well-known/oauth-authorization-server')
  const oauthResult = await safeFetchJson(oauthMetadataUrl)

  const jwksUrl = (oauthResult.json as any)?.jwks_uri as string | undefined
  const jwksResult = jwksUrl ? await safeFetchJson(jwksUrl) : null

  return NextResponse.json({
    baseUrl,
    cimd: cimdResult,
    oauth: oauthResult,
    jwks: jwksResult,
    links: {
      cimdUrl,
      oauthMetadataUrl,
      jwksUrl: jwksUrl || null,
    },
    timestamp: new Date().toISOString(),
  })
}

import { NextRequest, NextResponse } from 'next/server'
import { getPublicApiBase, getServerApiBase } from '@/lib/api-base'

const API_BASE = getServerApiBase()
const PUBLIC_API_BASE = getPublicApiBase()

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
  const requestedBase = req.nextUrl.searchParams.get('metadataBaseUrl')
  const requestedResource = req.nextUrl.searchParams.get('resource')
  const displayCandidate = requestedBase || requestedResource || PUBLIC_API_BASE
  const fetchCandidate = requestedBase ? API_BASE : requestedResource || API_BASE
  let displayParsed: URL
  let fetchParsed: URL
  try {
    displayParsed = new URL(displayCandidate)
    fetchParsed = new URL(fetchCandidate)
  } catch {
    return NextResponse.json({ error: 'metadataBaseUrl must be an absolute URL' }, { status: 400 })
  }

  const baseUrl = sanitizeBaseUrl(displayParsed.toString())
  const fetchBaseUrl = sanitizeBaseUrl(fetchParsed.toString())
  let cimdUrl = buildUrl(baseUrl, '/.well-known/mcp.json')
  let cimdFetchUrl = buildUrl(fetchBaseUrl, '/.well-known/mcp.json')
  if (requestedResource) {
    try {
      const resource = new URL(requestedResource).toString()
      const params = new URLSearchParams({ resource })
      cimdUrl = `${cimdUrl}?${params.toString()}`
      cimdFetchUrl = `${cimdFetchUrl}?${params.toString()}`
    } catch {
      return NextResponse.json({ error: 'resource must be an absolute URL' }, { status: 400 })
    }
  }
  const cimdResult = await safeFetchJson(cimdFetchUrl)

  const authorizationServer = (cimdResult.json as any)?.authorization_server as string | undefined
  const oauthMetadataUrl = authorizationServer || buildUrl(baseUrl, '/.well-known/oauth-authorization-server')
  const oauthFetchUrl = requestedBase ? buildUrl(fetchBaseUrl, '/.well-known/oauth-authorization-server') : oauthMetadataUrl
  const oauthResult = await safeFetchJson(oauthFetchUrl)

  const jwksUrl = (oauthResult.json as any)?.jwks_uri as string | undefined
  const jwksFetchUrl = requestedBase && jwksUrl ? buildUrl(fetchBaseUrl, '/.well-known/jwks.json') : jwksUrl
  const jwksResult = jwksFetchUrl ? await safeFetchJson(jwksFetchUrl) : null

  return NextResponse.json({
    baseUrl,
    resource: requestedResource || null,
    cimd: cimdResult,
    oauth: oauthResult,
    jwks: jwksResult,
    links: {
      cimdUrl,
      oauthMetadataUrl,
      jwksUrl: jwksUrl || (requestedBase ? buildUrl(baseUrl, '/.well-known/jwks.json') : null),
    },
    timestamp: new Date().toISOString(),
  })
}

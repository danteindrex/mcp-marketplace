const DEFAULT_SITE_URL = 'http://localhost:3000'

function normalizeUrl(value: string | undefined): string {
  const trimmed = value?.trim()
  if (!trimmed) {
    return DEFAULT_SITE_URL
  }
  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

export function getSiteUrl(): string {
  return normalizeUrl(
    process.env.SITE_URL ||
      process.env.NEXT_PUBLIC_SITE_URL ||
      process.env.NEXT_PUBLIC_APP_URL,
  )
}

export function getSiteUrlObject(): URL {
  return new URL(getSiteUrl())
}

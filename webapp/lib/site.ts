function normalizeUrl(value: string | undefined): string {
  const trimmed = value?.trim()
  if (!trimmed) {
    throw new Error('SITE_URL, NEXT_PUBLIC_SITE_URL, or NEXT_PUBLIC_APP_URL must be set')
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

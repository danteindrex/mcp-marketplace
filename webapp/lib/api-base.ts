function normalizeBaseUrl(value: string | undefined): string {
  const trimmed = value?.trim()
  if (!trimmed) {
    return ''
  }
  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

export function getPublicApiBase(): string {
  const value = normalizeBaseUrl(process.env.NEXT_PUBLIC_API_BASE_URL)
  if (!value) {
    throw new Error('NEXT_PUBLIC_API_BASE_URL is not configured')
  }
  return value
}

export function getServerApiBase(): string {
  const value =
    normalizeBaseUrl(process.env.API_BASE_URL) || normalizeBaseUrl(process.env.NEXT_PUBLIC_API_BASE_URL)
  if (!value) {
    throw new Error('API_BASE_URL or NEXT_PUBLIC_API_BASE_URL must be configured')
  }
  return value
}

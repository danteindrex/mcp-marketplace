function normalizeBaseUrl(value: string | undefined): string {
  const trimmed = value?.trim()
  if (!trimmed) {
    return ''
  }
  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

const LOCAL_DEFAULT_API_BASE = 'http://localhost:8080'

export function getPublicApiBase(): string {
  const value = normalizeBaseUrl(process.env.NEXT_PUBLIC_API_BASE_URL)
  return value || LOCAL_DEFAULT_API_BASE
}

export function getServerApiBase(): string {
  const value =
    normalizeBaseUrl(process.env.API_BASE_URL) || normalizeBaseUrl(process.env.NEXT_PUBLIC_API_BASE_URL)
  return value || LOCAL_DEFAULT_API_BASE
}

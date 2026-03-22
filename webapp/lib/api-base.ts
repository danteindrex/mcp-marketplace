function normalizeBaseUrl(value: string | undefined): string {
  const trimmed = value?.trim()
  if (!trimmed) {
    return ''
  }
  return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed
}

function requireConfiguredBaseUrl(value: string, envNames: string): string {
  if (!value) {
    throw new Error(`${envNames} must be set`)
  }
  return value
}

export function getPublicApiBase(): string {
  const value = normalizeBaseUrl(process.env.NEXT_PUBLIC_API_BASE_URL)
  return requireConfiguredBaseUrl(value, 'NEXT_PUBLIC_API_BASE_URL')
}

export function getServerApiBase(): string {
  const value =
    normalizeBaseUrl(process.env.API_BASE_URL) || normalizeBaseUrl(process.env.NEXT_PUBLIC_API_BASE_URL)
  return requireConfiguredBaseUrl(value, 'API_BASE_URL or NEXT_PUBLIC_API_BASE_URL')
}

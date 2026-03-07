'use client'

import { ServerError } from '@/components/blocks/application/error-pages'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <ServerError
      onRetry={reset}
      homeHref="/"
      errorId={error.digest ?? 'ERR-UNKNOWN'}
      supportEmail="support@mcp-marketplace.local"
    />
  )
}

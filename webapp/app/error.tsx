'use client'

import { ServerError } from '@/components/blocks/application/error-pages'
import { SUPPORT_EMAIL } from '@/lib/contact'

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
      supportEmail={SUPPORT_EMAIL}
    />
  )
}

'use client'

import { Generic } from '@/components/blocks/application/error-pages'

export default function GlobalError({
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <html lang="en">
      <body>
        <Generic
          title="Unexpected Application Error"
          description="A global rendering error occurred. Please retry or return home."
          actions={[
            { label: 'Try Again', onClick: reset },
            { label: 'Go Home', href: '/', variant: 'outline' },
          ]}
        />
      </body>
    </html>
  )
}

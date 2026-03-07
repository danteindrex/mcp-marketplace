'use client'

import { Generic } from '@/components/blocks/application/error-pages'

export default function GenericErrorPage() {
  return (
    <Generic
      title="Oops!"
      description="Something went wrong."
      actions={[
        { label: 'Try Again', onClick: () => window.location.reload() },
        { label: 'Go Home', href: '/', variant: 'outline' },
      ]}
    />
  )
}

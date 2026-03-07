'use client'

import { Offline } from '@/components/blocks/application/error-pages'

export default function OfflinePage() {
  return <Offline onRetry={() => window.location.reload()} />
}

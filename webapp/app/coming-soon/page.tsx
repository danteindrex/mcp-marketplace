'use client'

import { ComingSoon } from '@/components/blocks/application/error-pages'

export default function ComingSoonPage() {
  return (
    <ComingSoon
      launchDate={new Date('2026-12-01')}
      onNotify={(email) => {
        console.info('Notify request:', email)
      }}
    />
  )
}

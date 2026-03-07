import type { ReactElement } from 'react'

interface Breadcrumb {
  name: string
  url?: string
}

interface SEOProps {
  title: string
  description?: string
  keywords?: string
  canonical?: string
  breadcrumbs?: Breadcrumb[]
}

export function SEO(_: SEOProps): ReactElement | null {
  return null
}

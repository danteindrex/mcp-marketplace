import { getSiteUrl } from './site'

interface BreadcrumbItem {
  name: string
  path: string
}

export function toAbsoluteUrl(path: string): string {
  const siteUrl = getSiteUrl()
  if (!path.startsWith('/')) {
    return `${siteUrl}/${path}`
  }
  return `${siteUrl}${path}`
}

export function createBreadcrumbJsonLd(items: BreadcrumbItem[]) {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      item: toAbsoluteUrl(item.path),
    })),
  }
}

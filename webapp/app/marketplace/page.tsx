import { JsonLd } from '@/components/json-ld'
import { fetchPublicServers } from '@/lib/public-marketplace'
import { createBreadcrumbJsonLd, toAbsoluteUrl } from '@/lib/seo'
import { MarketplaceClientPage } from './client-page'

export default async function MarketplacePage() {
  const servers = await fetchPublicServers()
  const jsonLd = [
    createBreadcrumbJsonLd([
      { name: 'Home', path: '/' },
      { name: 'Marketplace', path: '/marketplace' },
    ]),
    {
      '@context': 'https://schema.org',
      '@type': 'CollectionPage',
      name: 'MCP Marketplace',
      description: 'Browse installable MCP servers across supported clients and pricing models.',
      url: toAbsoluteUrl('/marketplace'),
      mainEntity: {
        '@type': 'ItemList',
        itemListElement: servers.slice(0, 12).map((server, index) => ({
          '@type': 'ListItem',
          position: index + 1,
          url: toAbsoluteUrl(`/marketplace/${server.slug}`),
          name: server.name,
        })),
      },
    },
  ]

  return (
    <>
      <JsonLd data={jsonLd} />
      <MarketplaceClientPage initialServers={servers} />
    </>
  )
}

import { JsonLd } from '@/components/json-ld'
import { fetchPublicServers } from '@/lib/public-marketplace'
import { getSiteUrl } from '@/lib/site'
import { MarketplaceClientPage } from './client-page'

export default async function MarketplacePage() {
  const servers = await fetchPublicServers()
  const siteUrl = getSiteUrl()
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'ItemList',
    itemListElement: servers.slice(0, 12).map((server, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      url: `${siteUrl}/marketplace/${server.slug}`,
      name: server.name,
    })),
  }

  return (
    <>
      <JsonLd data={jsonLd} />
      <MarketplaceClientPage initialServers={servers} />
    </>
  )
}

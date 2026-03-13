import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { JsonLd } from '@/components/json-ld'
import { fetchPublicServerBySlug } from '@/lib/public-marketplace'
import { getSiteUrl } from '@/lib/site'
import { ServerDetailClientPage } from './client-page'

interface PageProps {
  params: Promise<{ serverId: string }>
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const { serverId } = await params
  const server = await fetchPublicServerBySlug(serverId)

  if (!server) {
    return {
      title: 'Server Not Found',
      robots: {
        index: false,
        follow: false,
      },
    }
  }

  return {
    title: server.name,
    description: server.description,
    alternates: {
      canonical: `/marketplace/${server.slug}`,
    },
    openGraph: {
      title: server.name,
      description: server.description,
      url: `/marketplace/${server.slug}`,
      type: 'website',
    },
  }
}

export default async function ServerDetailPage({ params }: PageProps) {
  const { serverId } = await params
  const server = await fetchPublicServerBySlug(serverId)

  if (!server) {
    notFound()
  }

  const siteUrl = getSiteUrl()
  const canonicalUrl = `${siteUrl}/marketplace/${server.slug}`
  const jsonLd = [
    {
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      itemListElement: [
        {
          '@type': 'ListItem',
          position: 1,
          name: 'Home',
          item: siteUrl,
        },
        {
          '@type': 'ListItem',
          position: 2,
          name: 'Marketplace',
          item: `${siteUrl}/marketplace`,
        },
        {
          '@type': 'ListItem',
          position: 3,
          name: server.name,
          item: canonicalUrl,
        },
      ],
    },
    {
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: server.name,
      description: server.description,
      applicationCategory: 'DeveloperApplication',
      operatingSystem: 'Web',
      offers: {
        '@type': 'Offer',
        price: Number(server.pricingAmount || 0),
        priceCurrency: 'USD',
        availability: 'https://schema.org/InStock',
      },
      aggregateRating: server.rating > 0
        ? {
            '@type': 'AggregateRating',
            ratingValue: server.rating,
            ratingCount: Math.max(server.installCount, 1),
          }
        : undefined,
      url: canonicalUrl,
    },
  ]

  return (
    <>
      <JsonLd data={jsonLd} />
      <ServerDetailClientPage serverId={serverId} initialServer={server} />
    </>
  )
}
